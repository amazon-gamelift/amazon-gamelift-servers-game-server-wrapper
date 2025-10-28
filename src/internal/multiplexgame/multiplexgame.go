/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package multiplexgame

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/internal/config"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/internal/multiplexgame/args"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/game"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/logging"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/observability"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/process"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/types/events"
)

// New creates a new MultiplexGame instance with the provided configuration and dependencies.
//
// Parameters:
//   - cfg: Configuration for the game server
//   - logger: Main logger for the game server
//   - sessionLoggerFactory: Factory for creating session-specific loggers
//   - spanner: Tracing and monitoring provider
//
// Returns:
//   - *MultiplexGame: New game server instance
//   - error: Any error during initialization
func New(cfg config.Config, logger *slog.Logger, sessionLoggerFactory SessionLoggerFactory, spanner observability.Spanner) (*MultiplexGame, error) {
	if logger == nil {

		return nil, errors.New("multiplex game initialization failed: logger not provided")
	}
	if sessionLoggerFactory == nil {
		return nil, errors.New("multiplex game initialization failed: session logger factory not provided")
	}
	if spanner == nil {
		return nil, errors.New("multiplex game initialization failed: spanner not provided")
	}
	multiplexGame := MultiplexGame{
		cfg:                  cfg,
		logger:               logger,
		sessionLoggerFactory: sessionLoggerFactory,
		spanner:              spanner,
	}
	return &multiplexGame, nil
}

// MultiplexGame represents a game server instance that can manage multiple game processes.
// It handles process lifecycle, logging, monitoring, and status management.
type MultiplexGame struct {
	cfg                  config.Config
	logger               *slog.Logger
	sessionLoggerFactory logging.Game
	spanner              observability.Spanner
	stdout, stderr       *logging.BufferedLogger
	proc                 process.Process
	status               events.GameStatus
	cancel               func()
}

// SessionLoggerFactory defines the interface for creating session-specific loggers.
type SessionLoggerFactory interface {
	New(ctx context.Context, name string, logDirectory string) (*logging.BufferedLogger, error)
}

// HealthCheck performs a health check on the game server and returns its current status.
//
// Parameters:
//   - ctx: Context for the health check operation
//
// Returns:
//   - events.GameStatus: Current status of the game server
func (multiplexGame *MultiplexGame) HealthCheck(ctx context.Context) events.GameStatus {
	var pid *int
	if multiplexGame.proc != nil {
		procState := multiplexGame.proc.State()

		if procState != nil {
			if procState.Pid != 0 {
				pid = &procState.Pid
			}

			if procState.Exited {
				multiplexGame.status = events.GameStatusFinished
			}
		}
	}
	_ = pid
	return multiplexGame.status
}

// Run starts the game server process with the provided arguments.
//
// Parameters:
//   - ctx: Context for the run operation
//   - startArgs: Arguments for starting the game server
//
// Returns:
//   - error: Any error during server execution
func (multiplexGame *MultiplexGame) Run(ctx context.Context, startArgs *game.StartArgs) error {
	ctx, span, _ := multiplexGame.spanner.NewSpan(ctx, "run", nil)
	defer span.End()

	if multiplexGame.cfg.Ports.GamePort == 0 {
		return errors.New("game server initialization failed: invalid game port: 0")
	}

	multiplexGame.logger.InfoContext(ctx, "Running multiplex build", "port", multiplexGame.cfg.Ports.GamePort)

	build := multiplexGame.cfg.BuildDetail

	err := multiplexGame.initProcess(ctx, build)
	if err != nil {
		multiplexGame.logger.ErrorContext(ctx, "Game process initialization failed",
			"error", err,
			"buildPath", build.RelativeExePath,
			"workingDir", build.WorkingDir)
		return fmt.Errorf("failed to initialize game process: %w", err)
	}

	multiplexGame.logger.InfoContext(ctx, "Starting multiplex game", "arguments", startArgs)

	ctx, cancel := context.WithCancel(ctx)
	multiplexGame.cancel = cancel

	gsPidChan := make(chan int)

	defer cancel()

	multiplexGame.logger.DebugContext(ctx, "Generating command line arguments")
	argGenerator, err := args.New(&args.Config{
		BuildDetail: build,
	})
	if err != nil {
		multiplexGame.logger.Error("Failed to create argument generator: ", "build", build, "error", err)
		return fmt.Errorf("failed to create argument generator: %w", err)
	}
	processArgs, err := argGenerator.Get(startArgs)
	if err != nil {
		multiplexGame.logger.Error("Failed to generate process arguments: ", "error", err)
		return fmt.Errorf("failed to generate process arguments: %w", err)
	}
	multiplexGame.logger.DebugContext(ctx, "cli args: ", "args", processArgs)

	multiplexGame.logger.DebugContext(ctx, "Creating log files")
	if err := multiplexGame.createLogStreams(ctx, startArgs.LogDirectory); err != nil {
		multiplexGame.logger.Error("failed to create log streams for stdout and stderr", "error", err)
		return fmt.Errorf("failed to create log streams: %w", err)
	}

	e := make(chan error)
	go func() {
		multiplexGame.status = events.GameStatusRunning
		multiplexGame.logger.DebugContext(ctx, "Calling process run")

		res, err := multiplexGame.proc.Run(ctx, &process.Args{
			CliArgs: processArgs,
			Stdout:  io.MultiWriter(os.Stdout, multiplexGame.stdout),
			Stderr:  io.MultiWriter(os.Stderr, multiplexGame.stderr),
		}, gsPidChan)

		multiplexGame.logger.DebugContext(ctx, "Process run finished", "result", res)

		if err != nil {
			multiplexGame.status = events.GameStatusErrored
			multiplexGame.logger.Error("Game process execution failed: ", "error", err)
			e <- fmt.Errorf("game process failure: %w", err)
			return
		}

		multiplexGame.status = events.GameStatusFinished

		e <- nil
	}()

	multiplexGame.logger.DebugContext(ctx, "Waiting on process result")
	err = <-e
	multiplexGame.logger.DebugContext(ctx, "Process result received", "error", err)

	return err
}

