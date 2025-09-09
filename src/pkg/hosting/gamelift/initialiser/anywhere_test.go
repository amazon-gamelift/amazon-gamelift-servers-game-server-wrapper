/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package initialiser

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/internal/mocks"
	"github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/pkg/config"
	"github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/pkg/hosting/gamelift/client"
	"github.com/aws/aws-sdk-go-v2/service/gamelift"
	"github.com/aws/aws-sdk-go-v2/service/gamelift/types"
	"github.com/aws/smithy-go/middleware"
	"github.com/stretchr/testify/assert"
)

type AnywhereMockHelper struct {
	logger         *slog.Logger
	logBuffer      *bytes.Buffer
	gameLiftSdk    *mocks.GameLiftSdkMock
	ctx            context.Context
	config         *config.Anywhere
	clientProvider *client.ClientProviderMock
	clientGameLift *client.ClientGameLiftMock
	anywhere       Service
}

// Constants that are provided to calls to RegisterCompute. They must be var so they can be passed by reference.
var (
	fleetArn = "arn:aws:gamelift:eu-west-2:645463837486:fleet/fleet-b87ed01d-4228-42c6-8291-d0ee0d5eb015"
	fleetId  = "fleet-b87ed01d-4228-42c6-8291-d0ee0d5eb015"
)

const (
	locationArn = "arn:aws:gamelift:eu-west-2:645463837486:location/custom-gamelift"
	IPv4Address = "192.168.1.1"
)

