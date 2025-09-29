/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package events

import (
	"github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/pkg/config"
)

// HostingStart represents the initialization configuration for a game server instance.
type HostingStart struct {
	CliArgs                   []config.CliArg
	ContainerPort             int
	DNSName                   string
	FleetId                   string
	GamePort                  int
	GameProperties            map[string]string
	GameSessionData           string
	GameSessionId             string
	GameSessionName           string
	IpAddress                 string
	LogDirectory              string
	MatchmakerData            string
	MaximumPlayerSessionCount int
	Provider                  config.Provider
}