// Init initializes the game server instance and prepares it for operation.
// This method sets up the initial state and returns metadata about the initialization.
//
// Parameters:
//   - ctx: Context for the initialization operation
//   - args: Initialization arguments containing setup parameters
//
// Returns:
//   - *game.InitMeta: Metadata about the initialized game server
//   - error: Any error during initialization
func (multiplexGame *MultiplexGame) Init(ctx context.Context, args *game.InitArgs) (*game.InitMeta, error) {
	multiplexGame.logger.DebugContext(ctx, "Starting multiplex game initialization", "args", args)
	multiplexGame.status = events.GameStatusWaiting
	meta := &game.InitMeta{}

	multiplexGame.logger.InfoContext(ctx, "Multiplex game initialized")
	return meta, nil
}

func (multiplexGame *MultiplexGame) initProcess(ctx context.Context, build config.BuildDetail) error {
	wd, err := os.Stat(build.WorkingDir)
	if err != nil {
		return fmt.Errorf("failed to access working directory %s: %w", build.WorkingDir, err)
	}

	if !wd.IsDir() {
		return fmt.Errorf("path %s is not a directory", build.WorkingDir)
	}

	multiplexGame.logger.DebugContext(ctx, "Working directory validated successfully", "dir", build.WorkingDir)

	envMap := make(map[string]string)
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) == 2 {
			envMap[pair[0]] = pair[1]
		}
	}
	multiplexGame.logger.DebugContext(ctx, "Passing wrapper's environment variables to game process", "envVarsCount", len(envMap))

	procCfg := &process.Config{
		EnvVars:          envMap,
		WorkingDirectory: build.WorkingDir,
		ExeName:          build.RelativeExePath,
	}
	multiplexGame.proc = process.New(procCfg, multiplexGame.logger)

	multiplexGame.logger.DebugContext(ctx, "Initializing game process with configuration", "procCfg", procCfg)
	if err := multiplexGame.proc.Init(ctx); err != nil {
		return fmt.Errorf("failed to initialize game process: %w", err)
	}
	return nil
}

// Stop gracefully stops the game server and performs cleanup operations.
// It handles the shutdown of all components and ensures proper resource cleanup.
//
// Parameters:
//   - ctx: Context for the stop operation
//
// Returns:
//   - error: Any error during shutdown
func (multiplexGame *MultiplexGame) Stop(ctx context.Context) error {
	multiplexGame.logger.InfoContext(ctx, "Initiating game server shutdown")
	if multiplexGame.cancel != nil {
		multiplexGame.logger.DebugContext(ctx, "Canceling game server context")
		multiplexGame.cancel()
	}

	if multiplexGame.stdout != nil {
		multiplexGame.logger.DebugContext(ctx, "Closing stdout log stream")
		if err := multiplexGame.stdout.Close(); err != nil {
			multiplexGame.logger.ErrorContext(ctx, "failed to close stdout log file", "err", err)
		}
	}

	if multiplexGame.stderr != nil {
		multiplexGame.logger.DebugContext(ctx, "closing stderr log stream")
		if err := multiplexGame.stderr.Close(); err != nil {
			multiplexGame.logger.ErrorContext(ctx, "failed to close stderr log file", "err", err)
		}
	}

	multiplexGame.logger.InfoContext(ctx, "Game server shutdown completed")

	return nil
}

func (multiplexGame *MultiplexGame) createLogStreams(ctx context.Context, logDirectory string) error {
	stdout, err := multiplexGame.sessionLoggerFactory.New(ctx, "game-stdout.log", logDirectory)
	if err != nil {
		return err
	}
	stderr, err := multiplexGame.sessionLoggerFactory.New(ctx, "game-stderr.log", logDirectory)
	if err != nil {
		return err
	}

	multiplexGame.stdout, multiplexGame.stderr = stdout, stderr

	return nil
}
