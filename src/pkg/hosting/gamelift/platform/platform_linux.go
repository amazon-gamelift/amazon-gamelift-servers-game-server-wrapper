/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

//go:build linux

package platform

func InstancePath() string {
	return "/local/game"
}
