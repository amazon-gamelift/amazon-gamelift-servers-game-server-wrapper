/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package initialiser

import (
	"log/slog"

	"github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/pkg/config"
	"github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/pkg/hosting/gamelift/sdk"
	"golang.org/x/net/context"
)

type InitialiserServiceMock struct {
	InitSdkError  error
	InitSdkCalled bool
}

func (initialiserServiceMock *InitialiserServiceMock) InitSdk(ctx context.Context) error {
	initialiserServiceMock.InitSdkCalled = true
	return initialiserServiceMock.InitSdkError
}

type InitialiserServiceFactoryMock struct {
	GetServiceResponse Service
	GetServiceError    error
}

func (initialiserServiceFactoryMock *InitialiserServiceFactoryMock) GetService(ctx context.Context, anywhere config.Anywhere, gameLiftSdk sdk.GameLiftSdk, logger *slog.Logger) (Service, error) {
	return initialiserServiceFactoryMock.GetServiceResponse, initialiserServiceFactoryMock.GetServiceError
}
