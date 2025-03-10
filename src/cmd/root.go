/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cmd

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"log/slog"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/internal"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/internal/config"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/internal/services"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/app"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/constants"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/logging"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/observability"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	appDir  string
	cfgFile string

	// cfgWrapper is a wrapper for the config.Config struct
	// that is used to unmarshal the config.yaml file and set up the config struct
	// it is not used for anything else
	cfg        = config.Config{}
	cfgWrapper = config.ConfigWrapper{}

	logger         *slog.Logger
	viperInstance  = viper.New()
	wrapperLogPath string
	logFile        *os.File

	gameLogger logging.Game

	obs                   *observability.Observability
	observabilityProvider *observability.Provider

	rootCmd = &cobra.Command{
		PersistentPreRunE: preRun,
		RunE:              run,
		SilenceErrors:     false,
		SilenceUsage:      false,
	}
)

func initConfig() {
	var cfgFilePath string
	if cfgFile != "" {
		viperInstance.SetConfigFile(cfgFile)
		cfgFilePath = cfgFile
	} else {
		viperInstance.AddConfigPath(appDir)
		viperInstance.SetConfigName("config")
		viperInstance.SetConfigType("yaml")
		cfgFilePath = filepath.Join(appDir, "config.yaml")
	}

	err := viperInstance.ReadInConfig()
	if err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			cobra.CheckErr(err)
			logger.Error("failed to parse config", "err", err, "path", cfgFilePath)
			os.Exit(1)
		} else {
			logger.Error("config file not found", "err", err, "path", cfgFilePath)
			os.Exit(1)
		}
	}

	viperInstance.AutomaticEnv()
	viperInstance.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

func init() {
	appInit()
	bindFlags()
}

func run(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	logger.Info("Initialized working directory", "directory", appDir)

	ctx, span, err := obs.Spanner.NewSpan(ctx, "run", nil)
	if err != nil {
		return err
	}

	defer func() {
		if err := recover(); err != nil {
			logger.Error("app panic", "err", err, "stack", string(debug.Stack()))
		}
		span.End()
	}()

	logger.DebugContext(ctx, fmt.Sprintf("starting %s", internal.AppName()), "args", args)

	svcs, err := services.Default(ctx, &cfg, logger, obs, gameLogger)
	if err != nil {
		return errors.Wrapf(err, "failed to construct services")
	}

	wrp := app.New(svcs.Logger, svcs.Runner, obs.Spanner)

	ctx, cancel := addSyscallInterrupt(ctx)
	defer cancel()

	return wrp.Run(ctx)
}

func Execute() {
	var err error
	exit1 := false
	ctx := context.WithValue(context.Background(), string(constants.ContextKeySource), internal.AppName())
	ctx = context.WithValue(ctx, string(constants.ContextKeyVersion), internal.SemVer())
	ctx = context.WithValue(ctx, string(constants.ContextKeyAppDir), appDir)

	ctx, err = setupLogging(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "failed to setup logging", "err", err)
		exit1 = true
	}

	err = rootCmd.ExecuteContext(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "failed to execute the root command", "err", err)
		exit1 = true
	}

	if observabilityProvider != nil {
		if err := observabilityProvider.Close(); err != nil {
			logger.ErrorContext(ctx, "failed to close observability provider", "err", err)
		}
	}

	_ = logFile.Close()

	if exit1 {
		os.Exit(1)
	}
}

func setupLogging(ctx context.Context) (context.Context, error) {
	runId := uuid.New()
	ctx = context.WithValue(ctx, constants.ContextKeyRunId, runId)

	runLogDir := filepath.Join(appDir, "logs", "run_"+runId.String())
	err := os.MkdirAll(runLogDir, 0755)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create log directory '%s'", runLogDir)
	}
	ctx = context.WithValue(ctx, constants.ContextKeyRunLogDir, runLogDir)

	logFileName := fmt.Sprintf("%s.log", internal.AppName())
	wrapperLogPath = filepath.Join(runLogDir, logFileName)
	ctx = context.WithValue(ctx, constants.ContextKeyWrapperLogPath, wrapperLogPath)

	return ctx, nil
}
