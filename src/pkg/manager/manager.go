/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manager

import (
	"context"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/game"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/hosting"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/observability"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/types/events"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"log/slog"
)

// Service defines the interface for managing the complete lifecycle of a game server.
type Service interface {
	Init(ctx context.Context, runId uuid.UUID) error
	Run(ctx context.Context, runId uuid.UUID) error
	Close(ctx context.Context) error
}

type Config struct {
}

type service struct {
	cfg      *Config
	harness  Harness
	hosting  hosting.Service
	logger   *slog.Logger
	spanner  observability.Spanner
	gameMeta *game.InitMeta
	initMeta *hosting.InitMeta
}

func (service *service) onHostingStart(ctx context.Context, h *events.HostingStart, end <-chan error) error {
	ctx, span, _ := service.spanner.NewSpan(ctx, "manager onHostingStart", nil)
	defer span.End()
	if err := service.harness.HostingStart(ctx, h, end); err != nil {
		return errors.Wrapf(err, "Failed to start hosting")
	}

	service.logger.DebugContext(ctx, "Manager onHostingStart ended", "event", h)
	return nil
}

func (service *service) onHostingTerminate(ctx context.Context, h *events.HostingTerminate) error {
	service.logger.DebugContext(ctx, "Manager onHostingTerminate started", "event", h)
	if err := service.harness.HostingTerminate(ctx, h); err != nil {
		return errors.Wrapf(err, "Failed to stop hosting")
	}
	return nil
}

func (service *service) onHealthCheck(ctx context.Context) events.GameStatus {
	return service.harness.HealthCheck(ctx)
}

// Init initializes both the hosting service and game harness.
//
// Parameters:
//   - ctx: Context for initialization operation
//   - runId: Unique identifier for this run
//
// Returns:
//   - error: If initialization of either component fails
func (service *service) Init(ctx context.Context, runId uuid.UUID) error {
	service.logger.DebugContext(ctx, "Initializing the hosting")
	initMeta, err := service.hosting.Init(ctx, &hosting.InitArgs{
		RunId: runId,
	})
	if err != nil {
		return errors.Wrap(err, "Failed to initialize the hosting")
	}
	service.initMeta = initMeta

	service.logger.DebugContext(ctx, "Initializing the game harness")
	meta, err := service.harness.Init(ctx, &game.InitArgs{
		RunId: runId,
	})
	if err != nil {
		return errors.Wrap(err, "Failed to initialize the game")
	}
	service.gameMeta = meta

	service.hosting.SetOnHostingTerminate(service.onHostingTerminate)
	service.hosting.SetOnHostingStart(service.onHostingStart)
	service.hosting.SetOnHealthCheck(service.onHealthCheck)

	return nil
}

// Run executes the service loop, managing the game and hosting components.
//
// Parameters:
//   - ctx: Context for running operation
//   - runId: Unique identifier for this run
//
// Returns:
//   - error: If either component encounters an error during execution
func (service *service) Run(ctx context.Context, runId uuid.UUID) error {
	service.logger.DebugContext(ctx, "Running manager")

	gameErrorChannel := make(chan error)
	hostingErrorChannel := make(chan error)

	go func() {
		gameErrorChannel <- service.harness.Run(ctx)
	}()

	go func() {
		hostingErrorChannel <- service.hosting.Run(ctx)
	}()

	select {
	case reason := <-ctx.Done():
		service.logger.DebugContext(ctx, "Manager context done", "reason", reason)
		return nil
	case err := <-gameErrorChannel:
		if err != nil {
			return errors.Wrapf(err, "Encountered game error")
		}
		return nil
	case err := <-hostingErrorChannel:
		if err != nil {
			return errors.Wrapf(err, "Encountered hosting error")
		}
		return nil
	}
}

// Close performs cleanup and closure of resources in the correct order.
//
// Parameters:
//   - ctx: Context for cleanup operation
//
// Returns:
//   - error: If cleanup of either component fails
func (service *service) Close(ctx context.Context) error {
	var err error
	// close the harness first to allow for logs finalised for the hosting provider to hoover up
	if harnessErr := service.harness.Close(ctx); harnessErr != nil {
		err = harnessErr
	}

	if hostingErr := service.hosting.Close(ctx); hostingErr != nil {
		if err == nil {
			err = hostingErr
		} else {
			err = errors.Wrapf(err, hostingErr.Error())
		}
	}

	return err
}

func New(cfg *Config, g game.Server, hosting hosting.Service, logger *slog.Logger, spanner observability.Spanner, harness Harness) *service {

	service := &service{
		harness: harness,
		hosting: hosting,
		logger:  logger,
		spanner: spanner,
		cfg:     cfg,
	}

	return service
}
