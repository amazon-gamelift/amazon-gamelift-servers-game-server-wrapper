/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runner

import (
	"bytes"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/internal/mocks"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/constants"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/observability"
	"github.com/google/uuid"
	"golang.org/x/net/context"
	"log/slog"
	"time"
)

type RunnerMockHelper struct {
	Logger         *slog.Logger
	LogBuffer      *bytes.Buffer
	Spanner        observability.Spanner
	Ctx            context.Context
	Runner         *Runner
	ManagerService *ManagerServiceMock
	RunnerService  *Services
	Condition      *bool
}

type ManagerServiceMock struct {
	InitError   error
	RunError    error
	CloseError  error
	InitCalled  bool
	RunCalled   bool
	CloseCalled bool
	InitCount   int
	RunCount    int
	CloseCount  int
	InitRunId   uuid.UUID
	RunRunId    uuid.UUID
}

func (managerServiceMock *ManagerServiceMock) Init(ctx context.Context, runId uuid.UUID) error {
	managerServiceMock.InitCalled = true
	managerServiceMock.InitCount++
	managerServiceMock.InitRunId = runId
	return managerServiceMock.InitError
}

func (managerServiceMock *ManagerServiceMock) Run(ctx context.Context, runId uuid.UUID) error {
	managerServiceMock.RunCalled = true
	managerServiceMock.RunCount++
	managerServiceMock.RunRunId = runId
	return managerServiceMock.RunError
}

func (managerServiceMock *ManagerServiceMock) Close(ctx context.Context) error {
	managerServiceMock.CloseCalled = true
	managerServiceMock.CloseCount++
	return managerServiceMock.CloseError
}

func CreateRunnerMockHelper() RunnerMockHelper {
	logBuffer := bytes.Buffer{}
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	ctx = context.WithValue(ctx, constants.ContextKeyRunId, uuid.New())
	logger := slog.New(slog.NewTextHandler(&logBuffer, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	spannerMock := mocks.SpannerMock{}
	managerServiceMock := ManagerServiceMock{}

	runner := New("UnitTest", &managerServiceMock, logger, &spannerMock)
	return RunnerMockHelper{
		Logger:         logger,
		LogBuffer:      &logBuffer,
		Spanner:        &spannerMock,
		Ctx:            ctx,
		Runner:         runner,
		ManagerService: &managerServiceMock,
	}
}
