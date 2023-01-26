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

	"github.com/dvaumoron/puzzlerightserver/model"
	pb "github.com/dvaumoron/puzzlerightservice"
	"gorm.io/gorm"
)

type empty = struct{}

// server is used to implement puzzlerightservice.RightServer.
type server struct {
	pb.UnimplementedRightServer
	db *gorm.DB
}

func New(db *gorm.DB) pb.RightServer {
	return server{db: db}
}

func (s server) AuthQuery(ctx context.Context, request *pb.RightRequest) (*pb.Response, error) {
	var user model.User
	err := s.db.Joins(
		"Roles", "object_id = ?", request.ObjectId,
	).First(
		&user, request.UserId,
	).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// unknown user are not authorized
			return &pb.Response{Success: false}, nil
		}
		return nil, err
	}

	success := false
	requestFlag := convertActionToFlag(request.Action)
	for _, role := range user.Roles {
		success = role.ActionFlags&requestFlag != 0
		if success {
			// the correct right exists
			break
		}
	}
	return &pb.Response{Success: success}, nil
}

func (s server) ListRoles(ctx context.Context, request *pb.ObjectIds) (*pb.Roles, error) {
	var roleNames []model.RoleName
	err := s.db.Joins(
		"Roles", "object_id IN (?)", request.Ids,
	).Find(&roleNames).Error
	if err != nil {
		return nil, err
	}
	return &pb.Roles{List: convertRoleNamesFromModel(roleNames)}, nil
}

func (s server) RoleRight(ctx context.Context, request *pb.RoleRequest) (*pb.Actions, error) {
	var roleName model.RoleName
	err := s.db.Joins(
		"Roles", "object_id = ?", request.ObjectId,
	).First(
		&roleName, "name = ?", request.Name,
	).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// ignore unknown role
			return &pb.Actions{}, nil
		}
		return nil, err
	}

	actions := convertActionsFromFlags(roleName.Roles[0].ActionFlags)
	return &pb.Actions{List: actions}, nil
}

func (s server) UpdateUser(ctx context.Context, request *pb.UserRight) (response *pb.Response, err error) {
	userId := request.UserId
	roles, err := s.loadRoles(request.List)
	if err != nil {
		return
	}
	if len(roles) == 0 {
		// delete unused user
		err = s.db.Delete(&model.User{}, userId).Error
		if err != nil {
			return
		}
		return &pb.Response{Success: true}, nil
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r == nil && err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()
	var user model.User
	if err = tx.First(&user, userId).Error; err == nil {
		err = tx.Model(&user).Association("Roles").Replace(roles)
		if err != nil {
			return
		}
		return &pb.Response{Success: true}, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return
	}
	user = model.User{ID: userId, Roles: roles}
	if err = tx.Save(&user).Error; err != nil {
		return
	}
	return &pb.Response{Success: true}, nil
}

func (s server) UpdateRole(ctx context.Context, request *pb.Role) (response *pb.Response, err error) {
	name := request.Name
	actionFlags := convertActionsToFlags(request.List)
	if actionFlags == 0 {
		// delete unused role
		var roleName model.RoleName
		if err = s.db.First(&roleName, "name = ?", name).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return &pb.Response{Success: true}, nil
			}
			return
		}
		var role model.Role
		err = s.db.First(&role,
			"role_name_id = ? AND object_id = ?", roleName.ID, request.ObjectId,
		).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return &pb.Response{Success: true}, nil
			}
			return
		}

		tx := s.db.Begin()
		defer func() {
			if r := recover(); r == nil && err == nil {
				tx.Commit()
			} else {
				tx.Rollback()
			}
		}()

		if err = tx.Delete(&model.Role{}, role.ID).Error; err != nil {
			return
		}
		if len(roleName.Roles) <= 1 {
			// we have deleted the last role with this name
			err = s.db.Delete(&model.RoleName{}, roleName.ID).Error
			if err != nil {
				return
			}
		}
		return &pb.Response{Success: true}, nil
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r == nil && err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()

	var roleName model.RoleName
	if err = tx.FirstOrCreate(&roleName, model.RoleName{Name: name}).Error; err != nil {
		return
	}
	var role model.Role
	err = tx.FirstOrCreate(&role, model.Role{
		RoleNameID: roleName.ID, ObjectId: request.ObjectId,
	}).Error
	if err != nil {
		return
	}
	if err = tx.Model(&role).Update("action_flags", actionFlags).Error; err != nil {
		return
	}
	return &pb.Response{Success: true}, nil
}

func (s server) ListUserRoles(ctx context.Context, request *pb.UserId) (*pb.Roles, error) {
	var user model.User
	err := s.db.Joins("Roles").First(&user, request.Id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// unknown user, send back an empty role list
			return &pb.Roles{}, nil
		}
		return nil, err
	}

	roles := user.Roles
	var roleNames []model.RoleName
	err = s.db.Find(&roleNames, "id IN (?)", extractRoleNameIds(roles)).Error
	if err != nil {
		return nil, err
	}
	return &pb.Roles{List: convertRolesFromModel(roles, roleNames)}, nil
}

func (s server) loadRoles(roles []*pb.RoleRequest) ([]model.Role, error) {
	resRoles := make([]model.Role, 0, len(roles)) // probably lot more space than necessary
	for name, objectIds := range extractNamesToObjectIds(roles) {
		var roleName model.RoleName
		err := s.db.Joins(
			"Roles", "object_id IN (?)", objectIds,
		).First(
			&roleName, "name = ?", name,
		).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// unknown roleName are ignored
				continue
			}
			return nil, err
		}
		resRoles = append(resRoles, roleName.Roles...)
	}
	return resRoles, nil
}

func convertRoleNamesFromModel(roleNames []model.RoleName) []*pb.Role {
	var resRoles []*pb.Role
	for _, roleName := range roleNames {
		name := roleName.Name
		for _, role := range roleName.Roles {
			resRoles = append(resRoles, &pb.Role{
				Name: name, ObjectId: role.ObjectId,
				List: convertActionsFromFlags(role.ActionFlags),
			})
		}
	}
	return resRoles
}

func convertRolesFromModel(roles []model.Role, roleNames []model.RoleName) []*pb.Role {
	idToName := map[uint64]string{}
	for _, roleName := range roleNames {
		idToName[roleName.ID] = roleName.Name
	}

	resRoles := make([]*pb.Role, 0, len(roles))
	for _, role := range roles {
		resRoles = append(resRoles, &pb.Role{
			Name: idToName[role.RoleNameID], ObjectId: role.ObjectId,
			List: convertActionsFromFlags(role.ActionFlags),
		})
	}
	return resRoles
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

func extractRoleNameIds(roles []model.Role) []uint64 {
	idSet := map[uint64]empty{}
	for _, role := range roles {
		idSet[role.RoleNameID] = empty{}
	}
	ids := make([]uint64, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}
	return ids
}
