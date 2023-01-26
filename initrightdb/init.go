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
	"errors"
	"log"
	"os"
	"strconv"

	dbclient "github.com/dvaumoron/puzzledbclient"
	"github.com/dvaumoron/puzzlerightserver/model"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

const adminGroupId = 1 // groupId corresponding to role administration
const administratorName = "Administrator"
const allActionFlags = 15

const dbErrorMsg = "Database error :"

func main() {
	if len(os.Args) < 2 {
		log.Print("Wait an id for the initial admin user as argument")
	}

	adminUserIdStr := os.Args[1]
	adminUserId, err := strconv.ParseUint(adminUserIdStr, 10, 64)
	if err != nil {
		log.Fatal("Failed to parse the id as an integer")
	}

	err = godotenv.Load()
	if err != nil {
		log.Fatal("Failed to load .env file")
	}

	db := dbclient.Create()

	db.AutoMigrate(&model.User{}, &model.Role{}, &model.RoleName{})

	var roleName model.RoleName
	err = db.Joins(
		"Roles", "object_id = ?", adminGroupId,
	).First(
		&roleName, "name = ?", administratorName,
	).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Fatal(dbErrorMsg, err)
		}

		// the rolename and role doesn't exist, create it
		roleName = model.RoleName{
			Name: administratorName, Roles: []model.Role{
				{ObjectId: adminGroupId, ActionFlags: allActionFlags},
			},
		}
		db.Create(&roleName)
	}

	var user model.User
	err = db.First(&user, adminUserId).Error
	if err == nil {
		// the user already exist, nothing to do
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Fatal(dbErrorMsg, err)
	}

	user = model.User{ID: adminUserId}
	err = db.Save(&user).Error
	if err != nil {
		log.Fatal(dbErrorMsg, err)
	}
	err = db.Model(&user).Association("Roles").Append(roleName.Roles)
	if err != nil {
		log.Fatal(dbErrorMsg, err)
	}
}
