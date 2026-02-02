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

func (runner *Runner) Run(parentCtx context.Context) error {
	// Guard parent ctx containing required context types
	if parentCtx == nil {
		runner.logger.Warn("parent context is nil; substituting background context")
		parentCtx = context.Background()
	}
	val := parentCtx.Value(constants.ContextKeyRunId)
	if val == nil {
		runner.logger.Error("context missing run ID")
		return errors.New("missing run ID in context")
	}
	runId, ok := val.(uuid.UUID)
	if !ok {
		runner.logger.Error("context run ID is wrong type")
		return errors.New("run ID in context not UUID")
	}

	cancelCtx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	ctx, span, err := runner.spanner.NewSpan(cancelCtx, runner.name, nil)
	if ctx == nil {
		runner.logger.WarnContext(ctx, "span returned nil context using parent context as fallback")
		ctx = cancelCtx
	}
	if err != nil {
		runner.logger.ErrorContext(ctx, "span setup failed", "err", err)
	}

	if span != nil {
		defer span.End()
	}

	defer span.End()

	runner.logger.DebugContext(ctx, "Starting a new run", "run-id", runId)

	ctx, span1, err := runner.spanner.NewSpan(cancelCtx, "init", nil)
	if ctx == nil {
		runner.logger.WarnContext(ctx, "span returned nil context using parent context as fallback")
		ctx = cancelCtx
	}
	if err != nil {
		runner.logger.ErrorContext(ctx, "span setup failed", "err", err)
	}
	err = runner.mgr.Init(ctx, runId)
	if span1 != nil {
		span1.End()
	}
	if err != nil {
		return errors.Wrapf(err, "Runner %s failed to initialize", runner.name)
	}

	runner.logger.DebugContext(ctx, "Executing the run")
	ctx, span2, err := runner.spanner.NewSpan(cancelCtx, runner.name, nil)
	if ctx == nil {
		runner.logger.WarnContext(ctx, "span returned nil context using parent context as fallback")
		ctx = cancelCtx
	}
	if err != nil {
		runner.logger.ErrorContext(ctx, "span setup failed", "err", err)
	}
	err = runner.mgr.Run(ctx, runId)

	if span2 != nil {
		span2.End()
	}
	if err != nil {
		return errors.Wrapf(err, "Runner %s failed to run", runner.name)
	}

	runner.logger.DebugContext(ctx, "Starting to clean up the run")

	ctx, span3, _ := runner.spanner.NewSpan(ctx, "close", nil)
	err = runner.mgr.Close(ctx)
	if span3 != nil {
		span3.End()
	}
	if err != nil {
		return errors.Wrapf(err, "Runner %s failed to close", runner.name)
	}

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
