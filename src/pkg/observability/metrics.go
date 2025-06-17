/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package observability

import (
	"context"
	"net/url"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

// MeterProvider defines the interface for managing metrics collection.
// It provides methods for creating meters, flushing pending metrics, and shutdown.
type MeterProvider interface {
	ForceFlush(ctx context.Context) error
	Shutdown(ctx context.Context) error
	Meter(name string, options ...metric.MeterOption) metric.Meter
}

// NewMetricsProvider creates a new MeterProvider configured with the provided
// resource and configuration options.
//
// Parameters:
//   - ctx: Context for setup operations
//   - resource: Resource information to be attached to metrics
//   - cfg: Configuration for the meter provider
//
// Returns:
//   - MeterProvider: Configured meter provider
//   - error: Any error encountered during setup
func NewMetricsProvider(ctx context.Context, resource *resource.Resource, cfg *Config) (MeterProvider, error) {
	opts := make([]metricsdk.Option, 0)
	opts = append(opts, metricsdk.WithResource(resource))

	if cfg.Enabled {
		cc, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()

		opt := make([]otlpmetrichttp.Option, 0)
		if cfg.Retry != nil {
			opt = append(opt, otlpmetrichttp.WithRetry(otlpmetrichttp.RetryConfig{
				Enabled:         cfg.Retry.Enabled,
				MaxInterval:     cfg.Retry.MaxInterval,
				InitialInterval: cfg.Retry.InitialInterval,
				MaxElapsedTime:  cfg.Retry.MaxElapsedTime,
			}))
		}

		if cfg.Auth != nil && len(cfg.Auth.User) > 0 && len(cfg.Auth.Password) > 0 {
			opt = append(opt, otlpmetrichttp.WithHeaders(basicAuthHeaders(cfg.Auth.User, cfg.Auth.Password)))
		}

		if len(cfg.Endpoint) > 0 {
			endpoint, err := url.JoinPath(cfg.Endpoint, "/v1/metrics")
			if err != nil {
				return nil, err
			}

			opt = append(opt, otlpmetrichttp.WithEndpointURL(endpoint))
		}

		if cfg.Insecure {
			opt = append(opt, otlpmetrichttp.WithInsecure())
		}

		exporter, err := otlpmetrichttp.New(cc, opt...)
		if err != nil {
			return nil, err
		}

		reader := metricsdk.NewPeriodicReader(exporter, metricsdk.WithInterval(time.Second*1))
		opts = append(opts, metricsdk.WithReader(reader))
	}

	provider := metricsdk.NewMeterProvider(opts...)

	return provider, nil
}
