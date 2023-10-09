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
	"database/sql"
	_ "embed"
	"os"

	grpcserver "github.com/dvaumoron/puzzlegrpcserver"
	"github.com/dvaumoron/puzzlerightserver/rightserver"
	pb "github.com/dvaumoron/puzzlerightservice"
	_ "github.com/jackc/pgx/v5"
	"github.com/open-policy-agent/opa/rego"
	"go.uber.org/zap"
)

//go:embed version.txt
var version string

func main() {
	// should start with this, to benefit from the call to godotenv
	ctx, initSpan, s := grpcserver.Init(rightserver.RightKey, version)

	data, err := os.ReadFile(os.Getenv("OPA_MODULE_FILE"))
	if err != nil {
		s.Logger.FatalContext(ctx, "Failed to load OPA module", zap.Error(err))
	}

	rule := rego.New(
		rego.Query("data.auth.allow"),
		rego.Module("auth.rego", string(data)),
	)

	query, err := rule.PrepareForEval(ctx)
	if err != nil {
		s.Logger.FatalContext(ctx, "Failed to initialize OPA module", zap.Error(err))
	}
	initSpan.End()

	db, err := sql.Open("postgres", os.Getenv("DB_SERVER_ADDR"))
	if err != nil {
		s.Logger.FatalContext(ctx, "Failed to initialize DB", zap.Error(err))
	}
	defer db.Close()

	pb.RegisterRightServer(s, rightserver.New(db, query, s.Logger))
	s.Start(ctx)
}
