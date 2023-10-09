/*
 *
 * Copyright 2022 puzzlerightserver authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package rightserver

import (
	"context"
	"database/sql"
	"errors"
	"sync"

	"github.com/dvaumoron/puzzlerightserver/model"
	pb "github.com/dvaumoron/puzzlerightservice"
	_ "github.com/jackc/pgx/v5"
	"github.com/open-policy-agent/opa/rego"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
)

const RightKey = "puzzleRight"

const publicObjectId = 0

const dbAccessMsg = "Failed to access database"

var errInternal = errors.New("internal service error")

type empty = struct{}

// server is used to implement puzzlerightservice.RightServer.
type server struct {
	pb.UnimplementedRightServer
	db            *sql.DB
	rule          rego.PreparedEvalQuery
	idToNameMutex sync.RWMutex
	idToName      map[uint64]string
	logger        *otelzap.Logger
}

func New(db *sql.DB, opaRule rego.PreparedEvalQuery, logger *otelzap.Logger) pb.RightServer {
	return &server{db: db, rule: opaRule, idToName: map[uint64]string{}, logger: logger}
}

func (s *server) AuthQuery(ctx context.Context, request *pb.RightRequest) (*pb.Response, error) {
	logger := s.logger.Ctx(ctx)
	conn, err := s.db.Conn(ctx)
	if err != nil {
		logger.Error(dbAccessMsg, zap.Error(err))
		return nil, errInternal
	}
	defer conn.Close()

	var roles []model.Role
	userId := request.UserId
	if userId != 0 {
		if roles, err = model.GetRolesByUserId(conn, ctx, userId); err != nil {
			logger.Error(dbAccessMsg, zap.Error(err))
			return nil, errInternal
		}
	}

	input := map[string]any{
		"userId": userId, "objectId": request.ObjectId,
		"actionFlag": convertActionToFlag(request.Action),
		"userRoles":  convertDataFromRolesModel(roles),
	}
	results, err := s.rule.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		logger.Error("OPA evaluation failed", zap.Error(err))
		return nil, errInternal
	}
	return &pb.Response{Success: results.Allowed()}, nil
}

func (s *server) ListRoles(ctx context.Context, request *pb.ObjectIds) (*pb.Roles, error) {
	logger := s.logger.Ctx(ctx)
	conn, err := s.db.Conn(ctx)
	if err != nil {
		logger.Error(dbAccessMsg, zap.Error(err))
		return nil, errInternal
	}
	defer conn.Close()

	roles, err := model.GetRolesByObjectIds(conn, ctx, request.Ids)
	if err != nil {
		logger.Error(dbAccessMsg, zap.Error(err))
		return nil, errInternal
	}

	resRoles, err := s.convertRolesFromModel(conn, logger, roles)
	if err != nil {
		return nil, err
	}
	return &pb.Roles{List: resRoles}, nil
}

func (s *server) RoleRight(ctx context.Context, request *pb.RoleRequest) (*pb.Actions, error) {
	logger := s.logger.Ctx(ctx)
	conn, err := s.db.Conn(ctx)
	if err != nil {
		logger.Error(dbAccessMsg, zap.Error(err))
		return nil, errInternal
	}
	defer conn.Close()

	role, err := model.GetRoleByNameAndObjectId(conn, ctx, request.Name, request.ObjectId)
	if err != nil {
		if err == sql.ErrNoRows {
			// ignore unknown role
			return &pb.Actions{}, nil
		}

		logger.Error(dbAccessMsg, zap.Error(err))
		return nil, errInternal
	}

	actions := convertActionsFromFlags(role.ActionFlags)
	return &pb.Actions{List: actions}, nil
}

func (s *server) UpdateUser(ctx context.Context, request *pb.UserRight) (response *pb.Response, err error) {
	logger := s.logger.Ctx(ctx)
	conn, err := s.db.Conn(ctx)
	if err != nil {
		logger.Error(dbAccessMsg, zap.Error(err))
		return nil, errInternal
	}
	defer conn.Close()

	roles, err := loadRoles(conn, logger, request.List)
	if err != nil {
		return
	}

	userId := request.UserId
	rolesLen := len(roles)
	if rolesLen == 0 {
		// delete unused user
		if _, err = model.DeleteUserRolesByUserId(conn, ctx, userId); err != nil {
			logger.Error(dbAccessMsg, zap.Error(err))
			return nil, errInternal
		}
		return &pb.Response{Success: true}, nil
	}

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		logger.Error(dbAccessMsg, zap.Error(err))
		return
	}
	defer commitOrRollBack(tx, logger, &err)

	if _, err = model.DeleteUserRolesByUserId(tx, ctx, userId); err != nil {
		logger.Error(dbAccessMsg, zap.Error(err))
		return nil, errInternal
	}

	for _, role := range roles {
		if err = model.MakeUserRole(0, userId, role.Id).Create(tx, ctx); err != nil {
			logger.Error(dbAccessMsg, zap.Error(err))
			return nil, errInternal
		}
	}
	return &pb.Response{Success: true}, nil
}

func (s *server) UpdateRole(ctx context.Context, request *pb.Role) (response *pb.Response, err error) {
	logger := s.logger.Ctx(ctx)
	conn, err := s.db.Conn(ctx)
	if err != nil {
		logger.Error(dbAccessMsg, zap.Error(err))
		return nil, errInternal
	}
	defer conn.Close()

	name := request.Name
	objectId := request.ObjectId

	if objectId == publicObjectId {
		// right on public part are not updatable, return false (bool default)
		return &pb.Response{}, nil
	}

	actionFlags := convertActionsToFlags(request.List)
	if actionFlags == 0 {
		// delete unused role
		role, err := model.GetRoleByNameAndObjectId(conn, ctx, name, objectId)
		if err != sql.ErrNoRows {
			if err == nil {
				err = role.Delete(conn, ctx)
			}
			if err != nil {
				logger.Error(dbAccessMsg, zap.Error(err))
				return nil, errInternal
			}
		}

		// we delete the names without roles
		if _, err = model.DeleteUnusedRoleNames(conn, ctx); err != nil {
			logger.Error(dbAccessMsg, zap.Error(err))
			return nil, errInternal
		}

		// invalidate the cache of name
		s.idToNameMutex.Lock()
		s.idToName = map[uint64]string{}
		s.idToNameMutex.Unlock()
		return &pb.Response{Success: true}, nil
	}

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		logger.Error(dbAccessMsg, zap.Error(err))
		return nil, errInternal
	}
	defer commitOrRollBack(tx, logger, &err)

	roleName, err := model.GetRoleNameByName(conn, ctx, name)
	if err == sql.ErrNoRows {
		if err = model.MakeRoleName(0, name).Create(conn, ctx); err == nil {
			// must retrieve the id
			roleName, err = model.GetRoleNameByName(conn, ctx, name)
		}
	}
	if err != nil {
		logger.Error(dbAccessMsg, zap.Error(err))
		return nil, errInternal
	}

	role, err := model.GetRoleByNameIdAndObjectId(tx, ctx, roleName.Id, objectId)
	if err == nil {
		role.ActionFlags = actionFlags
		if err = role.Update(tx, ctx); err != nil {
			logger.Error(dbAccessMsg, zap.Error(err))
			return nil, errInternal
		}

		return &pb.Response{Success: true}, nil
	}
	if err != sql.ErrNoRows {
		logger.Error(dbAccessMsg, zap.Error(err))
		return nil, errInternal
	}

	role = model.MakeRole(0, roleName.Id, objectId, actionFlags)
	if err = role.Create(tx, ctx); err != nil {
		logger.Error(dbAccessMsg, zap.Error(err))
		return nil, errInternal
	}
	return &pb.Response{Success: true}, nil
}

func (s *server) ListUserRoles(ctx context.Context, request *pb.UserId) (*pb.Roles, error) {
	logger := s.logger.Ctx(ctx)
	conn, err := s.db.Conn(ctx)
	if err != nil {
		logger.Error(dbAccessMsg, zap.Error(err))
		return nil, errInternal
	}
	defer conn.Close()

	roles, err := model.GetRolesByUserId(conn, ctx, request.Id)
	if err != nil {
		logger.Error(dbAccessMsg, zap.Error(err))
		return nil, errInternal
	}

	resRoles, err := s.convertRolesFromModel(conn, logger, roles)
	if err != nil {
		logger.Error(dbAccessMsg, zap.Error(err))
		return nil, errInternal
	}
	return &pb.Roles{List: resRoles}, nil
}

func loadRoles(conn *sql.Conn, logger otelzap.LoggerWithCtx, roles []*pb.RoleRequest) ([]model.Role, error) {
	ctx := logger.Context()
	resRoles := make([]model.Role, 0, len(roles)) // probably lot more space than necessary
	for name, objectIds := range extractNamesToObjectIds(roles) {
		roles, err := model.GetRolesByNameAndObjectIds(conn, ctx, name, objectIds)
		if err != nil {
			logger.Error(dbAccessMsg, zap.Error(err))
			return nil, errInternal
		}
		if len(roles) != 0 {
			resRoles = append(resRoles, roles...)
		}
	}
	return resRoles, nil
}

func (s *server) convertRolesFromModel(conn *sql.Conn, logger otelzap.LoggerWithCtx, roles []model.Role) ([]*pb.Role, error) {
	allThere := true
	resRoles := make([]*pb.Role, 0, len(roles))
	s.idToNameMutex.RLock()
	for _, role := range roles {
		var name string
		id := role.NameId
		name, allThere = s.idToName[id]
		if !allThere {
			break
		}
		resRoles = append(resRoles, convertRoleFromModel(name, role))
	}
	s.idToNameMutex.RUnlock()
	if allThere {
		return resRoles, nil
	}

	s.idToNameMutex.Lock()
	defer s.idToNameMutex.Unlock()
	allThere = true
	resRoles = resRoles[:0]
	missingIdSet := map[uint64]empty{}
	for _, role := range roles {
		id := role.NameId
		name, ok := s.idToName[id]
		if ok {
			resRoles = append(resRoles, convertRoleFromModel(name, role))
		} else {
			allThere = false
			missingIdSet[id] = empty{}
		}
	}
	if allThere {
		return resRoles, nil
	}

	queryIds := make([]uint64, 0, len(missingIdSet))
	for id := range missingIdSet {
		queryIds = append(queryIds, id)
	}

	roleNames, err := model.GetRoleNamesByIds(conn, logger.Context(), queryIds)
	if err != nil {
		logger.Error(dbAccessMsg, zap.Error(err))
		return nil, errInternal
	}

	for _, roleName := range roleNames {
		s.idToName[roleName.Id] = roleName.Name
	}

	resRoles = resRoles[:0]
	for _, role := range roles {
		resRoles = append(resRoles, convertRoleFromModel(s.idToName[role.NameId], role))
	}
	return resRoles, nil
}

func convertRoleFromModel(name string, role model.Role) *pb.Role {
	return &pb.Role{
		Name: name, ObjectId: role.ObjectId,
		List: convertActionsFromFlags(role.ActionFlags),
	}
}

func convertDataFromRolesModel(roles []model.Role) []any {
	res := make([]any, 0, len(roles))
	for _, role := range roles {
		res = append(res, map[string]any{
			"objectId":    role.ObjectId,
			"actionFlags": role.ActionFlags,
		})
	}
	return res
}

func commitOrRollBack(tx *sql.Tx, logger otelzap.LoggerWithCtx, err *error) {
	if r := recover(); r != nil {
		tx.Rollback()
		logger.Error(dbAccessMsg, zap.Any("recover", r))
	} else if *err == nil {
		tx.Commit()
	} else {
		tx.Rollback()
	}
}

func convertActionsFromFlags(actionFlags uint8) []pb.RightAction {
	resActions := make([]pb.RightAction, 0, 4)
	if actionFlags&1 != 0 {
		resActions = append(resActions, pb.RightAction_ACCESS)
	}
	if actionFlags&2 != 0 {
		resActions = append(resActions, pb.RightAction_CREATE)
	}
	if actionFlags&4 != 0 {
		resActions = append(resActions, pb.RightAction_UPDATE)
	}
	if actionFlags&8 != 0 {
		resActions = append(resActions, pb.RightAction_DELETE)
	}
	return resActions
}

func extractNamesToObjectIds(roles []*pb.RoleRequest) map[string][]uint64 {
	nameToObjectIdSet := map[string]map[uint64]empty{}
	for _, role := range roles {
		name := role.Name
		objectIdSet := nameToObjectIdSet[name]
		if objectIdSet == nil {
			objectIdSet = map[uint64]empty{}
			nameToObjectIdSet[name] = objectIdSet
		}
		objectIdSet[role.ObjectId] = empty{}
	}
	nameToObjectIds := map[string][]uint64{}
	for name, objectIdSet := range nameToObjectIdSet {
		objectIds := make([]uint64, 0, len(objectIdSet))
		for objectId := range objectIdSet {
			objectIds = append(objectIds, objectId)
		}
		nameToObjectIds[name] = objectIds
	}
	return nameToObjectIds
}

func convertActionsToFlags(actions []pb.RightAction) uint8 {
	var flags uint8
	for _, action := range actions {
		flags |= convertActionToFlag(action)
	}
	return flags
}

func convertActionToFlag(action pb.RightAction) uint8 {
	return 1 << uint8(action)
}
