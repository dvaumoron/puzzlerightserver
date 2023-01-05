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
package main

import (
	"log"
	"os"
	"strconv"

	"github.com/dvaumoron/puzzlerightserver/dbclient"
	"github.com/dvaumoron/puzzlerightserver/model"
	pb "github.com/dvaumoron/puzzlerightservice"
	"github.com/joho/godotenv"
)

const adminGroupId = 1 // groupId corresponding to role administration

func main() {
	adminUserIdStr := os.Args[1] // id
	adminUserId, err := strconv.ParseUint(adminUserIdStr, 10, 64)

	err = godotenv.Load()
	if err != nil {
		log.Fatal("Failed to load .env file")
	}

	db := dbclient.Create(os.Getenv("DB_SERVER_TYPE"), os.Getenv("DB_SERVER_ADDR"))

	db.AutoMigrate(&model.User{}, &model.Role{}, &model.Action{}, &model.RoleName{})

	ActionAccess := model.Action{ID: uint8(pb.RightAction_ACCESS)}
	ActionCreate := model.Action{ID: uint8(pb.RightAction_ACCESS)}
	ActionUpdate := model.Action{ID: uint8(pb.RightAction_ACCESS)}
	ActionDelete := model.Action{ID: uint8(pb.RightAction_ACCESS)}

	db.Create(&ActionAccess)
	db.Create(&ActionCreate)
	db.Create(&ActionUpdate)
	db.Create(&ActionDelete)

	adminName := &model.RoleName{Name: "Admin"}
	adminRole := &model.Role{
		RoleName: adminName, ObjectId: adminGroupId,
		Actions: []model.Action{ActionAccess, ActionCreate, ActionUpdate, ActionDelete},
	}
	db.Create(&model.User{ID: adminUserId, Roles: []*model.Role{adminRole}})
}
