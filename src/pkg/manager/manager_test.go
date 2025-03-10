/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manager

import (
	"bytes"
	"errors"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/internal/mocks"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/game"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/hosting"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/observability"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/types/events"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"log/slog"
	"testing"
	"time"
)

type ManagerTestHelper struct {
	Logger                *slog.Logger
	LogBuffer             *bytes.Buffer
	Spanner               observability.Spanner
	Ctx                   context.Context
	ManagerService        *service
	GameService           *GameServiceMock
	Harness               *HarnessMock
	HostingService        *hosting.HostingServiceMock
	HostingStartEvent     *events.HostingStart
	HostingTerminateEvent *events.HostingTerminate
}

func CreateManagerTestHelper() ManagerTestHelper {
	logBuffer := bytes.Buffer{}
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	logger := slog.New(slog.NewTextHandler(&logBuffer, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	spannerMock := mocks.SpannerMock{}
	gameService := GameServiceMock{}
	hostingService := hosting.HostingServiceMock{
		Delay: time.Second,
	}
	harness := HarnessMock{
		Delay: time.Second,
	}

	managerService := New(&Config{}, &gameService, &hostingService, logger, &spannerMock, &harness)
	return ManagerTestHelper{
		Logger:         logger,
		LogBuffer:      &logBuffer,
		Spanner:        &spannerMock,
		Ctx:            ctx,
		ManagerService: managerService,
		GameService:    &gameService,
		Harness:        &harness,
		HostingService: &hostingService,
	}
}

func Test_Manager_onHostingStart_HappyPath(t *testing.T) {
	//arrange
	managerTestHelper := CreateManagerTestHelper()
	managerTestHelper.HostingStartEvent = &events.HostingStart{}

	//act
	var doneErrorChannel = make(chan error)
	err := managerTestHelper.ManagerService.onHostingStart(managerTestHelper.Ctx, managerTestHelper.HostingStartEvent, doneErrorChannel)

	//assert
	assert.Nil(t, err)
	assert.True(t, managerTestHelper.Harness.HostingStartCalled)
	logBuffer := managerTestHelper.LogBuffer.String()
	assert.Contains(t, logBuffer, "Manager onHostingStart ended")
}

func Test_Manager_onHostingStart_Harness_Error(t *testing.T) {
	//arrange
	managerTestHelper := CreateManagerTestHelper()
	managerTestHelper.Harness.HostingStartError = errors.New("Unit Test")

	//act
	var doneErrorChannel = make(chan error)
	err := managerTestHelper.ManagerService.onHostingStart(managerTestHelper.Ctx, managerTestHelper.HostingStartEvent, doneErrorChannel)

	//assert
	assert.Errorf(t, err, "Failed to start hosting")
	assert.True(t, managerTestHelper.Harness.HostingStartCalled)
	logBuffer := managerTestHelper.LogBuffer.String()
	assert.NotContains(t, logBuffer, "Manager onHostingStart ended")
}

func Test_Manager_onHostingTerminate_HappyPath(t *testing.T) {
	//arrange
	managerTestHelper := CreateManagerTestHelper()
	managerTestHelper.HostingTerminateEvent = &events.HostingTerminate{
		Reason: "Unit Test",
	}

	//act
	err := managerTestHelper.ManagerService.onHostingTerminate(managerTestHelper.Ctx, managerTestHelper.HostingTerminateEvent)

	//assert
	assert.Nil(t, err)
	assert.True(t, managerTestHelper.Harness.HostingTerminateCalled)
	logBuffer := managerTestHelper.LogBuffer.String()
	assert.Contains(t, logBuffer, "Manager onHostingTerminate started")
}

func Test_Manager_onHostingTerminate_Harness_Error(t *testing.T) {
	//arrange
	managerTestHelper := CreateManagerTestHelper()
	managerTestHelper.HostingTerminateEvent = &events.HostingTerminate{
		Reason: "Unit Test",
	}
	managerTestHelper.Harness.HostingTerminateError = errors.New("Unit Test")

	//act
	err := managerTestHelper.ManagerService.onHostingTerminate(managerTestHelper.Ctx, managerTestHelper.HostingTerminateEvent)

	//assert
	assert.Errorf(t, err, "Failed to stop hosting")
	assert.True(t, managerTestHelper.Harness.HostingTerminateCalled)
	logBuffer := managerTestHelper.LogBuffer.String()
	assert.Contains(t, logBuffer, "Manager onHostingTerminate started")
}

func Test_Manager_onHealthCheck_HappyPath(t *testing.T) {
	//arrange
	managerTestHelper := CreateManagerTestHelper()
	gameStatus := events.GameStatusWaiting
	managerTestHelper.Harness.GameStatus = &gameStatus

	//act
	gameStatusResponse := managerTestHelper.ManagerService.onHealthCheck(managerTestHelper.Ctx)

	//assert
	assert.Equal(t, gameStatus, gameStatusResponse)
	assert.True(t, managerTestHelper.Harness.HealthCheckCalled)
}

func Test_Manager_Init_HappyPath(t *testing.T) {
	//arrange
	managerTestHelper := CreateManagerTestHelper()
	runId := uuid.New()

	managerTestHelper.HostingService.InitMeta = &hosting.InitMeta{
		InstanceWorkingDirectory: "InstanceWorkingDirectory",
	}

	managerTestHelper.Harness.InitMeta = &game.InitMeta{}

	//act
	err := managerTestHelper.ManagerService.Init(managerTestHelper.Ctx, runId)

	//assert
	assert.Nil(t, err)
	assert.Equal(t, runId, managerTestHelper.HostingService.InitArgs.RunId)
	assert.Equal(t, runId, managerTestHelper.Harness.InitArgs.RunId)
	logBuffer := managerTestHelper.LogBuffer.String()
	assert.True(t, managerTestHelper.HostingService.InitCalled)
	assert.True(t, managerTestHelper.Harness.InitCalled)
	assert.True(t, managerTestHelper.HostingService.SetOnHostingTerminateCalled)
	assert.True(t, managerTestHelper.HostingService.SetOnHostingStartCalled)
	assert.True(t, managerTestHelper.HostingService.SetOnHealthCheckCalled)
	assert.Contains(t, logBuffer, "Initializing the hosting")
	assert.Contains(t, logBuffer, "Initializing the game harness")
}

func Test_Manager_Init_Hosting_Error(t *testing.T) {
	//arrange
	managerTestHelper := CreateManagerTestHelper()
	runId := uuid.New()

	managerTestHelper.HostingService.InitMeta = &hosting.InitMeta{
		InstanceWorkingDirectory: "InstanceWorkingDirectory",
	}
	managerTestHelper.HostingService.InitError = errors.New("Unit Test")

	//act
	err := managerTestHelper.ManagerService.Init(managerTestHelper.Ctx, runId)

	//assert
	assert.Errorf(t, err, "Failed to initialize the hosting")
	assert.Equal(t, runId, managerTestHelper.HostingService.InitArgs.RunId)
	logBuffer := managerTestHelper.LogBuffer.String()
	assert.True(t, managerTestHelper.HostingService.InitCalled)
	assert.False(t, managerTestHelper.Harness.InitCalled)
	assert.False(t, managerTestHelper.HostingService.SetOnHostingTerminateCalled)
	assert.False(t, managerTestHelper.HostingService.SetOnHostingStartCalled)
	assert.False(t, managerTestHelper.HostingService.SetOnHealthCheckCalled)
	assert.Contains(t, logBuffer, "Initializing the hosting")
	assert.NotContains(t, logBuffer, "Initializing the game harness")
}

func Test_Manager_Init_Harness_Error(t *testing.T) {
	//arrange
	managerTestHelper := CreateManagerTestHelper()
	runId := uuid.New()

	managerTestHelper.HostingService.InitMeta = &hosting.InitMeta{
		InstanceWorkingDirectory: "InstanceWorkingDirectory",
	}
	managerTestHelper.Harness.InitError = errors.New("Unit Test")

	//act
	err := managerTestHelper.ManagerService.Init(managerTestHelper.Ctx, runId)

	//assert
	assert.Errorf(t, err, "Failed to initialize the game")
	assert.Equal(t, runId, managerTestHelper.HostingService.InitArgs.RunId)
	logBuffer := managerTestHelper.LogBuffer.String()
	assert.True(t, managerTestHelper.HostingService.InitCalled)
	assert.True(t, managerTestHelper.Harness.InitCalled)
	assert.False(t, managerTestHelper.HostingService.SetOnHostingTerminateCalled)
	assert.False(t, managerTestHelper.HostingService.SetOnHostingStartCalled)
	assert.False(t, managerTestHelper.HostingService.SetOnHealthCheckCalled)
	assert.Contains(t, logBuffer, "Initializing the hosting")
	assert.Contains(t, logBuffer, "Initializing the game harness")
}

func Test_Manager_Run_HappyPath(t *testing.T) {
	//arrange
	managerTestHelper := CreateManagerTestHelper()
	runId := uuid.New()

	managerTestHelper.HostingService.InitMeta = &hosting.InitMeta{
		InstanceWorkingDirectory: "InstanceWorkingDirectory",
	}

	managerTestHelper.Harness.InitMeta = &game.InitMeta{}

	//act
	err := managerTestHelper.ManagerService.Run(managerTestHelper.Ctx, runId)

	//assert
	assert.Nil(t, err)
	assert.True(t, managerTestHelper.Harness.RunCalled)
	assert.True(t, managerTestHelper.HostingService.RunCalled)
}

func Test_Manager_Run_Harness_Error(t *testing.T) {
	//arrange
	managerTestHelper := CreateManagerTestHelper()
	runId := uuid.New()

	managerTestHelper.HostingService.Delay = time.Second * 3
	managerTestHelper.Harness.RunError = errors.New("Unit Test")

	//act
	err := managerTestHelper.ManagerService.Run(managerTestHelper.Ctx, runId)

	time.Sleep(time.Second * 5)

	//assert
	assert.Errorf(t, err, "Encountered game error")
	assert.True(t, managerTestHelper.Harness.RunCalled)
	assert.True(t, managerTestHelper.HostingService.RunCalled)
}

func Test_Manager_Run_Hosting_Error(t *testing.T) {
	//arrange
	managerTestHelper := CreateManagerTestHelper()
	runId := uuid.New()

	managerTestHelper.Harness.Delay = time.Second * 3
	managerTestHelper.HostingService.RunError = errors.New("Unit Test")

	//act
	err := managerTestHelper.ManagerService.Run(managerTestHelper.Ctx, runId)

	//assert
	assert.Errorf(t, err, "Encountered hosting error")
	assert.True(t, managerTestHelper.Harness.RunCalled)
	assert.True(t, managerTestHelper.HostingService.RunCalled)
}

func Test_Manager_Run_Context_Done(t *testing.T) {
	//arrange
	managerTestHelper := CreateManagerTestHelper()
	runId := uuid.New()

	managerTestHelper.HostingService.Delay = time.Second * 3
	managerTestHelper.Harness.Delay = time.Second * 3
	managerTestHelper.Harness.InitMeta = &game.InitMeta{}
	managerTestHelper.Ctx, _ = context.WithTimeout(managerTestHelper.Ctx, time.Second*1)

	//act
	err := managerTestHelper.ManagerService.Run(managerTestHelper.Ctx, runId)

	//assert
	assert.Nil(t, err)
	assert.True(t, managerTestHelper.Harness.RunCalled)
	assert.True(t, managerTestHelper.HostingService.RunCalled)
	logBuffer := managerTestHelper.LogBuffer.String()
	assert.Contains(t, logBuffer, "Manager context done")
}

func Test_Manager_Close_HappyPath(t *testing.T) {
	//arrange
	managerTestHelper := CreateManagerTestHelper()

	//act
	err := managerTestHelper.ManagerService.Close(managerTestHelper.Ctx)

	//assert
	assert.Nil(t, err)
	assert.True(t, managerTestHelper.Harness.CloseCalled)
	assert.True(t, managerTestHelper.HostingService.CloseCalled)
}

func Test_Manager_Close_Harness_Error(t *testing.T) {
	//arrange
	managerTestHelper := CreateManagerTestHelper()
	managerTestHelper.Harness.CloseError = errors.New("Harness Unit Test")

	//act
	err := managerTestHelper.ManagerService.Close(managerTestHelper.Ctx)

	//assert
	assert.Errorf(t, err, "Harness Unit Test")
	assert.True(t, managerTestHelper.Harness.CloseCalled)
	assert.True(t, managerTestHelper.HostingService.CloseCalled)
}

func Test_Manager_Close_Hosting_Error(t *testing.T) {
	//arrange
	managerTestHelper := CreateManagerTestHelper()
	managerTestHelper.HostingService.CloseError = errors.New("Hosting Unit Test")

	//act
	err := managerTestHelper.ManagerService.Close(managerTestHelper.Ctx)

	//assert
	assert.Errorf(t, err, "Hosting Unit Test")
	assert.True(t, managerTestHelper.Harness.CloseCalled)
	assert.True(t, managerTestHelper.HostingService.CloseCalled)
}
