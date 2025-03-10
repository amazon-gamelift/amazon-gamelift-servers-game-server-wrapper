/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package types

// GameState represents the current state of a game server.
type GameState string

const (
	GameStateInitialising  GameState = "initialising"
	GameStateStartingGame  GameState = "starting"
	GameStateRunning       GameState = "running"
	GameStateFinishingGame GameState = "finishing"
	GameStateShuttingDown  GameState = "shutting-down"
	GameStateFailed        GameState = "failed"
)
