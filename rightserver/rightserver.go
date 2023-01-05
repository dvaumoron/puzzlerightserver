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
	user := &model.User{ID: request.UserId}
	err := s.DB.Preload("Roles").First(user).Error
	var response *pb.Response
	if err == nil {
		rObjectId := request.ObjectId
		rAction := uint8(request.Action)

		success := false
	RolesLoop:
		for _, role := range user.Roles {
			if role.ObjectId == rObjectId {
				err = s.DB.Preload("Actions").First(role).Error
				if err != nil {
					break RolesLoop
				}

				for _, action := range role.Actions {
					if action.ID == rAction {
						success = true
						break RolesLoop
					}
				}
			}
		}
		response = &pb.Response{Success: success}
	}
	return response, err
}

func (s *Server) ListRoles(ctx context.Context, request *pb.ObjectIds) (*pb.Roles, error) {
	ids := request.Ids
	var roles []model.Role
	err := s.DB.Preload("RoleName").Preload("Actions").Where("ObjectId IN (?)", ids).Find(&roles).Error
	var response *pb.Roles
	if err == nil {
		var list []*pb.Role
		for _, role := range roles {
			list = append(list, &pb.Role{
				Name: role.RoleName.Name, ObjectId: role.ObjectId,
				List: convertActionsFromModel(role.Actions),
			})
		}
		response = &pb.Roles{List: list}
	}
	return response, err
}

func (s *Server) RoleRight(ctx context.Context, request *pb.RoleRequest) (*pb.Actions, error) {
	var role model.Role
	err := s.DB.Preload("Actions").Where("name = ? AND object_id = ?", request.Name, request.ObjectId).First(&role).Error
	var actions *pb.Actions
	if err == nil {
		actions = &pb.Actions{List: convertActionsFromModel(role.Actions)}
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
	name := request.Name
	objectId := request.ObjectId
	role, err := loadRole(s.DB, name, objectId)
	if err == gorm.ErrRecordNotFound {
		roleName, err := loadOrCreateRoleName(s.DB, name)
		if err == nil {
			role = &model.Role{RoleName: roleName, ObjectId: objectId}
			err = s.DB.Create(role).Error
			if err == nil {
				updateRole(s.DB, role, request.List)
			}
		}
	} else if err == nil {
		updateRole(s.DB, role, request.List)
	}
	return &pb.Response{Success: err == nil}, err
}

func (s *Server) ListUserRoles(ctx context.Context, request *pb.UserId) (*pb.Roles, error) {
	// TODO
	return nil, nil
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
		var loadedRole *model.Role
		loadedRole, err = loadRole(db, role.Name, role.ObjectId)
		if err != nil {
			break
		}
		resRoles = append(resRoles, loadedRole)
	}
	return resRoles, err
}

func loadRole(db *gorm.DB, name string, objectId uint64) (*model.Role, error) {
	loadedRole := &model.Role{}
	err := db.Preload("RoleName").Preload("Actions").Where("name = ? AND object_id = ?", name, objectId).First(loadedRole).Error
	return loadedRole, err
}

func loadOrCreateRoleName(db *gorm.DB, name string) (*model.RoleName, error) {
	roleName := &model.RoleName{}
	err := db.Where("name = ?", name).First(roleName).Error
	if err == gorm.ErrRecordNotFound {
		roleName.Name = name
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
