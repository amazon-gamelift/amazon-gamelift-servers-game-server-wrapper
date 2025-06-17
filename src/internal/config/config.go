/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/go-playground/validator/v10"

	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/config"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/observability"
)

// Config represents the main configuration structure for the game server wrapper.
type Config struct {
	Observability observability.Config `mapstructure:"observability" yaml:"observability, omitempty"`
	LogLevel      string               `mapstructure:"logLevel" yaml:"logLevel"`
	BuildDetail   BuildDetail          `mapstructure:"buildDetail" yaml:"buildDetail" json:"buildDetail"`
	Ports         Ports                `mapstructure:"ports" yaml:"ports, omitempty"`
	Hosting       Hosting              `mapstructure:"hosting" yaml:"hosting"`
}

// ConfigWrapper provides a wrapper configuration structure for additional
// server configuration options, particularly for Amazon GameLift Anywhere setup.
type ConfigWrapper struct {
	LogConfig         LogConfig         `mapstructure:"log-config" yaml:"log-config"`
	Anywhere          AnywhereConfig    `mapstructure:"anywhere" yaml:"anywhere"`
	Ports             Ports             `mapstructure:"ports" yaml:"ports, omitempty"`
	GameServerDetails GameServerDetails `mapstructure:"game-server-details" yaml:"game-server-details"`
}

// LogConfig defines logging-specific configuration options.
type LogConfig struct {
	WrapperLogLevel   string `mapstructure:"wrapper-log-level" yaml:"wrapper-log-level"`
	GameServerLogsDir string `mapstructure:"game-server-logs-dir" yaml:"game-server-logs-dir"`
}

// AnywhereConfig defines Amazon GameLift Anywhere specific configuration settings.
type AnywhereConfig struct {
	Profile            string                   `mapstructure:"profile" yaml:"profile"`
	Provider           config.AwsConfigProvider `mapstructure:"provider" yaml:"provider"`
	ComputeName        string                   `mapstructure:"compute-name" yaml:"compute-name"`
	ServiceSdkEndpoint string                   `mapstructure:"service-sdk-endpoint" yaml:"service-sdk-endpoint"`
	AuthToken          string                   `mapstructure:"auth-token" yaml:"auth-token"`
	LocationArn        string                   `mapstructure:"location-arn" yaml:"location-arn"`
	FleetArn           string                   `mapstructure:"fleet-arn" yaml:"fleet-arn"`
	IPv4               string                   `mapstructure:"ipv4" yaml:"ipv4"`
}

// GameServerDetails contains configuration details for the game server executable.
type GameServerDetails struct {
	ExecutableFilePath string          `mapstructure:"executable-file-path" yaml:"executable-file-path"`
	GameServerArgs     []config.CliArg `mapstructure:"game-server-args" yaml:"game-server-args"`
}

// BuildDetail contains information about the game server build and execution environment.
type BuildDetail struct {
	WorkingDir      string          `mapstructure:"workingDirectory" yaml:"workingDirectory"`
	RelativeExePath string          `mapstructure:"exePath" yaml:"exePath"`
	DefaultArgs     []config.CliArg `mapstructure:"defaultArgs" yaml:"defaultArgs"`
}

// Validate performs validation of the Config structure.
// It ensures all required fields are properly set and contain valid values.
//
// Returns:
//   - error: An error if validation fails, nil if validation succeeds
func (cfg Config) Validate() error {
	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return fmt.Errorf("error validating config: %w", err)
	}
	return nil
}

func makePathRelative(basePath, targetPath string) (string, error) {
	// Clean and convert both paths to use OS-specific separators
	basePath = filepath.Clean(basePath)
	targetPath = filepath.Clean(targetPath)

	// Get absolute paths
	absBase, err := filepath.Abs(basePath)
	if err != nil {
		return "", fmt.Errorf("error getting absolute base path: %v", err)
	}

	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return "", fmt.Errorf("error getting absolute target path: %v", err)
	}

	// Get relative path
	relPath, err := filepath.Rel(absBase, absTarget)
	if err != nil {
		return "", fmt.Errorf("error getting relative path: %v", err)
	}

	// Convert to forward slashes for consistency
	relPath = filepath.ToSlash(relPath)

	// Ensure it starts with ./
	if !strings.HasPrefix(relPath, ".") {
		relPath = "./" + relPath
	}

	return relPath, nil
}

func makeAbsolutePath(basePath, targetPath string) (string, error) {
	// If targetPath is empty, return empty
	if targetPath == "" {
		return "", nil
	}

	// Clean the target path
	targetPath = filepath.Clean(targetPath)

	// If it's already absolute, just convert separators and return
	if filepath.IsAbs(targetPath) {
		return filepath.ToSlash(targetPath), nil
	}

	// If relative, make it absolute based on basePath
	absPath, err := filepath.Abs(filepath.Join(basePath, targetPath))
	if err != nil {
		return "", fmt.Errorf("error getting absolute path: %v", err)
	}

	// Convert to forward slashes and return
	return filepath.ToSlash(absPath), nil
}

