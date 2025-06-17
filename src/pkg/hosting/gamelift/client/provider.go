/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"
	"fmt"
	"log/slog"

	pkgConfig "github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	gl "github.com/aws/aws-sdk-go-v2/service/gamelift"
	"github.com/aws/smithy-go/logging"
	"github.com/pkg/errors"
)

var _ Provider = (*provider)(nil)

type Provider interface {
	Init(ctx context.Context) error
	GetAwsConfig(ctx context.Context) (aws.Config, error)
	GetGameLift(ctx context.Context) (GameLift, error)
}

type ConfigProvider interface {
	Init(ctx context.Context) error
	GetOpts(ctx context.Context) ([]func(*awscfg.LoadOptions) error, error)
}

type provider struct {
	optProvider ConfigProvider
	cfg         pkgConfig.Anywhere
	logger      *slog.Logger
}

func (provider *provider) Init(ctx context.Context) error {
	return provider.optProvider.Init(ctx)
}

type awslogger struct {
	logger *slog.Logger
	ctx    context.Context
}

func (awslogger *awslogger) Logf(classification logging.Classification, format string, v ...interface{}) {
	var lvl slog.Level
	switch classification {
	case logging.Debug:
		lvl = slog.LevelDebug
	case logging.Warn:
		lvl = slog.LevelWarn
	default:
		lvl = slog.LevelInfo
	}

	awslogger.logger.Log(awslogger.ctx, lvl, fmt.Sprintf(format, v...))
}

func (provider *provider) GetAwsConfig(ctx context.Context) (aws.Config, error) {
	opts, err := provider.optProvider.GetOpts(ctx)
	if err != nil {
		return aws.Config{}, err
	}

	opts = append(opts, awscfg.WithRegion(provider.cfg.Config.Region))

	return awscfg.LoadDefaultConfig(ctx, opts...)
}

func (provider *provider) GetGameLift(ctx context.Context) (GameLift, error) {
	cfg, err := provider.GetAwsConfig(ctx)
	if err != nil {
		return nil, err
	}

	return gl.NewFromConfig(cfg), nil
}

func NewProvider(cfg pkgConfig.Anywhere, logger *slog.Logger) (Provider, error) {
	var configProvider ConfigProvider
	if len(cfg.Config.Region) == 0 {
		return nil, errors.New("no region specified")
	}

	switch cfg.Config.Provider {
	case pkgConfig.AwsConfigProviderProfile:
		configProvider = NewProfileProvider(cfg.Config.Profile)
	case pkgConfig.AwsConfigProviderSSOFile:
		if len(cfg.Config.SSOFile) == 0 {
			return nil, fmt.Errorf("no sso env file path provided")
		}
		configProvider = NewSSOFileProvider(cfg.Config.SSOFile, logger)
	default:
		return nil, fmt.Errorf("unknown provider: %s", cfg.Config.Provider)
	}

	p := &provider{
		optProvider: configProvider,
		cfg:         cfg,
		logger:      logger,
	}

	return p, nil
}
