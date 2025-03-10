/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package internal

import "fmt"

var (
	appName = "GameServerWrapper"
	version = "DEV"
)

func AppName() string { return appName }

func Version() string {
	return fmt.Sprintf("%s v%s", appName, version)
}

func SemVer() string {
	return version
}
