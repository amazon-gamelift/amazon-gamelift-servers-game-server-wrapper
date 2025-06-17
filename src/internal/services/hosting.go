/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"log/slog"

	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/hosting/gamelift/initialiser"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/hosting/gamelift/sdk"

	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/internal/config"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/hosting"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/hosting/gamelift"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/observability"
)

func getHosting(ctx context.Context, cfg *config.Config, logger *slog.Logger, spanner observability.Spanner) (hosting.Service, error) {
	logger.DebugContext(ctx, "Initializing Amazon GameLift hosting service")
	return gamelift.New(ctx, &gamelift.Config{
		GamePort:               cfg.Ports.GamePort,
		Anywhere:               cfg.Hosting.GameLift.Anywhere,
		LogDirectory:           cfg.Hosting.LogDirectory,
		GameServerLogDirectory: cfg.Hosting.AbsoluteGameServerLogDirectory,
	},
		logger,
		spanner,
		&initialiser.InitialiserServiceFactory{},
		sdk.NewSdk(ctx, logger),
	)
}
