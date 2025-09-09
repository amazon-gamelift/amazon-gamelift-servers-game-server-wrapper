/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package logging

import (
	"context"
	"log/slog"

	"github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/pkg/observability"
	"github.com/pkg/errors"
)

type Game interface {
	New(ctx context.Context, name, logDirectory string) (*BufferedLogger, error)
}

type game struct {
	handler slog.Handler
	cfg     *observability.Config
}

func (game *game) New(ctx context.Context, name, logDirectory string) (*BufferedLogger, error) {
	resource, err := observability.NewResource(name)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create resource '%s'", name)
	}
	provider, err := observability.NewLoggingProvider(ctx, resource, game.cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create game log provider")
	}

	otelLogger := provider.Logger(name)
	slogger := observability.Slogger(otelLogger, nil)

	logger, err := NewBufferedLogger(ctx, slogger, name, logDirectory)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create logger")
	}

	logger.SetOnClosed(provider.ForceFlush)

	return logger, nil
}

func NewRealtime(handler slog.Handler, cfg *observability.Config) Game {
	return &game{
		handler: handler,
		cfg:     cfg,
	}
}
