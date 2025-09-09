/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package initialiser

import (
	"bytes"
	"log/slog"
	"testing"
	"time"

	"github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/internal/mocks"
	"github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/pkg/config"
	"github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/pkg/hosting/gamelift/client"
	"github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/pkg/hosting/gamelift/sdk"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

type InitialiserMockHelper struct {
	logger                    *slog.Logger
	logBuffer                 *bytes.Buffer
	gameLiftSdk               *mocks.GameLiftSdkMock
	initialiserService        *InitialiserServiceMock
	ctx                       context.Context
	config                    *config.Anywhere
	initialiserServiceFactory *InitialiserServiceFactory
}

func createInitialiserMockHelper(config *config.Anywhere) InitialiserMockHelper {
	logBuffer := bytes.Buffer{}
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	logger := slog.New(slog.NewTextHandler(&logBuffer, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	gameLiftSdkMock := mocks.GameLiftSdkMock{}
	initialiserService := InitialiserServiceMock{}
	initialiserServiceFactory := InitialiserServiceFactory{}
	return InitialiserMockHelper{
		logger:                    logger,
		logBuffer:                 &logBuffer,
		gameLiftSdk:               &gameLiftSdkMock,
		ctx:                       ctx,
		config:                    config,
		initialiserServiceFactory: &initialiserServiceFactory,
		initialiserService:        &initialiserService,
	}
}

func Test_GetService_HappyPath_Managed(t *testing.T) {
	//arrange

	anywhereConfig := config.Anywhere{
		Config: config.AwsConfig{},
		Host:   config.AnywhereHostConfig{},
	}
	initialiserMockHelper := createInitialiserMockHelper(&anywhereConfig)

	managedNewCalled := false
	managedNew = func(sdk sdk.GameLiftSdk, logger *slog.Logger) Service {
		managedNewCalled = true
		assert.Same(t, initialiserMockHelper.gameLiftSdk, sdk)
		return initialiserMockHelper.initialiserService
	}

	//act
	serviceResponse, err := initialiserMockHelper.initialiserServiceFactory.GetService(initialiserMockHelper.ctx, *initialiserMockHelper.config, initialiserMockHelper.gameLiftSdk, initialiserMockHelper.logger)

	//assert
	assert.NotNil(t, serviceResponse)
	assert.Nil(t, err)
	assert.True(t, managedNewCalled)
}

func Test_GetService_HappyPath_Anywhere(t *testing.T) {
	//arrange

	anywhereConfig := config.Anywhere{
		Config: config.AwsConfig{
			Region:   "eu-west-1",
			Provider: "aws-profile",
			Literal:  config.AwsConfigLiteral{},
			Profile:  "gamelift",
			SSOFile:  "",
		},
		Host: config.AnywhereHostConfig{
			HostName:    "gamelift",
			LocationArn: "custom-unit-tesg",
			FleetArn:    "arn:aws:gamelift:eu-west-2:645463837486:fleet/fleet-b87ed01d-4228-42c6-8291-d0ee0d5eb015",
			IPv4Address: "",
		},
	}
	initialiserMockHelper := createInitialiserMockHelper(&anywhereConfig)

	anywhereNewCalled := false
	anywhereNew = func(ctx context.Context, cfg *config.Anywhere, gl sdk.GameLiftSdk, logger *slog.Logger, clientProvider client.Provider) (Service, error) {
		anywhereNewCalled = true
		assert.Same(t, initialiserMockHelper.gameLiftSdk, gl)
		return initialiserMockHelper.initialiserService, nil
	}

	//act
	serviceResponse, err := initialiserMockHelper.initialiserServiceFactory.GetService(initialiserMockHelper.ctx, *initialiserMockHelper.config, initialiserMockHelper.gameLiftSdk, initialiserMockHelper.logger)

	//assert
	assert.NotNil(t, serviceResponse)
	assert.Nil(t, err)
	assert.True(t, anywhereNewCalled)
}
