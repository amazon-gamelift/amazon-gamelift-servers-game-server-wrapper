/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package app

import (
	"context"
	"log/slog"
	"runtime/debug"
	"time"

	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/internal"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/observability"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/runner"
	"github.com/pkg/errors"
)

// Service represents the application service that manages the game server lifecycle
type Service struct {
	logger  *slog.Logger          // Handles logging operations
	runner  *runner.Runner        // Manages game server execution
	spanner observability.Spanner // Provides tracing and monitoring
}

// Run executes the application logic with panic recovery and observability.
//
// Parameters:
//   - ctx: Context for the application lifecycle
//
// Returns:
//   - error: Any error that occurs during execution
func (service *Service) Run(ctx context.Context) error {
	defer func() {
		if err := recover(); err != nil {
			service.logger.ErrorContext(ctx, "App panic detected", "err", err, "stack", string(debug.Stack()))
		}
	}()

	service.logger.InfoContext(ctx, "Starting game server wrapper application", "version", internal.Version())

	service.logger.DebugContext(ctx, "Initializing game server runner")
	ctx, span, _ := service.spanner.NewSpan(ctx, "runner run", nil)
	if err := service.runner.Run(ctx); err != nil {
		return errors.Wrapf(err, "Game server execution failed")
	}
	span.End()

	service.logger.DebugContext(ctx, "Game server runner completed successfully")

	<-time.After(time.Second)

	return nil
}

// New creates a new Service instance with the provided parameters.
//
// Parameters:
//   - l: Logger for operational logging
//   - rnr: Runner for game server execution
//   - sp: Spanner for observability
//
// Returns:
//   - *Service: New service instance
func New(l *slog.Logger, rnr *runner.Runner, sp observability.Spanner) *Service {
	s := &Service{
		logger:  l,
		runner:  rnr,
		spanner: sp,
	}
	return s
}
