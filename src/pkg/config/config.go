/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

// Provider represents the hosting service provider type for the game server.
// It supports GameLift as the primary provider.
type Provider string

const (
	ProviderGameLift Provider = "gamelift"
)

// Hosting defines logging directory configurations.
type Hosting struct {
	LogDirectory                   string `mapstructure:"logDirectory" yaml:"logDirectory"`
	AbsoluteGameServerLogDirectory string `mapstructure:"gameServerLogDirectory" yaml:"gameServerLogDirectory"`
}

// ProcessInfo contains the game server process execution configuration.
type ProcessInfo struct {
	WorkingDir string `mapstructure:"-" yaml:"-"`
	ExePath    string `mapstructure:"exePath" yaml:"exePath"`
}

// GameLift configuration for Amazon GameLift service.
type GameLift struct {
	Anywhere Anywhere `mapstructure:"anywhere" yaml:"anywhere"`
	/// To be filled in by another source, not config
	Port      int `mapstructure:"-" yaml:"-"`
	QueryPort int `mapstructure:"-" yaml:"-"`
}

// AnywhereHostConfig defines the configuration for an Amazon GameLift Anywhere host.
type AnywhereHostConfig struct {
	HostName           string `mapstructure:"hostname" yaml:"hostName"`
	ServiceSdkEndpoint string `mapstructure:"serviceSdkEndpoint" yaml:"serviceSdkEndpoint"`
	AuthToken          string `mapstructure:"authToken" yaml:"authToken"`
	LocationArn        string `mapstructure:"locationArn" yaml:"locationArn"`
	FleetArn           string `mapstructure:"fleetArn" yaml:"fleetArn"`
	IPv4Address        string `mapstructure:"ipv4" yaml:"ipv4"`
}

// Anywhere defines the complete configuration for Amazon GameLift Anywhere deployment.
type Anywhere struct {
	Config AwsConfig          `mapstructure:"config" yaml:"config"`
	Host   AnywhereHostConfig `mapstructure:"host" yaml:"host"`
}

// AwsConfigProvider represents the type of AWS configuration provider to use.
// Supports profile-based and SSO-file based authentication.
type AwsConfigProvider string

const (
	AwsConfigProviderProfile AwsConfigProvider = "aws-profile"
	AwsConfigProviderSSOFile AwsConfigProvider = "sso-file"
)

// AwsConfig defines the configuration for AWS service access.
type AwsConfig struct {
	Region   string            `mapstructure:"region" yaml:"region,omitempty"`
	Provider AwsConfigProvider `mapstructure:"provider" yaml:"provider,omitempty"`
	Literal  AwsConfigLiteral  `mapstructure:"literal" yaml:"literal,omitempty"`
	Profile  string            `mapstructure:"profile" yaml:"profile,omitempty"`
	SSOFile  string            `mapstructure:"ssoFile" yaml:"ssoFile,omitempty"`
}

// AwsConfigLiteral contains direct AWS credentials.
type AwsConfigLiteral struct {
	AccessKeyId     string `mapstructure:"accessKey" yaml:"accessKeyId"`
	AccessKeySecret string `mapstructure:"secretKey" yaml:"accessKeySecret"`
}

// CliArg represents a command-line argument configuration for the game server.
type CliArg struct {
	Name     string `json:"arg" yaml:"arg" mapstructure:"arg" yaml:"arg"`
	Value    string `json:"val" yaml:"val" mapstructure:"val" yaml:"val"`
	Position int    `json:"pos" yaml:"pos" mapstructure:"pos" yaml:"pos"`
}
