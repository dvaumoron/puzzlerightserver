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
package dbclient

import (
	"log"
	"os"
	"strings"

	"github.com/glebarez/sqlite" // driver without cgo
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Create() *gorm.DB {
	kind := strings.ToLower(os.Getenv("DB_SERVER_TYPE"))
	addr := os.Getenv("DB_SERVER_ADDR")
	var dialector gorm.Dialector
	switch kind {
	case "sqlite":
		dialector = sqlite.Open(addr)
	case "postgres":
		dialector = postgres.Open(addr)
	default:
		log.Fatalf("Unknown database type : %v", kind)
	}
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		log.Fatalf("Database connection failed : %v", err)
	}
	return db
}
