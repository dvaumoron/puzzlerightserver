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

	"github.com/dvaumoron/puzzlerightserver/model"
	pb "github.com/dvaumoron/puzzlerightservice"
	"gorm.io/gorm"
)

// Server is used to implement puzzlerightservice.RightServer.
type Server struct {
	pb.UnimplementedRightServer
	DB *gorm.DB
}

func (s *Server) AuthQuery(ctx context.Context, request *pb.RightRequest) (*pb.Response, error) {
	user := model.User{ID: request.UserId}
	err := s.DB.Joins(
		"Roles", s.DB.Where(&model.Role{ObjectId: request.ObjectId}),
	).Joins(
		"Roles.Actions", s.DB.Where(&model.Action{ID: uint8(request.Action)}),
	).First(&user).Error
	var response *pb.Response
	if err == nil {
		success := false
		for _, role := range user.Roles {
			if len(role.Actions) != 0 {
				// if we reach here the correct result exists
				success = true
				break
			}
		}
		response = &pb.Response{Success: success}
	}
	return response, err
}

func (s *Server) ListRoles(ctx context.Context, request *pb.ObjectIds) (*pb.Roles, error) {
	var roles []*model.Role
	err := s.DB.Joins("RoleName").Joins("Actions").Where(
		"object_id IN (?)", request.Ids,
	).Find(&roles).Error
	var response *pb.Roles
	if err == nil {
		response = &pb.Roles{List: convertRolesFromModel(roles)}
	}
	return response, err
}

func (s *Server) RoleRight(ctx context.Context, request *pb.RoleRequest) (*pb.Actions, error) {
	roleName, err := loadOrCreateRoleName(s.DB, request.Name)
	var actions *pb.Actions
	if err == nil {
		var role *model.Role
		role, err = loadRole(s.DB, roleName.ID, request.ObjectId)
		if err == nil {
			actions = &pb.Actions{List: convertActionsFromModel(role.Actions)}
		}
	}
	return actions, err
}

func (s *Server) UpdateUser(ctx context.Context, request *pb.UserRight) (*pb.Response, error) {
	user := &model.User{ID: request.UserId}
	err := s.DB.First(user).Error
	if err == gorm.ErrRecordNotFound {
		if err = s.DB.Create(user).Error; err == nil {
			err = updateUser(s.DB, user, request.List)
		}
	} else if err == nil {
		err = updateUser(s.DB, user, request.List)
	}
	return &pb.Response{Success: err == nil}, nil
}

func (s *Server) UpdateRole(ctx context.Context, request *pb.Role) (*pb.Response, error) {
	roleName, err := loadOrCreateRoleName(s.DB, request.Name)
	if err == nil {
		objectId := request.ObjectId
		role, err := loadRole(s.DB, roleName.ID, objectId)
		if err == gorm.ErrRecordNotFound {
			role = &model.Role{RoleName: roleName, ObjectId: objectId}
			if err = s.DB.Create(role).Error; err == nil {
				updateRole(s.DB, role, request.List)
			}
		} else if err == nil {
			updateRole(s.DB, role, request.List)
		}
	}
	return &pb.Response{Success: err == nil}, nil
}

func (s *Server) ListUserRoles(ctx context.Context, request *pb.UserId) (*pb.Roles, error) {
	user := model.User{ID: request.Id}
	err := s.DB.Joins("Roles").Joins("Roles.RoleName").Joins("Roles.Actions").First(&user).Error
	var roles *pb.Roles
	if err == nil {
		roles = &pb.Roles{List: convertRolesFromModel(user.Roles)}
	}
	return roles, err
}

func convertRolesFromModel(roles []*model.Role) []*pb.Role {
	var resRoles []*pb.Role
	for _, role := range roles {
		resRoles = append(resRoles, &pb.Role{
			Name: role.RoleName.Name, ObjectId: role.ObjectId,
			List: convertActionsFromModel(role.Actions),
		})
	}
	return resRoles
}

func convertActionsFromModel(actions []model.Action) []pb.RightAction {
	resActions := []pb.RightAction{}
	for _, action := range actions {
		resActions = append(resActions, pb.RightAction(action.ID))
	}
	return resActions
}

func updateUser(db *gorm.DB, user *model.User, list []*pb.RoleRequest) error {
	roles, err := loadRoles(db, list)
	if err == nil {
		user.Roles = roles
		err = db.Save(user).Error
	}
	return err
}

func loadRoles(db *gorm.DB, roles []*pb.RoleRequest) ([]*model.Role, error) {
	resRoles := []*model.Role{}
	var err error
	for _, role := range roles {
		var roleName *model.RoleName
		roleName, err = loadOrCreateRoleName(db, role.Name)
		if err != nil {
			break
		}

		var loadedRole *model.Role
		loadedRole, err = loadRole(db, roleName.ID, role.ObjectId)
		if err != nil {
			break
		}
		resRoles = append(resRoles, loadedRole)
	}
	return resRoles, err
}

func loadRole(db *gorm.DB, nameId uint64, objectId uint64) (*model.Role, error) {
	role := &model.Role{}
	err := db.Joins("Actions").Where(
		&model.Role{RoleNameID: nameId, ObjectId: objectId},
	).First(role).Error
	return role, err
}

func loadOrCreateRoleName(db *gorm.DB, name string) (*model.RoleName, error) {
	roleName := &model.RoleName{}
	query := &model.RoleName{Name: name}
	err := db.Where(query).First(roleName).Error
	if err == gorm.ErrRecordNotFound {
		roleName = query
		err = db.Create(roleName).Error
	}
	return roleName, err
}

func updateRole(db *gorm.DB, role *model.Role, actions []pb.RightAction) error {
	role.Actions = convertActionsFromPB(actions)
	return db.Save(role).Error
}

func convertActionsFromPB(actions []pb.RightAction) []model.Action {
	resActions := []model.Action{}
	for _, action := range actions {
		resActions = append(resActions, model.Action{ID: uint8(action)})
	}
	return resActions
}
