/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cmd

import (
	"io"
	"log/slog"
	"os"

	"github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/internal"
	"github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/internal/config"
	"github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/pkg/logging"
	"github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/pkg/observability"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func preRun(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	err := viperInstance.Unmarshal(&cfgWrapper)
	if err != nil {
		return errors.Wrapf(err, "failed to initialize configuration")
	}

	// Translate cfgWrapper to config.Config
	err = config.AdaptConfigWrapperToConfig(&cfgWrapper, &cfg)
	if err != nil {
		return errors.Wrapf(err, "failed to process configuration")
	}

	logLevel, err := parseLogLevel(cfg.LogLevel)
	if err != nil {
		logger.Warn("Invalid log level was specified. Defaulting to Debug level.",
			"specified_level", cfg.LogLevel,
			"error", err)
		logLevel = slog.LevelDebug
	}

	slogOpts := &slog.HandlerOptions{
		AddSource: false,
		Level:     logLevel,
	}
	var slogHandler slog.Handler

	newLogFile, err := os.OpenFile(wrapperLogPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return errors.Wrapf(err, "failed to open log file '%s", wrapperLogPath)
	}
	logFile = newLogFile

	writer := io.MultiWriter(os.Stdout, logFile)

	slogHandler = slog.NewJSONHandler(writer, slogOpts)

	observability.SetLogger(slogHandler)
	cfg.Observability.Name = internal.AppName()
	observabilityProviderInstance, err := observability.NewProvider(ctx, &cfg.Observability)
	if err != nil {
		return err
	}
	observabilityProvider = observabilityProviderInstance

	obs = observabilityProvider.Stack("log", "trace", "metric", slogHandler)

	gameLogger = logging.NewRealtime(slogHandler, &cfg.Observability)

	contextHandler := logging.NewContextHandler(obs.LogHandler)

	logger = slog.New(contextHandler)

	return nil
}
