/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runner

import (
	"context"
	"log/slog"

	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/constants"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/manager"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/observability"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type Services struct {
	Manager manager.Service
	Logger  *slog.Logger
	Spanner observability.Spanner
}

type Runner struct {
	mgr     manager.Service
	logger  *slog.Logger
	name    string
	spanner observability.Spanner
}

func (runner *Runner) Run(ctx context.Context) error {
	runId := ctx.Value(constants.ContextKeyRunId).(uuid.UUID)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ctx, span, _ := runner.spanner.NewSpan(ctx, runner.name, nil)
	defer span.End()

	runner.logger.DebugContext(ctx, "Starting a new run", "run-id", runId)

	ctx, span1, _ := runner.spanner.NewSpan(ctx, "init", nil)
	err := runner.mgr.Init(ctx, runId)
	span1.End()
	if err != nil {
		return errors.Wrapf(err, "Runner %s failed to initialize", runner.name)
	}

	runner.logger.DebugContext(ctx, "Executing the run")
	ctx, span2, _ := runner.spanner.NewSpan(ctx, runner.name, nil)
	if err := runner.mgr.Run(ctx, runId); err != nil {
		span2.End()
		return errors.Wrapf(err, "Runner %s failed to run", runner.name)
	}
	span2.End()

	runner.logger.DebugContext(ctx, "Starting to clean up the run")

	ctx, span3, _ := runner.spanner.NewSpan(ctx, "close", nil)
	if err := runner.mgr.Close(ctx); err != nil {
		span3.End()
		return errors.Wrapf(err, "Runner %s failed to close", runner.name)
	}
	span3.End()

	runner.logger.DebugContext(ctx, "Run cleaned up")
	return nil
}

func New(name string, manager manager.Service, logger *slog.Logger, spanner observability.Spanner) *Runner {
	return &Runner{
		mgr:     manager,
		logger:  logger,
		name:    name,
		spanner: spanner,
	}
}
