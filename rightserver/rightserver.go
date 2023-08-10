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
	"errors"
	"sync"

	"github.com/dvaumoron/puzzlerightserver/model"
	pb "github.com/dvaumoron/puzzlerightservice"
	"github.com/open-policy-agent/opa/rego"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const RightKey = "puzzleRight"

const publicObjectId = 0

const dbAccessMsg = "Failed to access database"

var errInternal = errors.New("internal service error")

type empty = struct{}

// server is used to implement puzzlerightservice.RightServer.
type server struct {
	pb.UnimplementedRightServer
	db            *gorm.DB
	rule          rego.PreparedEvalQuery
	idToNameMutex sync.RWMutex
	idToName      map[uint64]string
	logger        *otelzap.Logger
}

func New(db *gorm.DB, opaRule rego.PreparedEvalQuery, logger *otelzap.Logger) pb.RightServer {
	db.AutoMigrate(&model.UserRoles{}, &model.Role{}, &model.RoleName{})
	return &server{db: db, rule: opaRule, idToName: map[uint64]string{}, logger: logger}
}

func (s *server) AuthQuery(ctx context.Context, request *pb.RightRequest) (*pb.Response, error) {
	db := s.db.WithContext(ctx)
	logger := s.logger.Ctx(ctx)

	var err error
	var roles []model.Role
	userId := request.UserId
	if userId != 0 {
		subQuery := db.Model(&model.UserRoles{}).Select("role_id").Where(
			"user_id = ?", userId,
		)
		if err = db.Find(&roles, "id in (?)", subQuery).Error; err != nil {
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
	db := s.db.WithContext(ctx)
	logger := s.logger.Ctx(ctx)

	var roles []model.Role
	err := db.Find(&roles, "object_id IN ?", request.Ids).Error
	if err != nil {
		logger.Error(dbAccessMsg, zap.Error(err))
		return nil, errInternal
	}

	resRoles, err := s.convertRolesFromModel(db, logger, roles)
	if err != nil {
		return nil, err
	}
	return &pb.Roles{List: resRoles}, nil
}

func (s *server) RoleRight(ctx context.Context, request *pb.RoleRequest) (*pb.Actions, error) {
	db := s.db.WithContext(ctx)
	logger := s.logger.Ctx(ctx)

	subQuery := db.Model(&model.RoleName{}).Select("id").Where("name = ?", request.Name)

	var role model.Role
	err := db.First(
		&role, "name_id IN (?) AND object_id = ?", subQuery, request.ObjectId,
	).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
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
	db := s.db.WithContext(ctx)
	logger := s.logger.Ctx(ctx)

	roles, err := loadRoles(db, logger, request.List)
	if err != nil {
		return
	}

	userId := request.UserId
	rolesLen := len(roles)
	if rolesLen == 0 {
		// delete unused user
		err = db.Delete(&model.UserRoles{}, "user_id = ?", userId).Error
		if err != nil {
			logger.Error(dbAccessMsg, zap.Error(err))
			return nil, errInternal
		}
		return &pb.Response{Success: true}, nil
	}

	tx := db.Begin()
	defer commitOrRollBack(tx, logger, &err)

	err = tx.Delete(&model.UserRoles{}, "user_id = ?", userId).Error
	if err != nil {
		logger.Error(dbAccessMsg, zap.Error(err))
		return nil, errInternal
	}

	userRoles := make([]model.UserRoles, 0, rolesLen)
	for _, role := range roles {
		userRoles = append(userRoles, model.UserRoles{UserId: userId, RoleId: role.ID})
	}
	if err = tx.Create(&userRoles).Error; err != nil {
		logger.Error(dbAccessMsg, zap.Error(err))
		return nil, errInternal
	}
	return &pb.Response{Success: true}, nil
}

func (s *server) UpdateRole(ctx context.Context, request *pb.Role) (response *pb.Response, err error) {
	db := s.db.WithContext(ctx)
	logger := s.logger.Ctx(ctx)

	name := request.Name
	objectId := request.ObjectId

	if objectId == publicObjectId {
		// right on public part are not updatable, return false (bool default)
		return &pb.Response{}, nil
	}

	actionFlags := convertActionsToFlags(request.List)
	if actionFlags == 0 {
		// delete unused role
		nameSubQuery := db.Model(&model.RoleName{}).Select("id").Where("name = ?", name)
		var role model.Role
		err = db.First(
			&role, "name_id IN (?) AND object_id = ?", nameSubQuery, objectId,
		).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return &pb.Response{Success: true}, nil
			}

			logger.Error(dbAccessMsg, zap.Error(err))
			return nil, errInternal
		}

		if err = db.Delete(&model.Role{}, role.ID).Error; err != nil {
			logger.Error(dbAccessMsg, zap.Error(err))
			return nil, errInternal
		}

		// we delete the names without roles
		roleSubQuery := db.Model(&model.Role{}).Distinct("name_id")
		err = db.Delete(&model.RoleName{}, "id NOT IN (?)", roleSubQuery).Error
		if err != nil {
			logger.Error(dbAccessMsg, zap.Error(err))
			return nil, errInternal
		}

		// invalidate the cache of name
		s.idToNameMutex.Lock()
		s.idToName = map[uint64]string{}
		s.idToNameMutex.Unlock()
		return &pb.Response{Success: true}, nil
	}

	tx := db.Begin()
	defer commitOrRollBack(tx, logger, &err)

	var roleName model.RoleName
	if err = tx.FirstOrCreate(&roleName, model.RoleName{Name: name}).Error; err != nil {
		logger.Error(dbAccessMsg, zap.Error(err))
		return nil, errInternal
	}

	var role model.Role
	err = db.First(&role, "name_id = ? AND object_id = ?", roleName.ID, objectId).Error
	if err == nil {
		if err = tx.Model(&role).Update("action_flags", actionFlags).Error; err != nil {
			logger.Error(dbAccessMsg, zap.Error(err))
			return nil, errInternal
		}

		return &pb.Response{Success: true}, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Error(dbAccessMsg, zap.Error(err))
		return nil, errInternal
	}

	role = model.Role{NameId: roleName.ID, ObjectId: objectId, ActionFlags: actionFlags}
	if err = tx.Create(&role).Error; err != nil {
		logger.Error(dbAccessMsg, zap.Error(err))
		return nil, errInternal
	}
	return &pb.Response{Success: true}, nil
}

