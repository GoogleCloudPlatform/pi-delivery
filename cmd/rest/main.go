// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"os"

	"github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
	server "github.com/googlecloudplatform/pi-delivery"
	"go.ajitem.com/zapdriver"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()
	l, err := zapdriver.NewDevelopment()
	if err != nil {
		zap.S().Fatalf("zap.NewDevelopment failed: %v", err)
	}
	defer l.Sync()
	zap.ReplaceGlobals(l)

	if err := funcframework.RegisterHTTPFunctionContext(ctx, "/", server.Get); err != nil {
		l.Sugar().Fatalf("funcframework.RegisterHTTPFunctionContext: %v\n", err)
	}
	// Use PORT environment variable, or default to 8080.
	port := "8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}
	if err := funcframework.Start(port); err != nil {
		l.Sugar().Fatalf("funcframework.Start: %v\n", err)
	}
}
