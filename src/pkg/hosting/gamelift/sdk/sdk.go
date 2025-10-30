/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package sdk

import (
	"context"
	"log/slog"

	"github.com/amazon-gamelift/amazon-gamelift-servers-go-server-sdk/v5/server"
)

// GameLiftSdk defines the interface for interacting with server SDK for Amazon GameLift Servers.
type GameLiftSdk interface {
	// InitSDK initializes the Amazon GameLift SDK.
	InitSDK(ctx context.Context, params server.ServerParameters) error

	// InitSDKFromEnvironment initializes the Amazon GameLift SDK using environment variables for configuration.
	InitSDKFromEnvironment(ctx context.Context) error

	// ProcessReady notifies the Amazon GameLift service that the server process is ready to host game sessions.
	ProcessReady(ctx context.Context, params server.ProcessParameters) error

	// ProcessEnding notifies the Amazon GameLift service that the server process is shutting down.
	ProcessEnding(ctx context.Context) error

	// ActivateGameSession notifies Amazon GameLift that the server process has
	// activated a game session and is now ready to receive player connections.
	ActivateGameSession(ctx context.Context) error

	// Destroy frees the server SDK for Amazon GameLift Servers from memory.
	Destroy(ctx context.Context) error
}

type Sdk struct {
	logger *slog.Logger
}

func (sdk *Sdk) InitSDK(ctx context.Context, params server.ServerParameters) error {
	redactedParams := params
	redactedParams.AuthToken = "<REDACTED>"
	redactedParams.AccessKey = "<REDACTED>"
	redactedParams.SecretKey = "<REDACTED>"
	redactedParams.SessionToken = "<REDACTED>"
	sdk.logger.DebugContext(ctx, "InitSDK called", "params", redactedParams)
	return server.InitSDK(params)
}

func (sdk *Sdk) InitSDKFromEnvironment(ctx context.Context) error {
	sdk.logger.DebugContext(ctx, "InitSDKFromEnvironment called")
	return server.InitSDKFromEnvironment()
}

func (sdk *Sdk) ProcessReady(ctx context.Context, params server.ProcessParameters) error {
	sdk.logger.DebugContext(ctx, "ProcessReady called", "port", params.Port, "logParams", params.LogParameters)
	return server.ProcessReady(params)
}

func (sdk *Sdk) ProcessEnding(ctx context.Context) error {
	sdk.logger.DebugContext(ctx, "ProcessEnding called")
	return server.ProcessEnding()
}

func (sdk *Sdk) ActivateGameSession(ctx context.Context) error {
	sdk.logger.DebugContext(ctx, "ActivateGameSession called")
	return server.ActivateGameSession()
}

func (sdk *Sdk) Destroy(ctx context.Context) error {
	sdk.logger.DebugContext(ctx, "Destroy called")
	return server.Destroy()
}

func NewSdk(ctx context.Context, logger *slog.Logger) *Sdk {
	s := &Sdk{
		logger: logger,
	}

	server.SetLoggerInterface(NewLogAdaptor(ctx, logger))

	return s
}
