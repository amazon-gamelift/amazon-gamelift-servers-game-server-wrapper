/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package multiplexgame

import (
	"bytes"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/internal/config"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/internal/mocks"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/game"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/logging"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/observability"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/process"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/types/events"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"log/slog"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

type MultiPlexGameMock struct {
	logger        *slog.Logger
	spanner       observability.Spanner
	logBuffer     *bytes.Buffer
	multiplexGame *MultiplexGame
	ctx           context.Context
	cfg           *config.Config
}

type MockSessionLoggerFactory struct {
	logger *slog.Logger
}

func (mockSessionLoggerFactory *MockSessionLoggerFactory) New(ctx context.Context, name string, logDirectory string) (*logging.BufferedLogger, error) {
	bufferedLogger, err := logging.NewBufferedLogger(ctx, mockSessionLoggerFactory.logger, name, logDirectory)
	if err != nil {
		return nil, err
	}
	return bufferedLogger, nil
}

func createMultiPlexGameWithMocks(cfg config.Config) MultiPlexGameMock {
	logBuffer := bytes.Buffer{}
	spannerMock := mocks.SpannerMock{}
	ctx, _ := context.WithTimeout(context.Background(), time.Hour*1)
	logger := slog.New(slog.NewTextHandler(&logBuffer, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	mockSessionLoggerFactory := MockSessionLoggerFactory{
		logger: logger,
	}
	multiplexGame, _ := New(cfg, logger, &mockSessionLoggerFactory, &spannerMock)
	return MultiPlexGameMock{
		logger:        logger,
		spanner:       &spannerMock,
		logBuffer:     &logBuffer,
		multiplexGame: multiplexGame,
		ctx:           ctx,
		cfg:           &cfg,
	}
}

func TestHealthCheckHappyPathNoProcess(t *testing.T) {
	// Arrange
	cfg := config.Config{}
	multiPlexGameMock := createMultiPlexGameWithMocks(cfg)
	multiPlexGameMock.multiplexGame.status = events.GameStatusWaiting

	// Act
	gameStatus := multiPlexGameMock.multiplexGame.HealthCheck(multiPlexGameMock.ctx)

	//Assert
	assert.Equal(t, events.GameStatusWaiting, gameStatus)
}

func TestHealthCheckHappyPathWithProcessExited(t *testing.T) {
	// Arrange
	cfg := config.Config{}
	multiPlexGameMock := createMultiPlexGameWithMocks(cfg)

	multiPlexGameMock.multiplexGame.proc = &mocks.ProcessMock{
		StateResponse: &process.State{
			Exited: true,
		},
	}

	// Act
	gameStatus := multiPlexGameMock.multiplexGame.HealthCheck(multiPlexGameMock.ctx)

	//Assert
	assert.Equal(t, events.GameStatusFinished, gameStatus)
}

func TestRunHappyPath(t *testing.T) {
	workingDir, err := filepath.Abs("../../internal/mocks")
	if err != nil {
		assert.Fail(t, "error getting mock working directory")
	}

	exePath := "./testrun.sh"
	switch runtime.GOOS {
	case "windows":
		exePath = "./testrun.bat"
	}

	cfg := config.Config{
		Ports: config.Ports{
			GamePort: 12345,
		},
		BuildDetail: config.BuildDetail{
			WorkingDir:      workingDir,
			RelativeExePath: exePath,
		},
	}

	multiPlexGameMock := createMultiPlexGameWithMocks(cfg)
	multiPlexGameMock.multiplexGame.status = events.GameStatusWaiting

	startArgs := game.StartArgs{
		&events.HostingStart{
			LogDirectory: workingDir,
		},
	}

	// Act
	err = multiPlexGameMock.multiplexGame.Run(multiPlexGameMock.ctx, &startArgs)

	//Assert
	assert.Nil(t, err)
	logString := multiPlexGameMock.logBuffer.String()
	assert.Contains(t, logString, "Starting multiplex game")
	assert.Contains(t, logString, "Generating command line arguments")
	assert.Contains(t, logString, "Creating log files")
	assert.Contains(t, logString, "Calling process run")
	assert.Contains(t, logString, "Waiting on process result")
	assert.Contains(t, logString, "Process run finished")
	assert.Contains(t, logString, "Process result received")
}

func TestInitHappyPath(t *testing.T) {
	// Arrange
	cfg := config.Config{}
	multiPlexGameMock := createMultiPlexGameWithMocks(cfg)

	initArgs := game.InitArgs{
		RunId: uuid.New(),
	}

	// Act
	initMeta, err := multiPlexGameMock.multiplexGame.Init(multiPlexGameMock.ctx, &initArgs)

	//Assert
	assert.Nil(t, err)
	assert.NotNil(t, initMeta)
	logString := multiPlexGameMock.logBuffer.String()
	assert.Contains(t, logString, "Starting multiplex game initialization")
	assert.Contains(t, logString, "Multiplex game initialized")
	assert.Equal(t, events.GameStatusWaiting, multiPlexGameMock.multiplexGame.status)
}

func TestStopHappyPath(t *testing.T) {
	// Arrange
	cfg := config.Config{}
	multiPlexGameMock := createMultiPlexGameWithMocks(cfg)

	// Act
	err := multiPlexGameMock.multiplexGame.Stop(multiPlexGameMock.ctx)

	//Assert
	assert.Nil(t, err)
	logString := multiPlexGameMock.logBuffer.String()
	assert.Contains(t, logString, "Initiating game server shutdown")
}
