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
package model

type RoleName struct {
	ID   uint64
	Name string
}

type Action struct {
	ID uint8 `gorm:"auto_increment:false;"`
}

type Role struct {
	ID         uint64
	RoleNameID uint64
	RoleName   *RoleName
	ObjectId   uint64
	Actions    []Action `gorm:"many2many:role_actions;"`
}

type User struct {
	ID    uint64
	Roles []*Role `gorm:"many2many:user_roles;"`
}