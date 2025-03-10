/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package observability

import "time"

// Config defines the main configuration for observability features.
type Config struct {
	Name           string        `mapstructure:"-"`
	Endpoint       string        `mapstructure:"endpoint,omitempty"`
	Enabled        bool          `mapstructure:"enabled,omitempty"`
	Retry          *RetryConfig  `mapstructure:"retry,omitempty"`
	Insecure       bool          `mapstructure:"insecure,omitempty"`
	Auth           *AuthConfig   `mapstructure:"auth,omitempty"`
	ExportInterval time.Duration `mapstructure:"exportInterval,omitempty"`
}

// AuthConfig defines authentication credentials for the observability configuration.
type AuthConfig struct {
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

// RetryConfig defines the behavior for retrying failed export operations.
type RetryConfig struct {
	// Enabled indicates whether to not retry sending batches in case of
	// export failure.
	Enabled bool `mapstructure:"enabled,omitempty"`
	// InitialInterval the time to wait after the first failure before
	// retrying.
	InitialInterval time.Duration `mapstructure:"initialInterval,omitempty"`
	// MaxInterval is the upper bound on backoff interval. Once this value is
	// reached the delay between consecutive retries will always be
	// `MaxInterval`.
	MaxInterval time.Duration `mapstructure:"maxInterval,omitempty"`
	// MaxElapsedTime is the maximum amount of time (including retries) spent
	// trying to send a request/batch.  Once this value is reached, the data
	// is discarded.
	MaxElapsedTime time.Duration `mapstructure:"maxElapsedTime,omitempty"`
}
