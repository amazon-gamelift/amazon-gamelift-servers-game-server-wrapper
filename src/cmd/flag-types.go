/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cmd

import (
	"time"

	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/config"
)

type hostingProvider config.Provider

func (hostingProviderInstance *hostingProvider) String() string {
	hs := string(*hostingProviderInstance)
	return hs
}

func (hostingProviderInstance *hostingProvider) Set(s string) error {
	*hostingProviderInstance = hostingProvider(s)
	return nil
}

func (hostingProviderInstance *hostingProvider) Type() string { return "hosting-provider" }

func newHostingProvider(p *config.Provider) *hostingProvider {
	*p = ""
	return (*hostingProvider)(p)
}

type timespan time.Duration

func (timeSpan *timespan) String() string {
	if timeSpan == nil {
		return ""
	}
	t := time.Duration(*timeSpan)
	hs := t.String()
	return hs
}

func (timeSpan *timespan) Set(s string) error {
	d, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	*timeSpan = timespan(d)
	return nil
}

func (timeSpan *timespan) Type() string { return "timespan" }

func newTimespan(duration *time.Duration, d time.Duration) *timespan {
	*duration = d
	return (*timespan)(duration)
}
