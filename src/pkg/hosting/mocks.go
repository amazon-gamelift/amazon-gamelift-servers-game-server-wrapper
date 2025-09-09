/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package hosting

import (
	"time"

	"github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/pkg/types/events"
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
