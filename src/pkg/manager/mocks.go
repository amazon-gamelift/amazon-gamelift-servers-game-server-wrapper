/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manager

import (
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/game"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/types/events"
	"golang.org/x/net/context"
	"time"
)

type GameServiceMock struct {
	InitError  error
	InitCalled bool
	InitCount  int

	RunError  error
	RunCalled bool
	RunCount  int

	HealthCheckCalled bool
	HealthCheckCount  int

	StopError  error
	StopCalled bool
	StopCount  int

	InitArgs   *game.InitArgs
	StartArgs  *game.StartArgs
	GameStatus *events.GameStatus
	InitMeta   *game.InitMeta

	Delay time.Duration
}

func (gameServiceMock *GameServiceMock) Init(ctx context.Context, args *game.InitArgs) (*game.InitMeta, error) {
	gameServiceMock.InitCalled = true
	gameServiceMock.InitCount++
	gameServiceMock.InitArgs = args
	return gameServiceMock.InitMeta, gameServiceMock.InitError
}

func (gameServiceMock *GameServiceMock) Run(ctx context.Context, args *game.StartArgs) error {
	gameServiceMock.RunCalled = true
	gameServiceMock.RunCount++
	gameServiceMock.StartArgs = args
	time.Sleep(gameServiceMock.Delay)
	return gameServiceMock.RunError
}

func (gameServiceMock *GameServiceMock) HealthCheck(ctx context.Context) events.GameStatus {
	gameServiceMock.HealthCheckCalled = true
	gameServiceMock.HealthCheckCount++
	return *gameServiceMock.GameStatus
}

func (gameServiceMock *GameServiceMock) Stop(ctx context.Context) error {
	gameServiceMock.StopCalled = true
	gameServiceMock.StopCount++
	return gameServiceMock.StopError
}

type HarnessMock struct {
	InitError  error
	InitCalled bool
	InitCount  int

	RunError  error
	RunCalled bool
	RunCount  int

	HostingStartCalled bool
	HostingStartCount  int
	HostingStartError  error

	HostingTerminateCalled bool
	HostingTerminateCount  int
	HostingTerminateError  error

	HealthCheckCalled bool
	HealthCheckCount  int

	CloseError  error
	CloseCalled bool
	CloseCount  int

	InitArgs              *game.InitArgs
	InitMeta              *game.InitMeta
	HostingStartEvent     *events.HostingStart
	HostingTerminateEvent *events.HostingTerminate
	GameStatus            *events.GameStatus

	Delay time.Duration
}

func (harnessMock *HarnessMock) Init(ctx context.Context, args *game.InitArgs) (*game.InitMeta, error) {
	harnessMock.InitCalled = true
	harnessMock.InitCount++
	harnessMock.InitArgs = args
	return harnessMock.InitMeta, harnessMock.InitError
}

func (harnessMock *HarnessMock) Run(ctx context.Context) error {
	harnessMock.RunCalled = true
	harnessMock.RunCount++
	time.Sleep(harnessMock.Delay)
	return harnessMock.RunError
}

func (harnessMock *HarnessMock) HostingStart(ctx context.Context, h *events.HostingStart, end <-chan error) error {
	harnessMock.HostingStartCalled = true
	harnessMock.HostingStartCount++
	harnessMock.HostingStartEvent = h
	return harnessMock.HostingStartError
}

func (harnessMock *HarnessMock) HostingTerminate(ctx context.Context, h *events.HostingTerminate) error {
	harnessMock.HostingTerminateCalled = true
	harnessMock.HostingTerminateCount++
	harnessMock.HostingTerminateEvent = h
	return harnessMock.HostingTerminateError
}

func (harnessMock *HarnessMock) HealthCheck(ctx context.Context) events.GameStatus {
	harnessMock.HealthCheckCalled = true
	harnessMock.HealthCheckCount++
	return *harnessMock.GameStatus
}

func (harnessMock *HarnessMock) Close(ctx context.Context) error {
	harnessMock.CloseCalled = true
	harnessMock.CloseCount++
	return harnessMock.CloseError
}
