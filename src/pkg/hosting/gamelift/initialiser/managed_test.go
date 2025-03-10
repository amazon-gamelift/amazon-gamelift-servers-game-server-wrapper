/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package initialiser

import (
	"bytes"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/internal/mocks"
	"golang.org/x/net/context"
	"log/slog"
	"time"
)

type ManagedMockHelper struct {
	logger            *slog.Logger
	logBuffer         *bytes.Buffer
	gameLiftSdk       *mocks.GameLiftSdkMock
	ctx               context.Context
	managed           Service
}

func createManagedMockHelper() ManagedMockHelper {
	logBuffer := bytes.Buffer{}
	ctx, _ := context.WithTimeout(context.Background(), time.Minute*5)
	logger := slog.New(slog.NewTextHandler(&logBuffer, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	gameLiftSdkMock := mocks.GameLiftSdkMock{}
	managed := newManaged(&gameLiftSdkMock, logger)
	return ManagedMockHelper{
		logger:            logger,
		logBuffer:         &logBuffer,
		gameLiftSdk:       &gameLiftSdkMock,
		ctx:               ctx,
		managed:           managed,
	}
}


