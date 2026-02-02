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

	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/config"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/observability"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/go-playground/validator/v10"
)

// Config represents the main configuration structure for the game server wrapper.
type Config struct {
	Observability observability.Config `mapstructure:"observability" yaml:"observability, omitempty"`
	LogLevel      string               `mapstructure:"logLevel" yaml:"logLevel"`
	BuildDetail   BuildDetail          `mapstructure:"buildDetail" yaml:"buildDetail" json:"buildDetail"`
	Ports         Ports                `mapstructure:"ports" yaml:"ports, omitempty"`
	Route53       Route53              `mapstructure:"route53" yaml:"route53"`
	Orchestration Orchestration        `mapstructure:"orchestration" yaml:"orchestration"`
	Hosting       Hosting              `mapstructure:"hosting" yaml:"hosting"`
	Datadog       Datadog              `mapstructure:"datadog" yaml:"datadog,omitempty"`
}

// ConfigWrapper provides a wrapper configuration structure for additional
// server configuration options, particularly for Amazon GameLift Anywhere setup.
type ConfigWrapper struct {
	LogConfig                  LogConfig         `mapstructure:"log-config" yaml:"log-config"`
	Anywhere                   AnywhereConfig    `mapstructure:"anywhere" yaml:"anywhere"`
	Ports                      Ports             `mapstructure:"ports" yaml:"ports, omitempty"`
	Route53                    Route53           `mapstructure:"route53" yaml:"route53"`
	Orchestration              Orchestration     `mapstructure:"orchestration" yaml:"orchestration"`
	GameServerDetails          GameServerDetails `mapstructure:"game-server-details" yaml:"game-server-details"`
	InjectFleetRoleCredentials bool              `mapstructure:"inject-fleet-role-credentials" yaml:"inject-fleet-role-credentials"`
	FleetRoleArn               string            `mapstructure:"fleet-role-arn" yaml:"fleet-role-arn"`
	FleetRoleSessionName       string            `mapstructure:"fleet-role-session-name" yaml:"fleet-role-session-name"`
	Datadog                    Datadog           `mapstructure:"datadog" yaml:"datadog,omitempty"`
}

// LogConfig defines logging-specific configuration options.
type LogConfig struct {
	WrapperLogLevel   string `mapstructure:"wrapper-log-level" yaml:"wrapper-log-level"`
	GameServerLogsDir string `mapstructure:"game-server-logs-dir" yaml:"game-server-logs-dir"`
	DelayStart        string `mapstructure:"delay-start" yaml:"delay-start"`
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
	WorkingDir         string          `mapstructure:"working-directory" yaml:"working-directory"`
	DelayStart         string          `mapstructure:"delay-start" yaml:"delay-start"`
	ExecutableFilePath string          `mapstructure:"executable-file-path" yaml:"executable-file-path"`
	GameServerArgs     []config.CliArg `mapstructure:"game-server-args" yaml:"game-server-args"`
	GameServerEnvVars  []config.EnvVar `mapstructure:"game-server-env-vars" yaml:"game-server-env-vars"`
}

