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

// Server is used to implement puzzlerightservice.RightServer.
type Server struct {
	pb.UnimplementedRightServer
	db *gorm.DB
}

func New(db *gorm.DB) *Server {
	return &Server{db: db}
}

func (s *Server) AuthQuery(ctx context.Context, request *pb.RightRequest) (*pb.Response, error) {
	var user model.User
	err := s.db.Joins(
		"Roles", "object_id = ?", request.ObjectId,
	).Joins(
		"Roles.Actions", "id = ?", uint8(request.Action),
	).First(&user, request.UserId).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// unknown user are not authorized
			return &pb.Response{Success: false}, nil
		}
		return nil, err
	}

	success := false
	for _, role := range user.Roles {
		if success = len(role.Actions) != 0; success {
			// if we reach here the correct result exists
			break
		}
	}
	return &pb.Response{Success: success}, nil
}

func (s *Server) ListRoles(ctx context.Context, request *pb.ObjectIds) (*pb.Roles, error) {
	var roleNames []*model.RoleName
	err := s.db.Joins(
		"Roles", "object_id IN (?)", request.Ids,
	).Joins("Roles.Actions").Find(&roleNames).Error
	if err != nil {
		return nil, err
	}
	return &pb.Roles{List: convertRolesFromModel(roleNames)}, nil
}

func (s *Server) RoleRight(ctx context.Context, request *pb.RoleRequest) (*pb.Actions, error) {
	roleName, err := loadRole(s.db, request.Name, request.ObjectId)
	if err != nil {
		return nil, err
	}

	actions := &pb.Actions{}
	if roles := roleName.Roles; len(roles) != 0 {
		actions.List = convertActionsFromModel(roles[0].Actions)
	}
	return actions, nil
}

func (s *Server) UpdateUser(ctx context.Context, request *pb.UserRight) (*pb.Response, error) {
	userId := request.UserId
	roles, err := loadRoles(s.db, request.List)
	if err != nil {
		return nil, err
	}
	if len(roles) == 0 {
		// delete unused user
		err = s.db.Delete(&model.User{}, userId).Error
		if err != nil {
			return nil, err
		}
		return &pb.Response{Success: true}, nil
	}

	var user model.User
	err = s.db.FirstOrCreate(&user, model.User{ID: userId}).Error
	if err != nil {
		return nil, err
	}
	err = s.db.Model(&user).Association("Roles").Replace(roles)
	if err != nil {
		return nil, err
	}
	return &pb.Response{Success: true}, nil
}

func (s *Server) UpdateRole(ctx context.Context, request *pb.Role) (*pb.Response, error) {
	name := request.Name
	actions := convertActionsFromRequest(request.List)
	if len(actions) == 0 {
		// delete unused role
		var roleName model.RoleName
		err := s.db.First(&roleName, "name = ?", name).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return &pb.Response{Success: true}, nil
			}
			return nil, err
		}
		var role model.Role
		err = s.db.First(&role,
			"role_name_id = ? AND object_id = ?", roleName.ID, request.ObjectId,
		).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return &pb.Response{Success: true}, nil
			}
			return nil, err
		}
		err = s.db.Delete(&model.Role{}, role.ID).Error
		if err != nil {
			return nil, err
		}
		if len(roleName.Roles) <= 1 {
			// we have deleted the last role with this name
			err = s.db.Delete(&model.RoleName{}, roleName.ID).Error
			if err != nil {
				return nil, err
			}
		}
		return &pb.Response{Success: true}, nil
	}

	var roleName model.RoleName
	err := s.db.FirstOrCreate(&roleName, model.RoleName{Name: name}).Error
	if err != nil {
		return nil, err
	}
	var role model.Role
	err = s.db.FirstOrCreate(&role, model.Role{
		RoleNameID: roleName.ID, ObjectId: request.ObjectId,
	}).Error
	if err != nil {
		return nil, err
	}
	err = s.db.Model(&role).Association("Actions").Replace(actions)
	if err != nil {
		return nil, err
	}
	return &pb.Response{Success: true}, nil
}

func (s *Server) ListUserRoles(ctx context.Context, request *pb.UserId) (*pb.Roles, error) {
	var user model.User
	err := s.db.Joins("Roles").First(&user, request.Id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// unknown user, send back an empty role list
			return &pb.Roles{}, nil
		}
		return nil, err
	}

	var roleNames []*model.RoleName
	err = s.db.Joins(
		"Roles", "id IN (?)", extractRoleIds(user.Roles),
	).Joins("Roles.Actions").Find(&roleNames).Error
	if err != nil {
		return nil, err
	}
	return &pb.Roles{List: convertRolesFromModel(roleNames)}, nil
}

func convertRolesFromModel(roleNames []*model.RoleName) []*pb.Role {
	var resRoles []*pb.Role
	for _, roleName := range roleNames {
		for _, role := range roleName.Roles {
			resRoles = append(resRoles, &pb.Role{
				Name: roleName.Name, ObjectId: role.ObjectId,
				List: convertActionsFromModel(role.Actions),
			})
		}
	}
	return resRoles
}

func convertActionsFromModel(actions []model.Action) []pb.RightAction {
	var resActions []pb.RightAction
	for _, action := range actions {
		resActions = append(resActions, pb.RightAction(action.ID))
	}
	return resActions
}

func loadRoles(db *gorm.DB, roles []*pb.RoleRequest) ([]*model.Role, error) {
	var resRoles []*model.Role
	for name, objectIds := range extractNamesToObjectIds(roles) {
		var roleName model.RoleName
		err := db.Joins(
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
		for objectId := range objectIdSet {
			nameToObjectIds[name] = append(nameToObjectIds[name], objectId)
		}
	}
	return nameToObjectIds
}

func loadRole(db *gorm.DB, name string, objectId uint64) (*model.RoleName, error) {
	var roleName model.RoleName
	err := db.Joins(
		"Roles", "object_id = ?", objectId,
	).Joins("Roles.Actions").First(
		&roleName, "name = ?", name,
	).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// ignore unknown role
			return nil, nil
		}
		return nil, err
	}
	return &roleName, nil
}

func convertActionsFromRequest(actions []pb.RightAction) []model.Action {
	var resActions []model.Action
	for _, action := range actions {
		resActions = append(resActions, model.Action{ID: uint8(action)})
	}
	return resActions
}

func extractRoleIds(roles []*model.Role) []uint64 {
	var ids []uint64
	for _, role := range roles {
		ids = append(ids, role.ID)
	}
	return ids
}
