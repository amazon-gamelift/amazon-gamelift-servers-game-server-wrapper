/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cmd

import (
	"log/slog"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type argError struct {
	err error
}

func (e argError) Error() string {
	return e.err.Error()
}

func bindFlags() {
	rootCmd.SetFlagErrorFunc(func(command *cobra.Command, err error) error {
		return &argError{err: err}
	})

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is ./config.yaml)")

	rootCmd.PersistentFlags().StringVarP(&cfgWrapper.LogConfig.WrapperLogLevel, "wrapper-log-level", "v", "warn", "log level, valid values are debug,info,warn,error")
	err := viperInstance.BindPFlag("log-config.wrapper-log-level", rootCmd.PersistentFlags().Lookup("wrapper-log-level"))
	if err != nil {
		logger.Error("Failed to bind flag to viper", "flag", "wrapper-log-level", "error", err)
	}

	rootCmd.PersistentFlags().IntVarP(&cfgWrapper.Ports.GamePort, "port", "p", 0, "game port")
	err = viperInstance.BindPFlag("ports.gamePort", rootCmd.PersistentFlags().Lookup("port"))
	if err != nil {
		logger.Error("Failed to bind flag to viper", "flag", "port", "error", err)
	}

	rootCmd.PersistentFlags().StringVarP(&cfgWrapper.Anywhere.FleetArn, "fleet-arn", "f", "", "anywhere fleet arn")
	err = viperInstance.BindPFlag("anywhere.fleet-arn", rootCmd.PersistentFlags().Lookup("fleet-arn"))
	if err != nil {
		logger.Error("Failed to bind flag to viper", "flag", "fleet-arn", "error", err)
	}

	rootCmd.PersistentFlags().StringVarP(&cfgWrapper.Anywhere.LocationArn, "location-arn", "l", "", "anywhere fleet location arn")
	err = viperInstance.BindPFlag("anywhere.location-arn", rootCmd.PersistentFlags().Lookup("location-arn"))
	if err != nil {
		logger.Error("Failed to bind flag to viper", "flag", "location", "error", err)
	}

	rootCmd.PersistentFlags().StringVarP(&cfgWrapper.Anywhere.AuthToken, "auth-token", "t", "", "anywhere fleet auth token")
	err = viperInstance.BindPFlag("anywhere.auth-token", rootCmd.PersistentFlags().Lookup("auth-token"))
	if err != nil {
		logger.Error("Failed to bind flag to viper", "flag", "auth-token", "error", err)
	}
}

func parseLogLevel(s string) (slog.Level, error) {
	l := strings.TrimSpace(strings.ToLower(s))
	switch l {
	case "d", "dbg", "debug":
		return slog.LevelDebug, nil
	case "i", "inf", "info":
		return slog.LevelInfo, nil
	case "w", "wrn", "warn":
		return slog.LevelWarn, nil
	case "e", "err", "error":
		return slog.LevelError, nil
	default:
		return -1, errors.Errorf("unknown log level %s", s)
	}
}
