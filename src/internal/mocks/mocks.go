/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mocks

import (
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/process"
	"github.com/amazon-gamelift/amazon-gamelift-servers-go-server-sdk/v5/server"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/net/context"
)

type SpannerMock struct {
}

func (spannerMock *SpannerMock) NewSpan(ctx context.Context, name string, meta map[string]string) (context.Context, trace.Span, error) {
	return ctx, trace.SpanFromContext(ctx), nil
}

func (spannerMock *SpannerMock) NewSpanWithTraceId(ctx context.Context, name string, traceId uuid.UUID, meta map[string]string) (context.Context, trace.Span, error) {
	return ctx, trace.SpanFromContext(ctx), nil
}

type ProcessMock struct {
	InitResponse      error
	RunResultResponse *process.Result
	RunErrorResponse  error
	StateResponse     *process.State
}

func (processMock *ProcessMock) Init(ctx context.Context) error {
	return processMock.InitResponse
}

func (processMock *ProcessMock) Run(ctx context.Context, args *process.Args, pidChan chan<- int) (*process.Result, error) {
	return processMock.RunResultResponse, processMock.RunErrorResponse
}

func (processMock *ProcessMock) State() *process.State {
	return processMock.StateResponse
}

type GameLiftSdkMock struct {
	InitSdkError                 error
	InitSDKFromEnvironmentError  error
	ProcessReadyError            error
	ProcessEndingError           error
	ActivateGameSessionError     error
	DestroyError                 error
	ProcessParameters            *server.ProcessParameters
	ServerParameters             *server.ServerParameters
	InitSdkCalled                bool
	InitSDKFromEnvironmentCalled bool
	ProcessReadyCalled           bool
	ProcessEndingCalled          bool
	ActivateGameSessionCalled    bool
	DestroyCalled                bool
}

func (gameLiftSdkMock *GameLiftSdkMock) InitSDK(ctx context.Context, params server.ServerParameters) error {
	gameLiftSdkMock.ServerParameters = &params
	gameLiftSdkMock.InitSdkCalled = true
	return gameLiftSdkMock.InitSdkError
}

func (gameLiftSdkMock *GameLiftSdkMock) InitSDKFromEnvironment(ctx context.Context) error {
	gameLiftSdkMock.InitSDKFromEnvironmentCalled = true
	return gameLiftSdkMock.InitSdkError
}

func (gameLiftSdkMock *GameLiftSdkMock) ProcessReady(ctx context.Context, params server.ProcessParameters) error {
	gameLiftSdkMock.ProcessReadyCalled = true
	gameLiftSdkMock.ProcessParameters = &params
	return gameLiftSdkMock.InitSdkError
}

func (gameLiftSdkMock *GameLiftSdkMock) ProcessEnding(ctx context.Context) error {
	gameLiftSdkMock.ProcessEndingCalled = true
	return gameLiftSdkMock.InitSdkError
}

func (gameLiftSdkMock *GameLiftSdkMock) ActivateGameSession(ctx context.Context) error {
	gameLiftSdkMock.ActivateGameSessionCalled = true
	return gameLiftSdkMock.InitSdkError
}

func (gameLiftSdkMock *GameLiftSdkMock) Destroy(ctx context.Context) error {
	gameLiftSdkMock.DestroyCalled = true
	return gameLiftSdkMock.InitSdkError
}
