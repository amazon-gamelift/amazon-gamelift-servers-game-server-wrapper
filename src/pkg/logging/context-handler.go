/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package logging

import (
	"context"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/constants"
	"log/slog"
)

type ContextHandler struct {
	next slog.Handler
	keys []string
}

func NewContextHandler(next slog.Handler) *ContextHandler {
	c := &ContextHandler{
		next: next,
		keys: []string{
			string(constants.ContextKeySource),
			string(constants.ContextKeyPackageName),
			string(constants.ContextKeyVersion),
		},
	}

	return c
}

func (contextHandler *ContextHandler) WithKeys(keys ...string) *ContextHandler {
	contextHandler.keys = append(contextHandler.keys, keys...)
	return contextHandler
}

func (contextHandler *ContextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return contextHandler.next.Enabled(ctx, level)
}

func (contextHandler *ContextHandler) Handle(ctx context.Context, record slog.Record) error {
	for _, key := range contextHandler.keys {
		val := ctx.Value(key)
		if val != nil {
			record.AddAttrs(slog.Any(key, val))
		}
	}

	return contextHandler.next.Handle(ctx, record)
}

func (contextHandler *ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	contextHandler.next = contextHandler.next.WithAttrs(attrs)
	return contextHandler
}

func (contextHandler *ContextHandler) WithGroup(name string) slog.Handler {
	contextHandler.next = contextHandler.next.WithGroup(name)
	return contextHandler
}
