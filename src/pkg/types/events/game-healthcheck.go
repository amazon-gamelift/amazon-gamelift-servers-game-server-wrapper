/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package events

// GameStatus represents the current operational state of a game server.
type GameStatus string

const (
	GameStatusWaiting    GameStatus = "waiting"
	GameStatusRunning    GameStatus = "running"
	GameStatusErrored    GameStatus = "errored"
	GameStatusTerminated GameStatus = "terminated"
	GameStatusFinished   GameStatus = "finished"
)
