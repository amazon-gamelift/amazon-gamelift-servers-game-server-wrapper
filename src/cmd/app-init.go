/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cmd

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/constants"
	"github.com/spf13/cobra"
)

// appInit initializes the application's core components including logging
// and working directory configuration. It sets up the application environment
// for proper execution.
func appInit() {
	// Initialize structured JSON logging with timestamp and level
	logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// goland refuses to set the working directory correctly
	if environmentStartDir, exists := os.LookupEnv(constants.EnvironmentKeyStartDir); exists {
		appDir = environmentStartDir
	} else {
		programFile := os.Args[0]
		exeLocation, err := filepath.Abs(programFile)
		if err != nil {
			logger.Error("Failed to get absolute path", "programFile", programFile, "error", err)
			os.Exit(1)
		}
		appDir = filepath.Dir(exeLocation)
	}

	cobra.OnInitialize(initConfig)
}
