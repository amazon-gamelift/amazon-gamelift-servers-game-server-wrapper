/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cmd

import (
	"fmt"

	"github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/internal"
	"github.com/spf13/cobra"
)

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "version",
		Long:  "Version information",
		RunE:  versionE,
	}
)

func versionE(cmd *cobra.Command, args []string) error {
	fmt.Println(internal.Version())
	return nil
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