func (s *server) ListUserRoles(ctx context.Context, request *pb.UserId) (*pb.Roles, error) {
	db := s.db.WithContext(ctx)
	logger := s.logger.Ctx(ctx)

	subQuery := db.Model(&model.UserRoles{}).Select("role_id").Where("user_id = ?", request.Id)

	var roles []model.Role
	err := db.Find(&roles, "id IN (?)", subQuery).Error
	if err != nil {
		logger.Error(dbAccessMsg, zap.Error(err))
		return nil, errInternal
	}

	resRoles, err := s.convertRolesFromModel(db, logger, roles)
	if err != nil {
		logger.Error(dbAccessMsg, zap.Error(err))
		return nil, errInternal
	}
	return &pb.Roles{List: resRoles}, nil
}

func loadRoles(db *gorm.DB, logger otelzap.LoggerWithCtx, roles []*pb.RoleRequest) ([]model.Role, error) {
	resRoles := make([]model.Role, 0, len(roles)) // probably lot more space than necessary
	for name, objectIds := range extractNamesToObjectIds(roles) {
		subQuery := db.Model(&model.RoleName{}).Select("id").Where("name = ?", name)

		var roles []model.Role
		err := db.Find(&roles, "name_id IN (?) AND object_id IN ?", subQuery, objectIds).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}

			logger.Error(dbAccessMsg, zap.Error(err))
			return nil, errInternal
		}
		if len(roles) != 0 {
			resRoles = append(resRoles, roles...)
		}
	}
	return resRoles, nil
}

func (s *server) convertRolesFromModel(db *gorm.DB, logger otelzap.LoggerWithCtx, roles []model.Role) ([]*pb.Role, error) {
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

	var roleNames []model.RoleName
	if err := db.Find(&roleNames, "id IN ?", queryIds).Error; err != nil {
		logger.Error(dbAccessMsg, zap.Error(err))
		return nil, errInternal
	}

	for _, roleName := range roleNames {
		s.idToName[roleName.ID] = roleName.Name
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

func commitOrRollBack(tx *gorm.DB, logger otelzap.LoggerWithCtx, err *error) {
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
