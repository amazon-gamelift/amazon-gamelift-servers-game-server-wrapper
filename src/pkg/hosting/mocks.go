/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package hosting

import (
	"time"

	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/types/events"
	"github.com/amazon-gamelift/amazon-gamelift-servers-go-server-sdk/v5/model"
	"golang.org/x/net/context"
)

type HostingServiceMock struct {
	InitError  error
	InitCalled bool
	InitCount  int
	InitArgs   *InitArgs

	RunError  error
	RunCalled bool
	RunCount  int

	SetOnHostingStartError  error
	SetOnHostingStartCalled bool
	SetOnHostingStartCount  int
	OnHostingStart          func(ctx context.Context, h *events.HostingStart, end <-chan error) error

	SetOnHostingTerminateCalled bool
	SetOnHostingTerminateCount  int
	OnHostingTerminate          func(ctx context.Context, h *events.HostingTerminate) error

	SetOnHealthCheckCalled bool
	SetOnHealthCheckCount  int
	OnHealthCheck          func(ctx context.Context) events.GameStatus

	CloseError  error
	CloseCalled bool
	CloseCount  int

	GameStatusEvent   *events.GameStatus
	HostingStartEvent *events.HostingStart
	InitMeta          *InitMeta

	Delay time.Duration
}

func (hostingServiceMock *HostingServiceMock) Init(ctx context.Context, args *InitArgs) (*InitMeta, error) {
	hostingServiceMock.InitCalled = true
	hostingServiceMock.InitCount++
	hostingServiceMock.InitArgs = args
	return hostingServiceMock.InitMeta, hostingServiceMock.InitError
}

func (hostingServiceMock *HostingServiceMock) Run(ctx context.Context) error {
	hostingServiceMock.RunCalled = true
	hostingServiceMock.RunCount++
	time.Sleep(hostingServiceMock.Delay)
	return hostingServiceMock.RunError
}

func (hostingServiceMock *HostingServiceMock) SetOnHostingStart(f func(ctx context.Context, h *events.HostingStart, end <-chan error) error) {
	hostingServiceMock.SetOnHostingStartCalled = true
	hostingServiceMock.SetOnHostingStartCount++
	hostingServiceMock.OnHostingStart = f
}

func (hostingServiceMock *HostingServiceMock) SetOnHostingTerminate(f func(ctx context.Context, h *events.HostingTerminate) error) {
	hostingServiceMock.SetOnHostingTerminateCalled = true
	hostingServiceMock.SetOnHostingTerminateCount++
	hostingServiceMock.OnHostingTerminate = f
}

func (hostingServiceMock *HostingServiceMock) SetOnHealthCheck(f func(ctx context.Context) events.GameStatus) {
	hostingServiceMock.SetOnHealthCheckCalled = true
	hostingServiceMock.CloseCount++
	hostingServiceMock.OnHealthCheck = f
}

func (hostingServiceMock *HostingServiceMock) Close(ctx context.Context) error {
	hostingServiceMock.CloseCalled = true
	hostingServiceMock.SetOnHealthCheckCount++
	return hostingServiceMock.CloseError
}

type CustomMessageSenderMock struct {
	OnUpdateGameSessionError  error
	OnUpdateGameSessionInput  model.UpdateGameSession
	OnUpdateGameSessionCalled bool
	OnUpdateGameSessionCount  int

	OnHealthCheckError  error
	OnHealthCheckCalled bool
	OnHealthCheckCount  int

	OnHostingTerminateError  error
	OnHostingTerminateCalled bool
	OnHostingTerminateCount  int

	OnStartGameSessionError  error
	OnStartGameSessionInput  model.GameSession
	OnStartGameSessionCalled bool
	OnStartGameSessionCount  int
}

func (c *CustomMessageSenderMock) OnUpdateGameSession(_ context.Context, gs model.UpdateGameSession) error {
	c.OnUpdateGameSessionCalled = true
	c.OnUpdateGameSessionCount++
	c.OnUpdateGameSessionInput = gs
	return c.OnUpdateGameSessionError
}

func (c *CustomMessageSenderMock) OnHealthCheck(_ context.Context) error {
	c.OnHealthCheckCalled = true
	c.OnHealthCheckCount++
	return c.OnHealthCheckError
}

func (c *CustomMessageSenderMock) OnHostingTerminate(_ context.Context) error {
	c.OnHostingTerminateCalled = true
	c.OnHostingTerminateCount++
	return c.OnHostingTerminateError
}

func (c *CustomMessageSenderMock) OnStartGameSession(_ context.Context, gs model.GameSession) error {
	c.OnStartGameSessionCalled = true
	c.OnStartGameSessionCount++
	c.OnStartGameSessionInput = gs
	return c.OnStartGameSessionError
}
