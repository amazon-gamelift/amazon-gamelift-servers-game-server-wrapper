/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manager

import (
	"bytes"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/internal/mocks"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/game"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/observability"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/types/events"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

type HarnessTestHelper struct {
	Logger                *slog.Logger
	LogBuffer             *bytes.Buffer
	Spanner               observability.Spanner
	Ctx                   context.Context
	Harness               *harness
	GameService           *GameServiceMock
	HostingStartEvent     *events.HostingStart
	HostingTerminateEvent *events.HostingTerminate
}

func CreateHarnessTestHelper(duration time.Duration) HarnessTestHelper {
	logBuffer := bytes.Buffer{}
	ctx, _ := context.WithTimeout(context.Background(), duration)
	logger := slog.New(slog.NewTextHandler(&logBuffer, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	spannerMock := mocks.SpannerMock{}
	gameService := &GameServiceMock{}

	harness := NewHarness(gameService, logger, &spannerMock)
	return HarnessTestHelper{
		Logger:      logger,
		LogBuffer:   &logBuffer,
		Spanner:     &spannerMock,
		Ctx:         ctx,
		Harness:     harness,
		GameService: gameService,
	}
}

func Test_Harness_HostingStart_HappyPath_Channel_Listener(t *testing.T) {
	//arrange
	harnessTestHelper := CreateHarnessTestHelper(time.Second * 5)
	harnessTestHelper.HostingStartEvent = &events.HostingStart{}

	var hostingStartEvent *events.HostingStart

	//act
	go func() {
		select {
		case hostingStartEvent = <-harnessTestHelper.Harness.hostingStart:
			return
		case <-harnessTestHelper.Ctx.Done():
			return
		}
	}()

	err := harnessTestHelper.Harness.HostingStart(harnessTestHelper.Ctx, harnessTestHelper.HostingStartEvent, nil)

	//assert
	assert.Nil(t, err)
	assert.NotNil(t, harnessTestHelper.Harness.hostingStart)
	assert.Same(t, harnessTestHelper.HostingStartEvent, hostingStartEvent)
}

func Test_Harness_HostingStart_HappyPath_Channel_Listener_End_Channel(t *testing.T) {
	//arrange
	harnessTestHelper := CreateHarnessTestHelper(time.Second * 5)
	harnessTestHelper.HostingStartEvent = &events.HostingStart{}

	//doneErrorChannel := make(chan error)
	var hostingStartEvent *events.HostingStart

	//act
	go func() {
		select {
		case hostingStartEvent = <-harnessTestHelper.Harness.hostingStart:
			return
		case <-harnessTestHelper.Ctx.Done():
			return
		}
	}()

	errorChannel := make(chan error)

	var err error
	go func() {
		err = harnessTestHelper.Harness.HostingStart(harnessTestHelper.Ctx, harnessTestHelper.HostingStartEvent, errorChannel)
	}()

	errorChannel <- errors.New("Unit Test")

	time.Sleep(time.Second * 2)

	//<-errorChannel

	//assert
	assert.NotNil(t, err)
	assert.NotNil(t, harnessTestHelper.Harness.hostingStart)
	assert.Same(t, harnessTestHelper.HostingStartEvent, hostingStartEvent)
}

func Test_Harness_HostingTerminate_HappyPath(t *testing.T) {
	//arrange
	harnessTestHelper := CreateHarnessTestHelper(time.Second * 5)
	harnessTestHelper.HostingTerminateEvent = &events.HostingTerminate{
		Reason: "Unit Test",
	}

	var hostingTerminateEvent *events.HostingTerminate

	//act
	go func() {
		select {
		case hostingTerminateEvent = <-harnessTestHelper.Harness.hostingTerminate:
			return
		case <-harnessTestHelper.Ctx.Done():

			return
		}
	}()

	err := harnessTestHelper.Harness.HostingTerminate(harnessTestHelper.Ctx, harnessTestHelper.HostingTerminateEvent)

	//assert
	assert.Nil(t, err)
	assert.NotNil(t, harnessTestHelper.Harness.hostingTerminate)
	assert.Same(t, harnessTestHelper.HostingTerminateEvent, hostingTerminateEvent)
}

func Test_Harness_HealthCheck_HappyPath(t *testing.T) {
	//arrange
	harnessTestHelper := CreateHarnessTestHelper(time.Second * 5)
	harnessTestHelper.HostingTerminateEvent = &events.HostingTerminate{
		Reason: "Unit Test",
	}

	gameStatusVariable := events.GameStatusWaiting

	harnessTestHelper.GameService.GameStatus = &gameStatusVariable

	//act
	gameStatus := harnessTestHelper.Harness.HealthCheck(harnessTestHelper.Ctx)

	//assert
	assert.True(t, harnessTestHelper.GameService.HealthCheckCalled)
	assert.Equal(t, 1, harnessTestHelper.GameService.HealthCheckCount)
	assert.Equal(t, gameStatusVariable, gameStatus)
}

func Test_Harness_Init_HappyPath(t *testing.T) {
	//arrange
	harnessTestHelper := CreateHarnessTestHelper(time.Second * 5)

	gameInitMeta := game.InitMeta{}
	harnessTestHelper.GameService.InitMeta = &gameInitMeta

	gameInitArgs := game.InitArgs{
		RunId: uuid.New(),
	}

	//act
	gameMetaData, err := harnessTestHelper.Harness.Init(harnessTestHelper.Ctx, &gameInitArgs)

	//assert
	assert.Nil(t, err)
	assert.True(t, harnessTestHelper.GameService.InitCalled)
	assert.Equal(t, 1, harnessTestHelper.GameService.InitCount)
	assert.Equal(t, &gameInitMeta, gameMetaData)
}

func Test_Harness_Run_HappyPath_Context_Done(t *testing.T) {
	//arrange
	harnessTestHelper := CreateHarnessTestHelper(time.Second * 5)

	//act
	err := harnessTestHelper.Harness.Run(harnessTestHelper.Ctx)

	//assert
	assert.Nil(t, err)
	logBuffer := harnessTestHelper.LogBuffer.String()
	assert.Contains(t, logBuffer, "Game server harness starting")
	assert.NotContains(t, logBuffer, "Received hosting start event")
	assert.NotContains(t, logBuffer, "Received hosting terminate event")
	assert.Contains(t, logBuffer, "Game server harness context done")
}

func Test_Harness_Run_HappyPath_Start_Received(t *testing.T) {
	//arrange
	harnessTestHelper := CreateHarnessTestHelper(time.Second * 5)

	//act
	var err error = nil
	go func() {
		err = harnessTestHelper.Harness.Run(harnessTestHelper.Ctx)
	}()

	//start the host
	harnessTestHelper.Harness.hostingStart <- &events.HostingStart{
		GameSessionId: "arn:aws:gamelift:eu-west-1::gamesession/fleet-8a8e55eb-6607-4eb3-9031-eed48907d5a4/dev-jim/6aa3a161-f2fb-4b53-bfd9-1f31c3b20cd2",
		FleetId:       "FleetId",
		Provider:      "Provider",
	}

	time.Sleep(time.Second * 1)

	//assert
	assert.Nil(t, err)
	logBuffer := harnessTestHelper.LogBuffer.String()
	assert.Contains(t, logBuffer, "Game server harness starting")
	assert.Contains(t, logBuffer, "Received hosting start event")
	assert.NotContains(t, logBuffer, "Received hosting terminate event")
	assert.NotContains(t, logBuffer, "Game server harness context done")
	assert.Contains(t, logBuffer, "Hosting start result")
	assert.NotContains(t, logBuffer, "Hosting terminate result")
	assert.Contains(t, logBuffer, "6aa3a161-f2fb-4b53-bfd9-1f31c3b20cd2")
	assert.True(t, harnessTestHelper.GameService.RunCalled)
}

func Test_Harness_Run_HappyPath_Terminate_Received(t *testing.T) {
	//arrange
	harnessTestHelper := CreateHarnessTestHelper(time.Second * 5)

	//act
	var err error = nil
	go func() {
		err = harnessTestHelper.Harness.Run(harnessTestHelper.Ctx)
	}()

	//start the host
	harnessTestHelper.Harness.hostingTerminate <- &events.HostingTerminate{
		Reason: "Unit Test",
	}

	time.Sleep(time.Second * 1)

	//assert
	assert.Nil(t, err)
	logBuffer := harnessTestHelper.LogBuffer.String()
	assert.Contains(t, logBuffer, "Game server harness starting")
	assert.NotContains(t, logBuffer, "Received hosting start event")
	assert.Contains(t, logBuffer, "Received hosting terminate event")
	assert.NotContains(t, logBuffer, "Game server harness context done")
	assert.NotContains(t, logBuffer, "Hosting start result")
	assert.Contains(t, logBuffer, "Hosting terminate result")

	assert.True(t, harnessTestHelper.GameService.StopCalled)
}

func Test_Harness_Close_HappyPath(t *testing.T) {
	//arrange
	harnessTestHelper := CreateHarnessTestHelper(time.Second * 5)

	//act
	err := harnessTestHelper.Harness.Close(harnessTestHelper.Ctx)

	//assert
	assert.Nil(t, err)
	assert.True(t, harnessTestHelper.GameService.StopCalled)
}
