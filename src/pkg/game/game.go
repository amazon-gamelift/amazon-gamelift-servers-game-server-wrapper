/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package game

import (
	"context"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/types/events"
	"github.com/google/uuid"
)

// StartArgs contains the arguments required to start a game server instance.
type StartArgs struct {
	*events.HostingStart
}

// InitMeta contains metadata returned after successful game server initialization.
type InitMeta struct {
}

// InitArgs contains the arguments required for game server initialization.
type InitArgs struct {
	RunId uuid.UUID
}

// Server defines the interface that need to be implemented by game servers.
// It provides the contract for the complete lifecycle of a game server instance.
type Server interface {
	// Init initializes the game server with the provided initialization arguments.
	//
	// Parameters:
	//   - ctx: Context for initialization operation
	//   - args: Initialization arguments including RunId
	//
	// Returns:
	//   - *InitMeta: Initialization metadata
	//   - error: Any error that occurred during initialization
	Init(ctx context.Context, args *InitArgs) (*InitMeta, error)

	// Run starts the game server with the provided start arguments.
	//
	// Parameters:
	//   - ctx: Context for the server execution
	//   - args: Start arguments including hosting parameters
	//
	// Returns:
	//   - error: Any error that occurred during server execution
	Run(ctx context.Context, args *StartArgs) error

	// HealthCheck performs a health check of the game server, it returns the current status of the game server.
	//
	// Parameters:
	//   - ctx: Context for the health check operation
	//
	// Returns:
	//   - events.GameStatus: Current status of the game server
	HealthCheck(ctx context.Context) events.GameStatus

	// Stop initiates a graceful shutdown of the game server.
	//
	// Parameters:
	//   - ctx: Context for the stop operation
	//
	// Returns:
	//   - error: Any error that occurred during shutdown
	Stop(ctx context.Context) error
}
