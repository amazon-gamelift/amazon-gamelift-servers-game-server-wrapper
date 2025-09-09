/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package initialiser

import (
	"context"
	"log/slog"

	"github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/pkg/config"
	"github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/pkg/hosting/gamelift/client"
	"github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/pkg/hosting/gamelift/sdk"
	"github.com/pkg/errors"
)

// Service defines the interface for hosting option initialization and management operations.
type Service interface {
	// InitSdk initializes the GameLift SDK for the server environment.
	InitSdk(ctx context.Context) error
}

var (
	anywhereNew = newAnywhere
	managedNew  = newManaged
)

type InitialiserServiceFactory struct {
}

// GetService creates and returns a hosting service based on the provided configuration.
//
// Parameters:
//   - ctx: Context for service creation
//   - anywhere: GameLift Anywhere configuration
//   - gameLiftSdk: GameLift SDK instance
//   - logger: Logger for operations
//
// Returns:
//   - Service: Configured hosting service
//   - error: Any error encountered during service creation
func (initialiserServiceFactory *InitialiserServiceFactory) GetService(ctx context.Context, anywhere config.Anywhere, gameLiftSdk sdk.GameLiftSdk, logger *slog.Logger) (Service, error) {

	if len(anywhere.Host.FleetArn) != 0 {
		var clientProvider client.Provider
		var err error
		// Only initialize the clientProvider if either ServiceSdkEndpoint or AuthToken are undefined
		if len(anywhere.Host.ServiceSdkEndpoint) == 0 || len(anywhere.Host.AuthToken) == 0 {
			clientProvider, err = client.NewProvider(anywhere, logger)
			if err != nil {
				return nil, err
			}

			if err = clientProvider.Init(ctx); err != nil {
				return nil, errors.Wrap(err, "failed to initialise client")
			}
		}

		return anywhereNew(ctx, &anywhere, gameLiftSdk, logger, clientProvider)
	}

	return managedNew(gameLiftSdk, logger), nil
}
