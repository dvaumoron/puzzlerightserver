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
	"gorm.io/gorm"
)

type empty = struct{}

// server is used to implement puzzlerightservice.RightServer.
type server struct {
	pb.UnimplementedRightServer
	db            *gorm.DB
	idToNameMutex sync.RWMutex
	idToName      map[uint64]string
}

func New(db *gorm.DB) pb.RightServer {
	return &server{db: db}
}

func (s *server) AuthQuery(ctx context.Context, request *pb.RightRequest) (*pb.Response, error) {
	subQuery := s.db.Model(&model.UserRoles{}).Select("role_id").Where(
		"user_id = ?", request.UserId,
	)
	var roles []model.Role
	err := s.db.Find(&roles, "id in (?)", subQuery).Error
	if err != nil {
		return nil, err
	}

	success := false
	requestFlag := convertActionToFlag(request.Action)
	for _, role := range roles {
		success = role.ActionFlags&requestFlag != 0
		if success {
			// the correct right exists
			break
		}
	}
	return &pb.Response{Success: success}, nil
}

func (s *server) ListRoles(ctx context.Context, request *pb.ObjectIds) (*pb.Roles, error) {
	var roles []model.Role
	err := s.db.Find(&roles, "object_id IN ?", request.Ids).Error
	if err != nil {
		return nil, err
	}

	resRoles, err := s.convertRolesFromModel(roles)
	if err != nil {
		return nil, err
	}
	return &pb.Roles{List: resRoles}, nil
}

func (s *server) RoleRight(ctx context.Context, request *pb.RoleRequest) (*pb.Actions, error) {
	subQuery := s.db.Model(&model.RoleName{}).Select("id").Where("name = ?", request.Name)

	var role model.Role
	err := s.db.First(
		&role, "name_id = (?) AND object_id = ?", subQuery, request.ObjectId,
	).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// ignore unknown role
			return &pb.Actions{}, nil
		}
		return nil, err
	}

	actions := convertActionsFromFlags(role.ActionFlags)
	return &pb.Actions{List: actions}, nil
}

func (s *server) UpdateUser(ctx context.Context, request *pb.UserRight) (response *pb.Response, err error) {
	roles, err := s.loadRoles(request.List)
	if err != nil {
		return
	}

	userId := request.UserId
	rolesLen := len(roles)
	if rolesLen == 0 {
		// delete unused user
		err = s.db.Delete(&model.UserRoles{}, "user_id = ?", userId).Error
		if err != nil {
			return
		}
		return &pb.Response{Success: true}, nil
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			if err2, ok := r.(error); ok {
				err = err2
			} else {
				panic(r)
			}
		} else if err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()

	err = tx.Delete(&model.UserRoles{}, "user_id = ?", userId).Error
	if err != nil {
		return
	}

	userRoles := make([]model.UserRoles, 0, rolesLen)
	for _, role := range roles {
		userRoles = append(userRoles, model.UserRoles{UserId: userId, RoleId: role.ID})
	}
	if err = tx.Create(&userRoles).Error; err != nil {
		return
	}
	return &pb.Response{Success: true}, nil
}

func (s *server) UpdateRole(ctx context.Context, request *pb.Role) (response *pb.Response, err error) {
	var tx *gorm.DB
	commitOrRollback := func() {
		if r := recover(); r != nil {
			tx.Rollback()
			if err2, ok := r.(error); ok {
				err = err2
			} else {
				panic(r)
			}
		} else if err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}

	name := request.Name
	objectId := request.ObjectId
	actionFlags := convertActionsToFlags(request.List)
	if actionFlags == 0 {
		// delete unused role
		subQuery := s.db.Model(&model.RoleName{}).Select("id").Where("name = ?", name)
		var role model.Role
		err = s.db.First(
			&role, "name_id IN (?) AND object_id = ?", subQuery, objectId,
		).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return &pb.Response{Success: true}, nil
			}
			return
		}

		tx = s.db.Begin()
		defer commitOrRollback()

		if err = tx.Delete(&model.Role{}, role.ID).Error; err != nil {
			return
		}

		var roles []model.Role
		err = s.db.Find(&roles, "name_id IN (?)", subQuery).Error
		if err != nil {
			return
		}
		if len(roles) == 0 {
			// we have deleted the last role with this name
			err = tx.Delete(&model.RoleName{}, "name = ?", name).Error
			if err != nil {
				return
			}
		}
		return &pb.Response{Success: true}, nil
	}

	tx = s.db.Begin()
	defer commitOrRollback()

	var roleName model.RoleName
	if err = tx.FirstOrCreate(&roleName, model.RoleName{Name: name}).Error; err != nil {
		return
	}
	var role model.Role
	err = s.db.First(
		&role, "name_id = ? AND object_id = ?", roleName.ID, objectId,
	).Error
	if err == nil {
		if err = tx.Model(&role).Update("action_flags", actionFlags).Error; err != nil {
			return
		}
		return &pb.Response{Success: true}, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return
	}
	role = model.Role{NameId: role.NameId, ObjectId: objectId, ActionFlags: actionFlags}
	if err = tx.Create(&role).Error; err != nil {
		return
	}
	return &pb.Response{Success: true}, nil
}