// BuildDetail contains information about the game server build and execution environment.
type BuildDetail struct {
	WorkingDir      string          `mapstructure:"workingDirectory" yaml:"workingDirectory"`
	RelativeExePath string          `mapstructure:"exePath" yaml:"exePath"`
	DefaultArgs     []config.CliArg `mapstructure:"defaultArgs" yaml:"defaultArgs"`
	EnvVars         []config.EnvVar `mapstructure:"envVars" yaml:"envVars"`
	DelayStart      string          `mapstructure:"delayStart" yaml:"delayStart"`
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

	var absWorkingDir string
	var err error

	if configWrapper.GameServerDetails.WorkingDir != "" {
		absWorkingDir, err = filepath.Abs(configWrapper.GameServerDetails.WorkingDir)
		if err != nil {
			return fmt.Errorf("error getting absolute working directory: %v", err)
		}
	} else {
		absWorkingDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("error getting current directory: %v", err)
		}
		absWorkingDir, err = filepath.Abs(absWorkingDir)
		if err != nil {
			return fmt.Errorf("error getting absolute path: %v", err)
		}
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
		EnvVars:         configWrapper.GameServerDetails.GameServerEnvVars,
		DelayStart:      configWrapper.GameServerDetails.DelayStart,
	}
	cfg.Ports = Ports{
		GamePort: configWrapper.Ports.GamePort,
	}
	cfg.Route53 = Route53{
		DoMapping:           configWrapper.Route53.DoMapping,
		HostDomain:          configWrapper.Route53.HostDomain,
		Type:                configWrapper.Route53.Type,
		Ttl:                 configWrapper.Route53.Ttl,
		Comment:             configWrapper.Route53.Comment,
		PublicHostedZoneId:  configWrapper.Route53.PublicHostedZoneId,
		PrivateHostedZoneId: configWrapper.Route53.PrivateHostedZoneId,
		TokenUrl:            configWrapper.Route53.TokenUrl,
		TokenHeaderKey:      configWrapper.Route53.TokenHeaderKey,
		TokenHeaderValue:    configWrapper.Route53.TokenHeaderValue,
		MetaDataUrl:         configWrapper.Route53.MetaDataUrl,
		MetaDataHeaderKey:   configWrapper.Route53.MetaDataHeaderKey,
		PublicIpUrl:         configWrapper.Route53.PublicIpUrl,
		PrivateIpUrl:        configWrapper.Route53.PrivateIpUrl,
		Region:              configWrapper.Route53.Region,
		Profile:             configWrapper.Route53.Profile,
	}
	cfg.Orchestration = Orchestration{
		AuthHeaderPrefix: configWrapper.Orchestration.AuthHeaderPrefix,
		EmitCustomEvents: configWrapper.Orchestration.EmitCustomEvents,
		Account:          configWrapper.Orchestration.Account,
		Resources:        configWrapper.Orchestration.Resources,
		Method:           configWrapper.Orchestration.Method,
		Url:              configWrapper.Orchestration.Url,
		HeaderKey:        configWrapper.Orchestration.HeaderKey,
		HeaderValue:      configWrapper.Orchestration.HeaderValue,
		GetTokenUrl:      configWrapper.Orchestration.GetTokenUrl,
		ClientId:         configWrapper.Orchestration.ClientId,
		ClientSecret:     configWrapper.Orchestration.ClientSecret,
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
			InjectFleetRoleCredentials: configWrapper.InjectFleetRoleCredentials,
			FleetRoleArn:               configWrapper.FleetRoleArn,
			FleetRoleSessionName:       configWrapper.FleetRoleSessionName,
		},
	}
	cfg.Datadog = Datadog{
		Enabled:    configWrapper.Datadog.Enabled,
		ConfigPath: configWrapper.Datadog.ConfigPath,
		Tags:       configWrapper.Datadog.Tags,
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

// Route53 defines all Route53 related configuration settings.
type Route53 struct {
	DoMapping           bool   `mapstructure:"doMapping" json:"doMapping" yaml:"doMapping"`
	HostDomain          string `mapstructure:"hostDomain" json:"hostDomain,omitempty" yaml:"hostDomain,omitempty"`
	Type                string `mapstructure:"type" json:"type,omitempty" yaml:"type,omitempty"`
	Ttl                 int64  `mapstructure:"ttl" json:"ttl,omitempty" yaml:"ttl,omitempty"`
	Comment             string `mapstructure:"comment" json:"comment,omitempty" yaml:"comment,omitempty"`
	PublicHostedZoneId  string `mapstructure:"publicHostedZoneId" json:"publicHostedZoneId,omitempty" yaml:"publicHostedZoneId,omitempty"`
	PrivateHostedZoneId string `mapstructure:"privateHostedZoneId" json:"privateHostedZoneId,omitempty" yaml:"privateHostedZoneId,omitempty"`
	TokenUrl            string `mapstructure:"tokenUrl" json:"tokenUrl,omitempty" yaml:"tokenUrl,omitempty"`
	TokenHeaderKey      string `mapstructure:"tokenHeaderKey" json:"tokenHeaderKey,omitempty" yaml:"tokenHeaderKey,omitempty"`
	TokenHeaderValue    string `mapstructure:"tokenHeaderValue" json:"tokenHeaderValue,omitempty" yaml:"tokenHeaderValue,omitempty"`
	MetaDataUrl         string `mapstructure:"metaDataUrl" json:"metaDataUrl,omitempty" yaml:"metaDataUrl,omitempty"`
	MetaDataHeaderKey   string `mapstructure:"metaDataHeaderKey" json:"metaDataHeaderKey,omitempty" yaml:"metaDataHeaderKey,omitempty"`
	PublicIpUrl         string `mapstructure:"publicIpUrl" json:"publicIpUrl,omitempty" yaml:"publicIpUrl,omitempty"`
	PrivateIpUrl        string `mapstructure:"privateIpUrl" json:"privateIpUrl,omitempty" yaml:"privateIpUrl,omitempty"`
	Region              string `mapstructure:"region" json:"region,omitempty" yaml:"region,omitempty"`
	Profile             string `mapstructure:"profile" json:"profile,omitempty" yaml:"profile,omitempty"`
}

// Orchestration defines all configuration settings related to the Orchestration Service.
type Orchestration struct {
	AuthHeaderPrefix string   `mapstructure:"authHeaderPrefix" yaml:"authHeaderPrefix"`
	EmitCustomEvents bool     `mapstructure:"emitCustomEvents" yaml:"emitCustomEvents"`
	Account          string   `mapstructure:"account" yaml:"account"`
	Resources        []string `mapstructure:"resources" yaml:"resources"`
	Method           string   `mapstructure:"method" yaml:"method"`
	Url              string   `mapstructure:"url" yaml:"url"`
	HeaderKey        string   `mapstructure:"headerKey" yaml:"headerKey"`
	HeaderValue      string   `mapstructure:"headerValue" yaml:"headerValue"`
	GetTokenUrl      string   `mapstructure:"getTokenUrl" yaml:"getTokenUrl"`
	ClientId         string   `mapstructure:"clientId" yaml:"clientId"`
	ClientSecret     string   `mapstructure:"clientSecret" yaml:"clientSecret"`
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

// Datadog defines the configuration for datadog agent integration.
type Datadog struct {
	Enabled    bool              `mapstructure:"enabled" yaml:"enabled"`
	ConfigPath string            `mapstructure:"config-path" yaml:"config-path"`
	Tags       map[string]string `mapstructure:"tags" yaml:"tags"`
}
