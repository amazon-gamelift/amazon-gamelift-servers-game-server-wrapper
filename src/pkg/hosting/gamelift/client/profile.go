/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/pkg/errors"
)

var _ ConfigProvider = (*ProfileProvider)(nil)

type ProfileProvider struct {
	name string
	opts []func(options *awscfg.LoadOptions) error
}

func (profileProvider *ProfileProvider) Init(ctx context.Context) error {
	if len(profileProvider.name) == 0 {
		return errors.New("no aws profile provided")
	}

	profileProvider.opts = []func(options *awscfg.LoadOptions) error{
		awscfg.WithSharedConfigProfile(profileProvider.name),
	}

	return nil
}

func (profileProvider *ProfileProvider) GetOpts(ctx context.Context) ([]func(options *awscfg.LoadOptions) error, error) {
	if profileProvider.opts == nil {
		return nil, errors.New("no aws config, did you call Init()?")
	}

	return profileProvider.opts, nil
}

func NewProfileProvider(name string) *ProfileProvider {
	p := &ProfileProvider{
		name: name,
	}

	return p
}
