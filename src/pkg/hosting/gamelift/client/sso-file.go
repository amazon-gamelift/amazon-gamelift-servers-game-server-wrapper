/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"
	"log/slog"
	"reflect"
	"strings"
	"sync"

	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/constants"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/fsnotify/fsnotify"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var _ ConfigProvider = (*SSOFileProvider)(nil)

type SSOFileProvider struct {
	path   string
	opts   []func(options *awscfg.LoadOptions) error
	logger *slog.Logger
	v      *viper.Viper
	mutex  sync.Mutex
}

type SSOFileConfig struct {
	AccessKey    string `mapstructure:"aws_access_key_id"`
	SecretKey    string `mapstructure:"aws_secret_access_key"`
	SessionToken string `mapstructure:"aws_session_token"`
}

func (ssoFileProvider *SSOFileProvider) onConfigChange(ctx context.Context, e fsnotify.Event) {
	ssoFileProvider.logger.DebugContext(ctx, "aws env file config file changed", "file", e.Name, "change", e.Op.String())

	if err := ssoFileProvider.loadOpts(ctx); err != nil {
		ssoFileProvider.logger.ErrorContext(ctx, "failed to load aws config", "err", err)
	}
}

func ssoDecodeHook(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.Map {
		return data, nil
	}

	dataIf, ok := data.(map[string]interface{})
	if !ok {
		return data, errors.New("not a map")
	}

	if len(dataIf) != 1 {
		return data, errors.New("only 1 entry permitted")
	}

	var inner interface{}
	for _, v := range dataIf {
		inner = v
	}

	var s SSOFileConfig
	if err := mapstructure.Decode(inner, &s); err != nil {
		return data, err
	}

	return s, nil
}

func (ssoFileProvider *SSOFileProvider) loadOpts(ctx context.Context) error {
	ssoFileProvider.mutex.Lock()
	defer ssoFileProvider.mutex.Unlock()

	ssoFileProvider.logger.DebugContext(ctx, "loading aws config")
	profile := SSOFileConfig{}
	if err := ssoFileProvider.v.Unmarshal(&profile, viper.DecodeHook(ssoDecodeHook)); err != nil {
		return errors.Wrap(err, "failed to unmarshal config")
	}

	profile.AccessKey = strings.TrimSpace(profile.AccessKey)
	profile.SecretKey = strings.TrimSpace(profile.SecretKey)
	profile.SessionToken = strings.TrimSpace(profile.SessionToken)

	if l := len(profile.AccessKey); l < 16 {
		return errors.Errorf("invalid access key length: %d", l)
	}

	if l := len(profile.SecretKey); l < 16 {
		return errors.Errorf("invalid secret key length: %d", l)
	}

	if l := len(profile.SessionToken); l < 16 {
		return errors.Errorf("invalid session token length: %d", l)
	}

	scp := credentials.NewStaticCredentialsProvider(profile.AccessKey, profile.SecretKey, profile.SessionToken)
	scp.Value.CanExpire = false

	ssoFileProvider.opts = []func(options *awscfg.LoadOptions) error{
		awscfg.WithCredentialsProvider(scp),
		awscfg.WithDefaultRegion("us-east-1"),
	}

	return nil
}

func (ssoFileProvider *SSOFileProvider) Init(ctx context.Context) error {

	ssoFileProvider.v = viper.New()
	appPath, ok := ctx.Value(string(constants.ContextKeyAppDir)).(string)
	if ok && len(appPath) > 0 {
		ssoFileProvider.v.AddConfigPath(appPath)
	}

	ssoFileProvider.v.SetConfigFile(ssoFileProvider.path)

	if err := ssoFileProvider.v.ReadInConfig(); err != nil {
		return errors.Wrap(err, "failed to read env file")
	}

	if err := ssoFileProvider.loadOpts(ctx); err != nil {
		return errors.Wrap(err, "failed to load config")
	}

	ssoFileProvider.v.WatchConfig()
	ssoFileProvider.v.OnConfigChange(func(e fsnotify.Event) { ssoFileProvider.onConfigChange(ctx, e) })

	return nil
}

func (ssoFileProvider *SSOFileProvider) GetOpts(ctx context.Context) ([]func(options *awscfg.LoadOptions) error, error) {
	ssoFileProvider.mutex.Lock()
	defer ssoFileProvider.mutex.Unlock()

	if ssoFileProvider.opts == nil {
		return nil, errors.New("no aws config, did you call Init()?")
	}

	return ssoFileProvider.opts, nil
}

func NewSSOFileProvider(path string, logger *slog.Logger) *SSOFileProvider {
	p := &SSOFileProvider{
		path:   path,
		logger: logger,
	}

	return p
}