func (s *server) ListUserRoles(ctx context.Context, request *pb.UserId) (*pb.Roles, error) {
	subQuery := s.db.Model(&model.UserRoles{}).Select("role_id").Where(
		"user_id = ?", request.Id,
	)

	var roles []model.Role
	err := s.db.Find(&roles, "id IN (?)", subQuery).Error
	if err != nil {
		return nil, err
	}

	resRoles, err := s.convertRolesFromModel(roles)
	if err != nil {
		return nil, err
	}
	return &pb.Roles{List: resRoles}, nil
}

func (s *server) loadRoles(roles []*pb.RoleRequest) ([]model.Role, error) {
	resRoles := make([]model.Role, 0, len(roles)) // probably lot more space than necessary
	for name, objectIds := range extractNamesToObjectIds(roles) {
		subQuery := s.db.Model(&model.RoleName{}).Select("id").Where("name = ?", name)

		var roles []model.Role
		err := s.db.First(
			&roles, "name_id IN (?) AND object_id IN ?", subQuery, objectIds,
		).Error
		if err != nil {
			return nil, err
		}
		if len(roles) != 0 {
			resRoles = append(resRoles, roles...)
		}
	}
	return resRoles, nil
}

func (s *server) convertRolesFromModel(roles []model.Role) ([]*pb.Role, error) {
	idSet := map[uint64]empty{}
	s.idToNameMutex.RLock()
	for _, role := range roles {
		id := role.NameId
		_, ok := s.idToName[id]
		if !ok {
			idSet[id] = empty{}
		}
	}
	s.idToNameMutex.RUnlock()

	if len(idSet) != 0 {
		queryIds := make([]uint64, 0, len(idSet))
		for id := range idSet {
			queryIds = append(queryIds, id)
		}

		var roleNames []model.RoleName
		err := s.db.Find(&roleNames, "id IN ?", queryIds).Error
		if err != nil {
			return nil, err
		}
		s.idToNameMutex.Lock()
		for _, roleName := range roleNames {
			s.idToName[roleName.ID] = roleName.Name
		}
		s.idToNameMutex.Unlock()
	}

	resRoles := make([]*pb.Role, 0, len(roles))
	for _, role := range roles {
		s.idToNameMutex.RLock()
		name := s.idToName[role.NameId]
		s.idToNameMutex.RUnlock()
		resRoles = append(resRoles, &pb.Role{
			Name: name, ObjectId: role.ObjectId,
			List: convertActionsFromFlags(role.ActionFlags),
		})
	}
	return resRoles, nil
}

func convertActionsFromFlags(actionFlags uint8) []pb.RightAction {
	var resActions []pb.RightAction
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
		}
		objectIdSet[role.ObjectId] = empty{}
		nameToObjectIdSet[name] = objectIdSet
	}
	nameToObjectIds := map[string][]uint64{}
	for name, objectIdSet := range nameToObjectIdSet {
		nameToObjectIds[name] = make([]uint64, 0, len(objectIdSet))
		for objectId := range objectIdSet {
			nameToObjectIds[name] = append(nameToObjectIds[name], objectId)
		}
	}
	return nameToObjectIds
}

func convertActionsToFlags(actions []pb.RightAction) uint8 {
	var flags uint8
	for _, action := range actions {
		flags &= convertActionToFlag(action)
	}
	return flags
}

func convertActionToFlag(action pb.RightAction) uint8 {
	return 1 << uint8(action)
}
