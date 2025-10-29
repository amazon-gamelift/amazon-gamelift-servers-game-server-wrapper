/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package sdk

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/constants"
	"github.com/amazon-gamelift/amazon-gamelift-servers-go-server-sdk/v5/server/log"
)

type logAdaptor struct {
	ctx    context.Context
	logger *slog.Logger
}

func (logAdaptor logAdaptor) Debugf(s string, a ...any) {
	msg := fmt.Sprintf(s, a...)
	logAdaptor.logger.DebugContext(logAdaptor.ctx, msg)
}

func (logAdaptor logAdaptor) Warnf(s string, a ...any) {
	msg := fmt.Sprintf(s, a...)
	logAdaptor.logger.WarnContext(logAdaptor.ctx, msg)
}

func (logAdaptor logAdaptor) Errorf(s string, a ...any) {
	msg := fmt.Sprintf(s, a...)
	logAdaptor.logger.ErrorContext(logAdaptor.ctx, msg)
}

func NewLogAdaptor(ctx context.Context, logger *slog.Logger) log.ILogger {
	a := &logAdaptor{
		logger: logger,
		ctx:    context.WithValue(ctx, string(constants.ContextKeySource), "GameLiftServerSDK"),
	}

	return a
}
