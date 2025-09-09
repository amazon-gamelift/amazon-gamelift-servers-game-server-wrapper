/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package gamelift

import (
	"bytes"
	"errors"
	"log/slog"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/internal"
	"github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/internal/mocks"
	config2 "github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/pkg/config"
	"github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/pkg/constants"
	"github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/pkg/hosting"
	"github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/pkg/hosting/gamelift/initialiser"
	"github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/pkg/observability"
	"github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/pkg/types/events"
	"github.com/amazon-gamelift/amazon-gamelift-servers-go-server-sdk/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

type GameLiftMockHelper struct {
	logger                        *slog.Logger
	logBuffer                     *bytes.Buffer
	spanner                       observability.Spanner
	gameLiftSdk                   *mocks.GameLiftSdkMock
	ctx                           context.Context
	gamelift                      *gamelift
	config                        *Config
	initialiserService            *initialiser.InitialiserServiceMock
	initialiserServiceFactoryMock *initialiser.InitialiserServiceFactoryMock
}

func createGameLiftMockHelper(config *Config) GameLiftMockHelper {
	logBuffer := bytes.Buffer{}
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	ctx = context.WithValue(ctx, constants.ContextKeyRunLogDir, config.LogDirectory)
	logger := slog.New(slog.NewTextHandler(&logBuffer, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	spannerMock := mocks.SpannerMock{}
	gameLiftSdkMock := mocks.GameLiftSdkMock{}
	initialiserServiceMock := initialiser.InitialiserServiceMock{
		InitSdkError: nil,
	}
	initialiserServiceFactoryMock := initialiser.InitialiserServiceFactoryMock{
		GetServiceResponse: &initialiserServiceMock,
		GetServiceError:    nil,
	}
	gamelift, _ := New(ctx, config, logger, &spannerMock, &initialiserServiceFactoryMock, &gameLiftSdkMock)
	return GameLiftMockHelper{
		logger:                        logger,
		logBuffer:                     &logBuffer,
		spanner:                       &spannerMock,
		gameLiftSdk:                   &gameLiftSdkMock,
		ctx:                           ctx,
		gamelift:                      gamelift,
		config:                        config,
		initialiserService:            &initialiserServiceMock,
		initialiserServiceFactoryMock: &initialiserServiceFactoryMock,
	}
}

func TestGamelift_Init_HappyPath(t *testing.T) {
	//arrange

	config := Config{
		GamePort:               100,
		Anywhere:               config2.Anywhere{},
		LogDirectory:           os.TempDir(),
		GameServerLogDirectory: os.TempDir(),
	}
	gameLiftMockHelper := createGameLiftMockHelper(&config)

	hostingInitArgs := hosting.InitArgs{
		RunId: uuid.New(),
	}

	//act
	initMetaResponse, err := gameLiftMockHelper.gamelift.Init(gameLiftMockHelper.ctx, &hostingInitArgs)

	//assert
	assert.NotNil(t, initMetaResponse)
	assert.Nil(t, err)
	switch runtime.GOOS {
	case "windows":
		assert.Equal(t, "C:\\game", initMetaResponse.InstanceWorkingDirectory)
		break
	default:
		assert.Equal(t, "/local/game", initMetaResponse.InstanceWorkingDirectory)
		break
	}
}

func TestGamelift_Init_InitSdkFailed(t *testing.T) {
	//arrange

	config := Config{
		GamePort:               100,
		Anywhere:               config2.Anywhere{},
		LogDirectory:           os.TempDir(),
		GameServerLogDirectory: os.TempDir(),
	}
	gameLiftMockHelper := createGameLiftMockHelper(&config)
	gameLiftMockHelper.initialiserService.InitSdkError = errors.New("Unit Test")

	hostingInitArgs := hosting.InitArgs{
		RunId: uuid.New(),
	}

	//act
	initMetaResponse, err := gameLiftMockHelper.gamelift.Init(gameLiftMockHelper.ctx, &hostingInitArgs)

	//assert
	assert.Nil(t, initMetaResponse)
	assert.Errorf(t, err, "failed to init gamelift: Unit Test")
}

func TestGamelift_Run_HappyPath(t *testing.T) {
	//arrange

	config := Config{
		GamePort:               100,
		Anywhere:               config2.Anywhere{},
		LogDirectory:           os.TempDir(),
		GameServerLogDirectory: os.TempDir(),
	}
	gameLiftMockHelper := createGameLiftMockHelper(&config)

	//act
	err := gameLiftMockHelper.gamelift.Run(gameLiftMockHelper.ctx)

	//assert
	assert.Nil(t, err)
	assert.True(t, gameLiftMockHelper.gameLiftSdk.ProcessReadyCalled)
	assert.Equal(t, internal.AppName(), os.Getenv(constants.EnvironmentKeySDKToolName))
	assert.Equal(t, internal.SemVer(), os.Getenv(constants.EnvironmentKeySDKToolVersion))
}

func TestGamelift_Run_HappyPath_Call_HealthCheck(t *testing.T) {
	//arrange

	config := Config{
		GamePort:               100,
		Anywhere:               config2.Anywhere{},
		LogDirectory:           os.TempDir(),
		GameServerLogDirectory: os.TempDir(),
	}
	gameLiftMockHelper := createGameLiftMockHelper(&config)
	callback := false
	gameLiftMockHelper.gamelift.SetOnHealthCheck(func(ctx context.Context) events.GameStatus {
		callback = true
		return events.GameStatusWaiting
	})

	//act
	err := gameLiftMockHelper.gamelift.Run(gameLiftMockHelper.ctx)
	assert.Nil(t, err)

	//invoke callback
	gameLiftMockHelper.gameLiftSdk.ProcessParameters.OnHealthCheck()

	//assert
	assert.True(t, callback)
}

func TestGamelift_Run_HappyPath_Call_ProcessTerminate(t *testing.T) {
	//arrange

	config := Config{
		GamePort:               100,
		Anywhere:               config2.Anywhere{},
		LogDirectory:           os.TempDir(),
		GameServerLogDirectory: os.TempDir(),
	}
	gameLiftMockHelper := createGameLiftMockHelper(&config)
	callback := false
	var hostingTerminate *events.HostingTerminate
	gameLiftMockHelper.gamelift.SetOnHostingTerminate(func(ctx context.Context, h *events.HostingTerminate) error {
		hostingTerminate = h
		callback = true
		return nil
	})

	//act
	err := gameLiftMockHelper.gamelift.Run(gameLiftMockHelper.ctx)
	assert.Nil(t, err)

	//invoke callback
	gameLiftMockHelper.gameLiftSdk.ProcessParameters.OnProcessTerminate()

	//assert
	assert.True(t, callback)
	assert.Equal(t, events.HostingTerminateReasonHostingShutdown, hostingTerminate.Reason)
}

func TestGamelift_Run_HappyPath_Call_StartGameSession(t *testing.T) {
	//arrange

	config := Config{
		GamePort:               100,
		Anywhere:               config2.Anywhere{},
		LogDirectory:           os.TempDir(),
		GameServerLogDirectory: os.TempDir(),
	}
	gameLiftMockHelper := createGameLiftMockHelper(&config)
	callback := false
	var hostingStart *events.HostingStart

	gameLiftMockHelper.gamelift.SetOnHostingStart(func(ctx context.Context, h *events.HostingStart, end <-chan error) error {
		callback = true
		hostingStart = h
		return nil
	})

	hostingInitArgs := hosting.InitArgs{
		RunId: uuid.New(),
	}

	//act
	_, err := gameLiftMockHelper.gamelift.Init(gameLiftMockHelper.ctx, &hostingInitArgs)
	assert.Nil(t, err)

	err = gameLiftMockHelper.gamelift.Run(gameLiftMockHelper.ctx)
	assert.Nil(t, err)

	//invoke callback
	gameSession := model.GameSession{
		GameSessionID:             "arn:aws:gamelift:eu-west-1::gamesession/fleet-8a8e55eb-6607-4eb3-9031-eed48907d5a4/dev-jim/6aa3a161-f2fb-4b53-bfd9-1f31c3b20cd2",
		GameSessionData:           "gameSessionData",
		Name:                      "gameSessionName",
		MatchmakerData:            "{\"matchId\":\"bcf9757e-3c61-4949-b52d-dae5996a70bb\",\"matchmakingConfigurationArn\":\"arn:aws:gamelift:us-west-2:123456789012:matchmakingconfiguration/SinglePlayerMatchmaker\"",
		FleetID:                   "FleetID",
		Location:                  "Location",
		MaximumPlayerSessionCount: 0,
		IPAddress:                 "192.168.1.1",
		Port:                      0,
		DNSName:                   "",
		GameProperties: map[string]string{
			"meta1": "alpha",
			"meta2": "beta",
			"meta3": "charlie",
		},
		Status:       nil,
		StatusReason: "",
	}
	gameLiftMockHelper.gameLiftSdk.ProcessParameters.OnStartGameSession(gameSession)

	//assert
	assert.True(t, callback)
	assert.NotNil(t, hostingStart)
	assert.Equal(t, gameSession.Name, hostingStart.GameSessionName)
	assert.Equal(t, gameSession.GameSessionData, hostingStart.GameSessionData)
	assert.Equal(t, "{\"meta1\":\"alpha\",\"meta2\":\"beta\",\"meta3\":\"charlie\"}", hostingStart.GameProperties)
	assert.Equal(t, gameSession.MatchmakerData, hostingStart.MatchmakerData)
	assert.True(t, gameLiftMockHelper.gameLiftSdk.ActivateGameSessionCalled)
	logString := gameLiftMockHelper.logBuffer.String()
	assert.Contains(t, logString, "start game sessions called")
	assert.Contains(t, logString, "manager onHostingStart")
}

func TestGamelift_OnStartGameSession_ContainerPort(t *testing.T) {
	tests := []struct {
		description           string
		fleetId               string
		configPort            int
		expectedContainerPort int
	}{
		{
			description:           "Container fleet should use config port",
			fleetId:               "containerfleet-123",
			configPort:            8080,
			expectedContainerPort: 8080,
		},
		{
			description:           "Non-container fleet should not set container port",
			fleetId:               "fleet-123",
			configPort:            8080,
			expectedContainerPort: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {

			config := Config{
				GamePort:     tt.configPort,
				LogDirectory: os.TempDir(),
			}

			gameLiftMockHelper := createGameLiftMockHelper(&config)
			var hostingStart *events.HostingStart
			callback := false

			gameLiftMockHelper.gamelift.onHostingStart = func(ctx context.Context, h *events.HostingStart, end <-chan error) error {
				hostingStart = h
				callback = true
				return nil
			}

			err := gameLiftMockHelper.gamelift.Run(gameLiftMockHelper.ctx)
			assert.Nil(t, err)

			// Create and invoke game session
			gameSession := model.GameSession{
				GameSessionID:             "test-session",
				GameSessionData:           "gameSessionData",
				Name:                      "gameSessionName",
				MatchmakerData:            "matchmakerData",
				FleetID:                   tt.fleetId,
				Location:                  "Location",
				MaximumPlayerSessionCount: 0,
				IPAddress:                 "192.168.1.1",
				Port:                      0,
				DNSName:                   "",
				GameProperties: map[string]string{
					"meta1": "alpha",
					"meta2": "beta",
				},
				Status:       nil,
				StatusReason: "",
			}
			gameLiftMockHelper.gameLiftSdk.ProcessParameters.OnStartGameSession(gameSession)

			// Assert results
			assert.True(t, callback)
			assert.NotNil(t, hostingStart)
			assert.Equal(t, tt.expectedContainerPort, hostingStart.ContainerPort)
			assert.Equal(t, gameSession.Name, hostingStart.GameSessionName)
			assert.Equal(t, gameSession.GameSessionData, hostingStart.GameSessionData)
			assert.Equal(t, gameSession.MatchmakerData, hostingStart.MatchmakerData)
			assert.Equal(t, tt.fleetId, hostingStart.FleetId)
			assert.True(t, gameLiftMockHelper.gameLiftSdk.ActivateGameSessionCalled)
			expectedGameProperties := `{"meta1":"alpha","meta2":"beta"}`
			assert.Equal(t, expectedGameProperties, hostingStart.GameProperties)
		})
	}
}
