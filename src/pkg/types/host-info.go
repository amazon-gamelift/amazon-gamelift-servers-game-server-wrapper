/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package types

import "net"

// HostInfo defines the network and identification information for a game server host.
type HostInfo struct {
	IpAddress  net.IPAddr
	Identifier string
}
