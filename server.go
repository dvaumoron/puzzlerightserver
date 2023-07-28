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
	"context"
	_ "embed"

	dbclient "github.com/dvaumoron/puzzledbclient"
	grpcserver "github.com/dvaumoron/puzzlegrpcserver"
	"github.com/dvaumoron/puzzlerightserver/rightserver"
	pb "github.com/dvaumoron/puzzlerightservice"
	"github.com/open-policy-agent/opa/sdk"
	"go.uber.org/zap"
)

//go:embed version.txt
var version string

func main() {
	// should start with this, to benefit from the call to godotenv
	s := grpcserver.Make(rightserver.RightKey, version)

	ctx := context.Background()
	opa, err := sdk.New(ctx, sdk.Options{})
	if err != nil {
		s.Logger.FatalContext(ctx, "Failed to initialize OPA client", zap.Error(err))
	}
	defer opa.Stop(ctx)

	pb.RegisterRightServer(s, rightserver.New(dbclient.Create(s.Logger), s.Logger))
	s.Start()
}
