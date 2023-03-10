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

type Role struct {
	ID          uint64
	NameId      uint64
	ObjectId    uint64
	ActionFlags uint8
}

type UserRoles struct {
	ID     uint64
	UserId uint64
	RoleId uint64
}
