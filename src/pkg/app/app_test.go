/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package app

import (
	"errors"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/internal/mocks"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/constants"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/observability"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/runner"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"log/slog"
	"testing"
	"time"
)

type AppMock struct {
	logger  *slog.Logger
	spanner observability.Spanner
	Service *Service
	ctx     context.Context
	logs    *[]slog.Record
}

type MockSlogHandler struct {
	EnabledResponse   bool
	HandleError       error
	WithAttrsResponse *MockSlogHandler
	WithGroupResponse *MockSlogHandler
	Records           *[]slog.Record
}

func (mockHandler MockSlogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return mockHandler.EnabledResponse
}

func (mockHandler MockSlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return *mockHandler.WithAttrsResponse
}

func (mockHandler MockSlogHandler) WithGroup(name string) slog.Handler {
	return *mockHandler.WithGroupResponse
}

func (mockHandler MockSlogHandler) HasRecord(predicate func(record slog.Record) bool) bool {
	for _, record := range *mockHandler.Records {
		if predicate(record) {
			return true
		}
	}
	return false
}

func (mockHandler MockSlogHandler) DoesNotHaveRecord(predicate func(record slog.Record) bool) bool {
	for _, record := range *mockHandler.Records {
		if predicate(record) {
			return false
		}
	}
	return true
}

func (mockHandler MockSlogHandler) Handle(_ context.Context, r slog.Record) error {
	*mockHandler.Records = append(*mockHandler.Records, r)
	return mockHandler.HandleError
}

func createAppWithMocks(runner runner.Runner) (AppMock, *MockSlogHandler) {
	spannerMock := mocks.SpannerMock{}
	ctx, _ := context.WithTimeout(context.Background(), time.Hour*1)
	ctx = context.WithValue(ctx, constants.ContextKeyRunId, uuid.New())
	logs := []slog.Record{}
	mockHandler := MockSlogHandler{Records: &logs, EnabledResponse: true}
	logger := slog.New(mockHandler)
	return AppMock{
		logger:  logger,
		spanner: &spannerMock,
		Service: New(logger, &runner, &spannerMock),
		ctx:     ctx,
		logs:    &logs,
	}, &mockHandler
}

func TestRunHappyPath(t *testing.T) {
	// Arrange
	runnerMockHelper := runner.CreateRunnerMockHelper()
	appMock, mockSlogHandler := createAppWithMocks(*runnerMockHelper.Runner)

	// Act
	err := appMock.Service.Run(appMock.ctx)
	if err != nil {
		assert.Fail(t, "TestGetByIdRedisCacheHit Errored")
	}

	//Assert
	assert.Nil(t, err)
	assert.True(t, mockSlogHandler.HasRecord(func(record slog.Record) bool {
		return record.Message == "Starting game server wrapper application"
	}))
	assert.True(t, mockSlogHandler.HasRecord(func(record slog.Record) bool {
		return record.Message == "Initializing game server runner"
	}))
	assert.True(t, mockSlogHandler.HasRecord(func(record slog.Record) bool {
		return record.Message == "Game server runner completed successfully"
	}))
}

func TestRunnerRunFailed(t *testing.T) {
	// Arrange
	runnerMockHelper := runner.CreateRunnerMockHelper()
	runnerMockHelper.ManagerService.RunError = errors.New("test error")
	appMock, mockSlogHandler := createAppWithMocks(*runnerMockHelper.Runner)

	// Act
	err := appMock.Service.Run(appMock.ctx)

	//Assert
	assert.ErrorContains(t, err, "Game server execution failed")
	assert.ErrorContains(t, err, "test error")
	assert.True(t, mockSlogHandler.HasRecord(func(record slog.Record) bool {
		return record.Message == "Starting game server wrapper application"
	}))
	assert.True(t, mockSlogHandler.HasRecord(func(record slog.Record) bool {
		return record.Message == "Initializing game server runner"
	}))
	assert.True(t, mockSlogHandler.DoesNotHaveRecord(func(record slog.Record) bool {
		return record.Message == "Game server runner completed successfully"
	}))
}
