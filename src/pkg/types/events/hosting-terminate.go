/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package events

// HostingTerminateReason represents the reason for a game server termination event.
type HostingTerminateReason string

const (
	HostingTerminateReasonUnspecified     HostingTerminateReason = "Unspecified"
	HostingTerminateReasonContextExpiry   HostingTerminateReason = "ContextExpired"
	HostingTerminateReasonHostingShutdown HostingTerminateReason = "HostingShutdown"
)

// HostingTerminate represents a termination event for a game server instance.
// It includes the reason for termination.
type HostingTerminate struct {
	Reason HostingTerminateReason
}