// AdaptConfigWrapperToConfig converts a ConfigWrapper instance to a Config instance,
// handling path conversions and establishing the proper directory structure for the game server.
//
// Parameters:
//   - configWrapper: *ConfigWrapper - Source configuration containing raw settings
//   - cfg: *Config - Destination configuration to be populated with adapted settings
//
// Returns:
//   - error: An error if any of the following operations fail:
//   - Getting current working directory
//   - Converting paths to absolute/relative format
//   - Path validation
//   - Directory access
func AdaptConfigWrapperToConfig(configWrapper *ConfigWrapper, cfg *Config) error {
	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %v", err)
	}

	// Get absolute working directory
	absWorkingDir, err := filepath.Abs(currentDir)
	if err != nil {
		return fmt.Errorf("error getting absolute path: %v", err)
	}

	// Clean working directory path
	absWorkingDir = filepath.Clean(absWorkingDir)

	// Make ExePath relative to WorkingDir
	relExePath, err := makePathRelative(absWorkingDir, configWrapper.GameServerDetails.ExecutableFilePath)
	if err != nil {
		return fmt.Errorf("error making executable path relative: %v", err)
	}

	// Make game server logs directory absolute if it's relative
	gameServerLogsDir, err := makeAbsolutePath(absWorkingDir, configWrapper.LogConfig.GameServerLogsDir)
	if err != nil {
		return fmt.Errorf("error making game server logs path absolute: %v", err)
	}

	anywhereAwsRegion, err := getAnywhereRegionAndValidate(&configWrapper.Anywhere)
	if err != nil {
		return fmt.Errorf("error validating anywhere config: %v", err)
	}

	cfg.LogLevel = configWrapper.LogConfig.WrapperLogLevel
	cfg.BuildDetail = BuildDetail{
		WorkingDir:      absWorkingDir,
		RelativeExePath: relExePath,
		DefaultArgs:     configWrapper.GameServerDetails.GameServerArgs,
	}
	cfg.Ports = Ports{
		GamePort: configWrapper.Ports.GamePort,
	}
	cfg.Hosting = Hosting{
		Hosting: config.Hosting{
			LogDirectory:                   absWorkingDir,
			AbsoluteGameServerLogDirectory: gameServerLogsDir,
		},
		GameLift: config.GameLift{
			Anywhere: config.Anywhere{
				Config: config.AwsConfig{
					Region:   anywhereAwsRegion,
					Provider: configWrapper.Anywhere.Provider,
					Profile:  configWrapper.Anywhere.Profile,
				},
				Host: config.AnywhereHostConfig{
					HostName:           configWrapper.Anywhere.ComputeName,
					ServiceSdkEndpoint: configWrapper.Anywhere.ServiceSdkEndpoint,
					AuthToken:          configWrapper.Anywhere.AuthToken,
					LocationArn:        configWrapper.Anywhere.LocationArn,
					FleetArn:           configWrapper.Anywhere.FleetArn,
					IPv4Address:        configWrapper.Anywhere.IPv4,
				},
			},
		},
	}

	return nil
}

func getAnywhereRegionAndValidate(anywhereConfig *AnywhereConfig) (string, error) {
	if anywhereConfig == nil {
		return "", nil
	}

	locationArnDefined := anywhereConfig.LocationArn != ""
	fleetArnDefined := anywhereConfig.FleetArn != ""
	ipv4Defined := anywhereConfig.IPv4 != ""
	if (locationArnDefined != fleetArnDefined) || (locationArnDefined != ipv4Defined) {
		return "", fmt.Errorf("anywhere.location-arn, anywhere.fleet-arn, and anywhere.ipv4 must be either all empty or all non-empty")
	}

	if (anywhereConfig.ComputeName == "") != (anywhereConfig.ServiceSdkEndpoint == "") {
		return "", fmt.Errorf("anywhere.compute-name and anywhere.service-sdk-endpoint must be either both empty or both non-empty")
	}

	if (anywhereConfig.AuthToken != "") && (anywhereConfig.ComputeName == "") {
		return "", fmt.Errorf("auth-token can only be provided when anywhere.compute-name is provided")
	}

	if locationArnDefined {
		locationRegion, err := getRegionFromArn(anywhereConfig.LocationArn)
		if err != nil {
			return "", fmt.Errorf("error getting region from location-arn: %v", err)
		}
		fleetRegion, err := getRegionFromArn(anywhereConfig.FleetArn)
		if err != nil {
			return "", fmt.Errorf("error getting region from fleet-arn: %v", err)
		}
		if locationRegion != fleetRegion {
			return "", fmt.Errorf("location-arn and fleet-arn must be in the same region")
		}
		return locationRegion, nil
	}

	return "", nil
}

func getRegionFromArn(arnStr string) (string, error) {
	parsedArn, err := arn.Parse(arnStr)
	if err != nil {
		return "", err
	}
	return parsedArn.Region, nil
}

// Ports defines the network port configuration for the game server.
type Ports struct {
	GamePort int `mapstructure:"gamePort" json:"gamePort" yaml:"gamePort, omitempty" validate:"required,gt=0"`
}

// Hosting defines all hosting-related configuration settings.
type Hosting struct {
	config.Hosting `mapstructure:",squash"`
	GameLift       config.GameLift `mapstructure:"gamelift" yaml:"gameLift"`
}

// Game defines the game process configuration and its launch parameters.
type Game struct {
	config.ProcessInfo `mapstructure:",squash" `
	DefaultArgs        []config.CliArg `mapstructure:"defaultArgs"`
}

// Scryer defines the configuration for the monitoring component.
type Scryer struct {
	config.ProcessInfo `mapstructure:",squash"`
	Interval           time.Duration `mapstructure:"interval"`
	LogLevel           string        `mapstructure:"logLevel"`
	Output             string        `mapstructure:"output"`
}

// Client defines the configuration for game client connections.
type Client struct {
	GamePort  int  `mapstructure:"gamePort"`
	QueryPort int  `mapstructure:"queryPort"`
	Game      Game `mapstructure:"game"`
}
