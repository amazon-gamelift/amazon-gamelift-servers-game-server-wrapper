/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/internal/config"
	"github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/internal/multiplexgame"
	"github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/pkg/game"
	"github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/pkg/logging"
	"github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/pkg/observability"
	"github.com/pkg/errors"
)

func getGame(ctx context.Context, cfg *config.Config, logger *slog.Logger, gl logging.Game, spanner observability.Spanner) (game.Server, error) {
	if cfg == nil {
		return nil, errors.New("Configuration not provided when getting the game")
	}

	var err error

	// Clean paths and build a path to be checked
	pathToCheck := filepath.Clean(cfg.BuildDetail.RelativeExePath)
	if !filepath.IsAbs(pathToCheck) {
		pathToCheck = filepath.Join(filepath.Clean(cfg.BuildDetail.WorkingDir), pathToCheck)
	}

	// Check if game server executable file exists
	if _, err := os.Stat(pathToCheck); os.IsNotExist(err) {
		return nil, errors.Errorf("Game server executable not found at path: %s", pathToCheck)
	}

	multiplexGame, err := multiplexgame.New(*cfg, logger, gl, spanner)

	if err != nil {
		return nil, errors.Wrap(err, "Failed to initialize multiplex game")
	}

	return multiplexGame, nil
}
