/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package initialiser

import (
	"aws/amazon-gamelift-go-sdk/common"
	"aws/amazon-gamelift-go-sdk/server"
	"context"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/config"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/hosting/gamelift/client"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/hosting/gamelift/sdk"
	"github.com/aws/aws-sdk-go-v2/service/gamelift"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"log/slog"
	"os"
	"regexp"
	"strings"
)

type anywhere struct {
	processId          string
	hostname           string
	serviceSdkEndpoint string
	authToken          string
	fleetId            string
	location           string
	clientProvider     client.Provider
	sdk                sdk.GameLiftSdk
	cfg                *config.Anywhere
	logger             *slog.Logger
}

// InitSdk initializes the GameLift SDK for Anywhere fleet usage.
//
// Parameters:
//   - ctx: Context for initialization
//
// Returns:
//   - error: Any error encountered during initialization
func (anywhere *anywhere) InitSdk(ctx context.Context) error {
	anywhere.logger.InfoContext(ctx, "using Amazon GameLift Anywhere", "version", common.SdkVersion)

	var glClient client.GameLift
	var err error

	wssEndpoint := anywhere.serviceSdkEndpoint
	if len(wssEndpoint) != 0 {
		anywhere.logger.DebugContext(ctx, "using pre-set Anywhere configuration", "computeName", anywhere.hostname, "serviceSdkEndpoint", wssEndpoint)
	} else {
		anywhere.logger.DebugContext(ctx, "listing compute")

		glClient, err = anywhere.clientProvider.GetGameLift(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to get Amazon GameLift client provider")
		}

		computes, err := glClient.ListCompute(ctx, &gamelift.ListComputeInput{
			FleetId: &anywhere.fleetId,
		})

		if err != nil {
			return errors.Wrap(err, "failed to list compute")
		}

		for i := range computes.ComputeList {
			if c := computes.ComputeList[i]; c.ComputeName != nil && *c.ComputeName == anywhere.hostname {
				wssEndpoint = *c.GameLiftServiceSdkEndpoint
				anywhere.logger.DebugContext(ctx, "found compute", "ComputeName", &anywhere.hostname, "serviceSdkEndpoint", wssEndpoint)
				break
			}
		}

		if len(wssEndpoint) == 0 {
			anywhere.logger.DebugContext(ctx, "registering compute", "ComputeName", &anywhere.hostname)
			compute, err := glClient.RegisterCompute(ctx, &gamelift.RegisterComputeInput{
				ComputeName: &anywhere.hostname,
				FleetId:     &anywhere.fleetId,
				IpAddress:   &anywhere.cfg.Host.IPv4Address,
				Location:    &anywhere.location,
			})

			if err != nil {
				return errors.Wrap(err, "failed to register initialiser")
			}

			wssEndpoint = *compute.Compute.GameLiftServiceSdkEndpoint
		}
	}

	authToken := anywhere.authToken
	if len(authToken) != 0 {
		anywhere.logger.DebugContext(ctx, "using pre-set Anywhere configuration", "authToken", authToken)
	} else {
		anywhere.logger.DebugContext(ctx, "getting auth token")

		if glClient == nil {
			glClient, err = anywhere.clientProvider.GetGameLift(ctx)
			if err != nil {
				return errors.Wrap(err, "failed to get Amazon GameLift client provider")
			}
		}

		token, err := glClient.GetComputeAuthToken(ctx, &gamelift.GetComputeAuthTokenInput{
			ComputeName: &anywhere.hostname,
			FleetId:     &anywhere.fleetId,
		})
		if err != nil {
			return errors.Wrap(err, "failed to get Amazon GameLift Anywhere token")
		}
		authToken = *token.AuthToken
	}

	params := server.ServerParameters{
		HostID:       anywhere.hostname,
		FleetID:      anywhere.fleetId,
		AuthToken:    authToken,
		ProcessID:    anywhere.processId,
		WebSocketURL: wssEndpoint,
	}

	anywhere.logger.DebugContext(ctx, "initialising sdk")
	if err := anywhere.sdk.InitSDK(ctx, params); err != nil {
		return errors.Wrap(err, "failed to init sdk for anywhere")
	}

	return nil
}

func getHostname() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", errors.Wrap(err, "failed to get machine hostname...")
	}

	// Ensure that hostname complies with the string requirements for RegisterCompute
	hostname = regexp.MustCompile("[^a-zA-Z0-9\\-]").ReplaceAllString(hostname, "")
	if len(hostname) > 128 {
		hostname = hostname[:128]
	}

	return hostname, nil
}

func newAnywhere(ctx context.Context, cfg *config.Anywhere, gl sdk.GameLiftSdk, logger *slog.Logger, clientProvider client.Provider) (Service, error) {
	var err error

	processId := common.GetEnvStringOrDefault(common.EnvironmentKeyProcessID, uuid.New().String())
	serviceSdkEndpoint := common.GetEnvStringOrDefault(common.EnvironmentKeyWebsocketURL, cfg.Host.ServiceSdkEndpoint)
	authToken := common.GetEnvStringOrDefault(common.EnvironmentKeyAuthToken, cfg.Host.AuthToken)
	hostname := common.GetEnvStringOrDefault(common.EnvironmentKeyHostID, cfg.Host.HostName)

	if len(hostname) == 0 {
		if hostname, err = getHostname(); err != nil {
			return nil, err
		}
	}

	fleetParts := strings.Split(cfg.Host.FleetArn, "/")
	if len(fleetParts) != 2 {
		return nil, errors.Errorf("invalid fleet arn '%s'", cfg.Host.FleetArn)
	}
	fleetId := fleetParts[1]

	locationParts := strings.Split(cfg.Host.LocationArn, "/")
	if len(locationParts) != 2 {
		return nil, errors.Errorf("invalid location arn '%s'", cfg.Host.LocationArn)
	}
	location := locationParts[1]

	a := &anywhere{
		processId:          processId,
		hostname:           hostname,
		serviceSdkEndpoint: serviceSdkEndpoint,
		authToken:          authToken,
		fleetId:            fleetId,
		location:           location,
		clientProvider:     clientProvider,
		sdk:                gl,
		cfg:                cfg,
		logger:             logger,
	}

	return a, nil
}