func createAnywhereMockHelper(config *config.Anywhere) AnywhereMockHelper {
	logBuffer := bytes.Buffer{}
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	logger := slog.New(slog.NewTextHandler(&logBuffer, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	gameLiftSdkMock := mocks.GameLiftSdkMock{}
	clientProvider := client.ClientProviderMock{}
	clientGameLift := client.ClientGameLiftMock{}
	anywhere, _ := newAnywhere(ctx, config, &gameLiftSdkMock, logger, &clientProvider)
	return AnywhereMockHelper{
		logger:         logger,
		logBuffer:      &logBuffer,
		gameLiftSdk:    &gameLiftSdkMock,
		ctx:            ctx,
		config:         config,
		clientProvider: &clientProvider,
		clientGameLift: &clientGameLift,
		anywhere:       anywhere,
	}
}

func Test_Anywhere_InitSdk_HappyPath(t *testing.T) {
	//arrange
	anyWhereConfig := config.Anywhere{
		Config: config.AwsConfig{},
		Host: config.AnywhereHostConfig{
			LocationArn: locationArn,
			FleetArn:    fleetArn,
			IPv4Address: IPv4Address,
		},
	}
	gameLiftMockHelper := createAnywhereMockHelper(&anyWhereConfig)
	gameLiftMockHelper.clientProvider.GetGameLiftResponse = gameLiftMockHelper.clientGameLift
	gameLiftServiceSdkEndpoint := "GameLiftServiceSdkEndpoint"
	gameLiftMockHelper.clientGameLift.RegisterComputeResult = &gamelift.RegisterComputeOutput{
		Compute: &types.Compute{
			ComputeArn:                 nil,
			ComputeName:                nil,
			ComputeStatus:              "",
			CreationTime:               nil,
			DnsName:                    nil,
			FleetArn:                   &fleetArn,
			FleetId:                    &fleetId,
			GameLiftServiceSdkEndpoint: &gameLiftServiceSdkEndpoint,
			IpAddress:                  nil,
			Location:                   nil,
			OperatingSystem:            "",
			Type:                       "",
		},
		ResultMetadata: middleware.Metadata{},
	}
	gameLiftMockHelper.clientGameLift.ListComputeResult = &gamelift.ListComputeOutput{
		ComputeList:    []types.Compute{},
		NextToken:      nil,
		ResultMetadata: middleware.Metadata{},
	}

	getComputeAuthTokenOutputComputeName := "GetComputeAuthTokenOutputComputeName"
	getComputeAuthTokenOutputAuthToken := "GetComputeAuthTokenOutputAuthToken"
	gameLiftMockHelper.clientGameLift.GetComputeAuthTokenResult = &gamelift.GetComputeAuthTokenOutput{
		AuthToken:           &getComputeAuthTokenOutputAuthToken,
		ComputeArn:          nil,
		ComputeName:         &getComputeAuthTokenOutputComputeName,
		ExpirationTimestamp: nil,
		FleetArn:            &fleetArn,
		FleetId:             &fleetId,
		ResultMetadata:      middleware.Metadata{},
	}

	//act
	err := gameLiftMockHelper.anywhere.InitSdk(gameLiftMockHelper.ctx)

	//assert
	assert.Nil(t, err)
	logString := gameLiftMockHelper.logBuffer.String()
	assert.Contains(t, logString, "using Amazon GameLift Anywhere")
	assert.Contains(t, logString, "listing compute")
	assert.Contains(t, logString, "registering compute")
	assert.Contains(t, logString, "getting auth token")
	assert.Contains(t, logString, "initialising sdk")
	assert.True(t, gameLiftMockHelper.clientProvider.GetGameLiftCalled)
	assert.Equal(t, "fleet-b87ed01d-4228-42c6-8291-d0ee0d5eb015", *gameLiftMockHelper.clientGameLift.ListComputeInput.FleetId)
	assert.Equal(t, "fleet-b87ed01d-4228-42c6-8291-d0ee0d5eb015", *gameLiftMockHelper.clientGameLift.RegisterComputeInput.FleetId)
	assert.Equal(t, "custom-gamelift", *gameLiftMockHelper.clientGameLift.RegisterComputeInput.Location)
	assert.Contains(t, *gameLiftMockHelper.clientGameLift.RegisterComputeInput.ComputeName, anyWhereConfig.Host.HostName)
	assert.Equal(t, "fleet-b87ed01d-4228-42c6-8291-d0ee0d5eb015", *gameLiftMockHelper.clientGameLift.GetComputeAuthTokenInput.FleetId)
	assert.Contains(t, *gameLiftMockHelper.clientGameLift.GetComputeAuthTokenInput.ComputeName, anyWhereConfig.Host.HostName)
	assert.True(t, gameLiftMockHelper.gameLiftSdk.InitSdkCalled)
	assert.Equal(t, gameLiftServiceSdkEndpoint, gameLiftMockHelper.gameLiftSdk.ServerParameters.WebSocketURL)
	hostname, _ := getHostname()
	assert.Equal(t, gameLiftMockHelper.gameLiftSdk.ServerParameters.HostID, hostname)
	assert.Equal(t, gameLiftMockHelper.gameLiftSdk.ServerParameters.WebSocketURL, gameLiftServiceSdkEndpoint)
	assert.Equal(t, gameLiftMockHelper.gameLiftSdk.ServerParameters.AuthToken, getComputeAuthTokenOutputAuthToken)
}

func Test_Anywhere_InitSdk_HappyPath_ProvidedCompute(t *testing.T) {
	//arrange
	anyWhereConfig := config.Anywhere{
		Config: config.AwsConfig{},
		Host: config.AnywhereHostConfig{
			HostName:           "UnitTest",
			ServiceSdkEndpoint: "endpoint",
			LocationArn:        locationArn,
			FleetArn:           fleetArn,
			IPv4Address:        IPv4Address,
		},
	}
	gameLiftMockHelper := createAnywhereMockHelper(&anyWhereConfig)
	gameLiftMockHelper.clientProvider.GetGameLiftResponse = gameLiftMockHelper.clientGameLift

	getComputeAuthTokenOutputComputeName := "GetComputeAuthTokenOutputComputeName"
	getComputeAuthTokenOutputAuthToken := "GetComputeAuthTokenOutputAuthToken"
	gameLiftMockHelper.clientGameLift.GetComputeAuthTokenResult = &gamelift.GetComputeAuthTokenOutput{
		AuthToken:           &getComputeAuthTokenOutputAuthToken,
		ComputeArn:          nil,
		ComputeName:         &getComputeAuthTokenOutputComputeName,
		ExpirationTimestamp: nil,
		FleetArn:            &fleetArn,
		FleetId:             &fleetId,
		ResultMetadata:      middleware.Metadata{},
	}

	//act
	err := gameLiftMockHelper.anywhere.InitSdk(gameLiftMockHelper.ctx)

	//assert
	assert.Nil(t, err)
	logString := gameLiftMockHelper.logBuffer.String()
	assert.Contains(t, logString, "using Amazon GameLift Anywhere")
	assert.NotContains(t, logString, "listing compute")
	assert.NotContains(t, logString, "registering compute")
	assert.Contains(t, logString, "using pre-set Anywhere configuration")
	assert.Contains(t, logString, "getting auth token")
	assert.Contains(t, logString, "initialising sdk")
	assert.True(t, gameLiftMockHelper.clientProvider.GetGameLiftCalled)
	assert.Equal(t, "fleet-b87ed01d-4228-42c6-8291-d0ee0d5eb015", *gameLiftMockHelper.clientGameLift.GetComputeAuthTokenInput.FleetId)
	assert.Contains(t, *gameLiftMockHelper.clientGameLift.GetComputeAuthTokenInput.ComputeName, anyWhereConfig.Host.HostName)
	assert.True(t, gameLiftMockHelper.gameLiftSdk.InitSdkCalled)
	assert.Equal(t, gameLiftMockHelper.gameLiftSdk.ServerParameters.HostID, anyWhereConfig.Host.HostName)
	assert.Equal(t, gameLiftMockHelper.gameLiftSdk.ServerParameters.WebSocketURL, anyWhereConfig.Host.ServiceSdkEndpoint)
	assert.Equal(t, gameLiftMockHelper.gameLiftSdk.ServerParameters.AuthToken, getComputeAuthTokenOutputAuthToken)
}

func Test_Anywhere_InitSdk_HappyPath_ProvidedComputeAndAuthToken(t *testing.T) {
	//arrange
	anyWhereConfig := config.Anywhere{
		Config: config.AwsConfig{},
		Host: config.AnywhereHostConfig{
			HostName:           "UnitTest",
			ServiceSdkEndpoint: "endpoint",
			AuthToken:          "authToken",
			LocationArn:        locationArn,
			FleetArn:           fleetArn,
			IPv4Address:        IPv4Address,
		},
	}
	gameLiftMockHelper := createAnywhereMockHelper(&anyWhereConfig)
	gameLiftMockHelper.clientProvider.GetGameLiftResponse = gameLiftMockHelper.clientGameLift

	//act
	err := gameLiftMockHelper.anywhere.InitSdk(gameLiftMockHelper.ctx)

	//assert
	assert.Nil(t, err)
	logString := gameLiftMockHelper.logBuffer.String()
	assert.Contains(t, logString, "using Amazon GameLift Anywhere")
	assert.NotContains(t, logString, "listing compute")
	assert.NotContains(t, logString, "registering compute")
	assert.Contains(t, logString, "using pre-set Anywhere configuration")
	assert.NotContains(t, logString, "getting auth token")
	assert.Contains(t, logString, "initialising sdk")
	assert.False(t, gameLiftMockHelper.clientProvider.GetGameLiftCalled)
	assert.True(t, gameLiftMockHelper.gameLiftSdk.InitSdkCalled)
	assert.Equal(t, gameLiftMockHelper.gameLiftSdk.ServerParameters.HostID, anyWhereConfig.Host.HostName)
	assert.Equal(t, gameLiftMockHelper.gameLiftSdk.ServerParameters.WebSocketURL, anyWhereConfig.Host.ServiceSdkEndpoint)
	assert.Equal(t, gameLiftMockHelper.gameLiftSdk.ServerParameters.AuthToken, anyWhereConfig.Host.AuthToken)
}

func Test_Anywhere_InitSdk_GetGameLift_Error(t *testing.T) {
	//arrange
	anyWhereConfig := config.Anywhere{
		Config: config.AwsConfig{},
		Host: config.AnywhereHostConfig{
			HostName:    "UnitTest",
			LocationArn: locationArn,
			FleetArn:    fleetArn,
			IPv4Address: IPv4Address,
		},
	}
	gameLiftMockHelper := createAnywhereMockHelper(&anyWhereConfig)
	gameLiftMockHelper.clientProvider.GetGameLiftError = errors.New("Unit Test")

	//act
	err := gameLiftMockHelper.anywhere.InitSdk(gameLiftMockHelper.ctx)

	//assert
	assert.NotNil(t, err)
	assert.Errorf(t, err, "failed to get Amazon GameLift client provider")
	logString := gameLiftMockHelper.logBuffer.String()
	assert.Contains(t, logString, "using Amazon GameLift Anywhere")
	assert.True(t, gameLiftMockHelper.clientProvider.GetGameLiftCalled)
}

func Test_Anywhere_InitSdk_ListCompute_Error(t *testing.T) {
	//arrange
	anyWhereConfig := config.Anywhere{
		Config: config.AwsConfig{},
		Host: config.AnywhereHostConfig{
			HostName:    "UnitTest",
			LocationArn: locationArn,
			FleetArn:    fleetArn,
			IPv4Address: IPv4Address,
		},
	}
	gameLiftMockHelper := createAnywhereMockHelper(&anyWhereConfig)
	gameLiftMockHelper.clientProvider.GetGameLiftResponse = gameLiftMockHelper.clientGameLift
	gameLiftMockHelper.clientGameLift.ListComputeError = errors.New("Unit Test")

	//act
	err := gameLiftMockHelper.anywhere.InitSdk(gameLiftMockHelper.ctx)

	//assert
	assert.NotNil(t, err)
	assert.Errorf(t, err, "failed to list compute")
	logString := gameLiftMockHelper.logBuffer.String()
	assert.Contains(t, logString, "using Amazon GameLift Anywhere")
	assert.Contains(t, logString, "listing compute")
	assert.True(t, gameLiftMockHelper.clientProvider.GetGameLiftCalled)
	assert.Equal(t, "fleet-b87ed01d-4228-42c6-8291-d0ee0d5eb015", *gameLiftMockHelper.clientGameLift.ListComputeInput.FleetId)
}

func Test_Anywhere_InitSdk_RegisterCompute_Error(t *testing.T) {
	//arrange
	anyWhereConfig := config.Anywhere{
		Config: config.AwsConfig{},
		Host: config.AnywhereHostConfig{
			HostName:    "UnitTest",
			LocationArn: locationArn,
			FleetArn:    fleetArn,
			IPv4Address: IPv4Address,
		},
	}
	gameLiftMockHelper := createAnywhereMockHelper(&anyWhereConfig)
	gameLiftMockHelper.clientProvider.GetGameLiftResponse = gameLiftMockHelper.clientGameLift
	gameLiftMockHelper.clientGameLift.ListComputeResult = &gamelift.ListComputeOutput{
		ComputeList:    []types.Compute{},
		NextToken:      nil,
		ResultMetadata: middleware.Metadata{},
	}

	gameLiftMockHelper.clientGameLift.RegisterComputeError = errors.New("Unit Test")

	//act
	err := gameLiftMockHelper.anywhere.InitSdk(gameLiftMockHelper.ctx)

	//assert
	assert.NotNil(t, err)
	assert.Errorf(t, err, "failed to list compute")
	logString := gameLiftMockHelper.logBuffer.String()
	assert.Contains(t, logString, "using Amazon GameLift Anywhere")
	assert.Contains(t, logString, "listing compute")
	assert.Contains(t, logString, "registering compute")
	assert.True(t, gameLiftMockHelper.clientProvider.GetGameLiftCalled)
	assert.Equal(t, "fleet-b87ed01d-4228-42c6-8291-d0ee0d5eb015", *gameLiftMockHelper.clientGameLift.ListComputeInput.FleetId)
	assert.Equal(t, "fleet-b87ed01d-4228-42c6-8291-d0ee0d5eb015", *gameLiftMockHelper.clientGameLift.RegisterComputeInput.FleetId)
	assert.Equal(t, "custom-gamelift", *gameLiftMockHelper.clientGameLift.RegisterComputeInput.Location)
	assert.Contains(t, *gameLiftMockHelper.clientGameLift.RegisterComputeInput.ComputeName, anyWhereConfig.Host.HostName)
}

func Test_Anywhere_InitSdk_GetComputeAuthToken_Error(t *testing.T) {
	//arrange
	anyWhereConfig := config.Anywhere{
		Config: config.AwsConfig{},
		Host: config.AnywhereHostConfig{
			HostName:    "UnitTest",
			LocationArn: locationArn,
			FleetArn:    fleetArn,
			IPv4Address: IPv4Address,
		},
	}
	gameLiftMockHelper := createAnywhereMockHelper(&anyWhereConfig)
	gameLiftMockHelper.clientProvider.GetGameLiftResponse = gameLiftMockHelper.clientGameLift
	gameLiftServiceSdkEndpoint := "GameLiftServiceSdkEndpoint"
	gameLiftMockHelper.clientGameLift.RegisterComputeResult = &gamelift.RegisterComputeOutput{
		Compute: &types.Compute{
			ComputeArn:                 nil,
			ComputeName:                nil,
			ComputeStatus:              "",
			CreationTime:               nil,
			DnsName:                    nil,
			FleetArn:                   &fleetArn,
			FleetId:                    &fleetId,
			GameLiftServiceSdkEndpoint: &gameLiftServiceSdkEndpoint,
			IpAddress:                  nil,
			Location:                   nil,
			OperatingSystem:            "",
			Type:                       "",
		},
		ResultMetadata: middleware.Metadata{},
	}
	gameLiftMockHelper.clientGameLift.ListComputeResult = &gamelift.ListComputeOutput{
		ComputeList:    []types.Compute{},
		NextToken:      nil,
		ResultMetadata: middleware.Metadata{},
	}

	gameLiftMockHelper.clientGameLift.GetComputeAuthTokenError = errors.New("Unit Test")

	//act
	err := gameLiftMockHelper.anywhere.InitSdk(gameLiftMockHelper.ctx)

	//assert
	assert.NotNil(t, err)
	assert.Errorf(t, err, "failed to get Amazon GameLift Anywhere token")
	logString := gameLiftMockHelper.logBuffer.String()
	assert.Contains(t, logString, "using Amazon GameLift Anywhere")
	assert.Contains(t, logString, "listing compute")
	assert.Contains(t, logString, "registering compute")
	assert.Contains(t, logString, "getting auth token")
	assert.True(t, gameLiftMockHelper.clientProvider.GetGameLiftCalled)
	assert.Equal(t, "fleet-b87ed01d-4228-42c6-8291-d0ee0d5eb015", *gameLiftMockHelper.clientGameLift.ListComputeInput.FleetId)
	assert.Equal(t, "fleet-b87ed01d-4228-42c6-8291-d0ee0d5eb015", *gameLiftMockHelper.clientGameLift.RegisterComputeInput.FleetId)
	assert.Equal(t, "custom-gamelift", *gameLiftMockHelper.clientGameLift.RegisterComputeInput.Location)
	assert.Contains(t, *gameLiftMockHelper.clientGameLift.RegisterComputeInput.ComputeName, anyWhereConfig.Host.HostName)
	assert.Equal(t, "fleet-b87ed01d-4228-42c6-8291-d0ee0d5eb015", *gameLiftMockHelper.clientGameLift.GetComputeAuthTokenInput.FleetId)
	assert.Contains(t, *gameLiftMockHelper.clientGameLift.GetComputeAuthTokenInput.ComputeName, anyWhereConfig.Host.HostName)
}

func Test_Anywhere_InitSdk_InitSDK_Error(t *testing.T) {
	//arrange
	anyWhereConfig := config.Anywhere{
		Config: config.AwsConfig{},
		Host: config.AnywhereHostConfig{
			HostName:    "UnitTest",
			LocationArn: locationArn,
			FleetArn:    fleetArn,
			IPv4Address: IPv4Address,
		},
	}
	gameLiftMockHelper := createAnywhereMockHelper(&anyWhereConfig)
	gameLiftMockHelper.clientProvider.GetGameLiftResponse = gameLiftMockHelper.clientGameLift
	gameLiftServiceSdkEndpoint := "GameLiftServiceSdkEndpoint"
	gameLiftMockHelper.clientGameLift.RegisterComputeResult = &gamelift.RegisterComputeOutput{
		Compute: &types.Compute{
			ComputeArn:                 nil,
			ComputeName:                nil,
			ComputeStatus:              "",
			CreationTime:               nil,
			DnsName:                    nil,
			FleetArn:                   &fleetArn,
			FleetId:                    &fleetId,
			GameLiftServiceSdkEndpoint: &gameLiftServiceSdkEndpoint,
			IpAddress:                  nil,
			Location:                   nil,
			OperatingSystem:            "",
			Type:                       "",
		},
		ResultMetadata: middleware.Metadata{},
	}
	gameLiftMockHelper.clientGameLift.ListComputeResult = &gamelift.ListComputeOutput{
		ComputeList:    []types.Compute{},
		NextToken:      nil,
		ResultMetadata: middleware.Metadata{},
	}

	getComputeAuthTokenOutputComputeName := "GetComputeAuthTokenOutputComputeName"
	getComputeAuthTokenOutputAuthToken := "GetComputeAuthTokenOutputAuthToken"
	gameLiftMockHelper.clientGameLift.GetComputeAuthTokenResult = &gamelift.GetComputeAuthTokenOutput{
		AuthToken:           &getComputeAuthTokenOutputAuthToken,
		ComputeArn:          nil,
		ComputeName:         &getComputeAuthTokenOutputComputeName,
		ExpirationTimestamp: nil,
		FleetArn:            &fleetArn,
		FleetId:             &fleetId,
		ResultMetadata:      middleware.Metadata{},
	}

	gameLiftMockHelper.gameLiftSdk.InitSdkError = errors.New("Unit Test")

	//act
	err := gameLiftMockHelper.anywhere.InitSdk(gameLiftMockHelper.ctx)

	//assert
	assert.NotNil(t, err)
	assert.Errorf(t, err, "failed to init sdk for Amazon GameLift Anywhere")
	logString := gameLiftMockHelper.logBuffer.String()
	assert.Contains(t, logString, "using Amazon GameLift Anywhere")
	assert.Contains(t, logString, "listing compute")
	assert.Contains(t, logString, "registering compute")
	assert.Contains(t, logString, "getting auth token")
	assert.Contains(t, logString, "initialising sdk")
	assert.True(t, gameLiftMockHelper.clientProvider.GetGameLiftCalled)
	assert.Equal(t, "fleet-b87ed01d-4228-42c6-8291-d0ee0d5eb015", *gameLiftMockHelper.clientGameLift.ListComputeInput.FleetId)
	assert.Equal(t, "fleet-b87ed01d-4228-42c6-8291-d0ee0d5eb015", *gameLiftMockHelper.clientGameLift.RegisterComputeInput.FleetId)
	assert.Equal(t, "custom-gamelift", *gameLiftMockHelper.clientGameLift.RegisterComputeInput.Location)
	assert.Contains(t, *gameLiftMockHelper.clientGameLift.RegisterComputeInput.ComputeName, anyWhereConfig.Host.HostName)
	assert.Equal(t, "fleet-b87ed01d-4228-42c6-8291-d0ee0d5eb015", *gameLiftMockHelper.clientGameLift.GetComputeAuthTokenInput.FleetId)
	assert.Contains(t, *gameLiftMockHelper.clientGameLift.GetComputeAuthTokenInput.ComputeName, anyWhereConfig.Host.HostName)
	assert.True(t, gameLiftMockHelper.gameLiftSdk.InitSdkCalled)
	assert.Equal(t, gameLiftServiceSdkEndpoint, gameLiftMockHelper.gameLiftSdk.ServerParameters.WebSocketURL)
}
