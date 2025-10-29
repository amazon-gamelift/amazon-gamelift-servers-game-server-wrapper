/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package initialiser

import (
	"context"
	"log/slog"
	"sync"

	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/hosting/gamelift/sdk"
	"github.com/amazon-gamelift/amazon-gamelift-servers-go-server-sdk/v5/server"
)

type managed struct {
	sdk    sdk.GameLiftSdk
	mutex  sync.Mutex
	logger *slog.Logger
}

// InitSdk initializes the GameLift SDK for managed EC2 or managed Containers environments.
func (managed *managed) InitSdk(ctx context.Context) error {
	managed.logger.InfoContext(ctx, "using Amazon GameLift Managed EC2 or Managed Containers")
	params := server.ServerParameters{}
	return managed.sdk.InitSDK(ctx, params)
}

func newManaged(sdk sdk.GameLiftSdk, logger *slog.Logger) Service {
	a := &managed{
		sdk:    sdk,
		logger: logger,
	}

	return a
}
