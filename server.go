/*
 *
 * Copyright 2022 puzzlesessionserver authors.
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
	"context"
	"log"
	"net"
	"os"
	"strings"

	pb "github.com/dvaumoron/puzzlerightservice"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// server is used to implement puzzlerightservice.RightServer.
type server struct {
	pb.UnimplementedRightServer
	db *gorm.DB
}

func (s *server) AuthQuery(context.Context, *pb.RightRequest) (*pb.Response, error) {
	return nil, nil
}

func (s *server) ListRoles(context.Context, *pb.ObjectIds) (*pb.Roles, error) {
	return nil, nil
}

func (s *server) RoleRight(context.Context, *pb.RoleRequest) (*pb.Actions, error) {
	return nil, nil
}

func (s *server) UpdateUser(context.Context, *pb.UserRight) (*pb.Response, error) {
	return nil, nil
}

func (s *server) UpdateRole(context.Context, *pb.Role) (*pb.Response, error) {
	return nil, nil
}

func (s *server) ListUserRoles(context.Context, *pb.UserId) (*pb.Roles, error) {
	return nil, nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Failed to load .env file")
	}

	lis, err := net.Listen("tcp", ":"+os.Getenv("SERVICE_PORT"))
	if err != nil {
		log.Fatalf("Failed to listen : %v", err)
	}

	db := createDB(os.Getenv("DB_SERVER_TYPE"), os.Getenv("DB_SERVER_ADDR"))

	s := grpc.NewServer()
	pb.RegisterRightServer(s, &server{db: db})
	log.Printf("Listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve : %v", err)
	}
}

func createDB(kind, addr string) *gorm.DB {
	var dialector gorm.Dialector
	switch strings.ToLower(kind) {
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
