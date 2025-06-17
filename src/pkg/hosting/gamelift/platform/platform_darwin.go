//go:build darwin

/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package platform

func InstancePath() string {
	return "/local/game"
}
