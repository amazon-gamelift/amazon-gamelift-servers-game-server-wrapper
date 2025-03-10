/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package hosting

import (
	"context"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/types/events"
	"github.com/google/uuid"
	"net/netip"
)

// IpAddresses contains the network addressing information for a hosting option.
type IpAddresses struct {
	IPv4 *netip.Addr
	IPv6 *netip.Addr
}

// InitArgs contains the initialization arguments required to start the hosting.
type InitArgs struct {
	RunId        uuid.UUID
	LogDirectory string
}

// InitMeta contains metadata about the initialized instance.
type InitMeta struct {
	InstanceWorkingDirectory string
}

// Service defines the interface for managing a game server's lifecycle
// within the Amazon GameLift environment. It provides methods for
// initialization, runtime management, and event handling.
type Service interface {
	Init(ctx context.Context, args *InitArgs) (*InitMeta, error)
	Run(ctx context.Context) error
	SetOnHostingStart(f func(ctx context.Context, h *events.HostingStart, end <-chan error) error)
	SetOnHostingTerminate(f func(ctx context.Context, h *events.HostingTerminate) error)
	SetOnHealthCheck(f func(ctx context.Context) events.GameStatus)
	Close(ctx context.Context) error
}
