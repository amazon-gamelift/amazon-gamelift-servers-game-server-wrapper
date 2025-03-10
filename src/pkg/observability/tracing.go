/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package observability

import (
	"context"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"net/url"
	"time"
)

// TracerProvider defines the interface for managing trace collection
// and export operations. It provides methods for creating tracers,
// flushing pending spans, and shutdown.
type TracerProvider interface {
	ForceFlush(ctx context.Context) error
	Shutdown(ctx context.Context) error
	Tracer(name string, options ...trace.TracerOption) trace.Tracer
}


// NewTracingProvider creates a new TracerProvider configured with the provided
// resource and configuration options.
//
// Parameters:
//   - ctx: Context for setup operations
//   - resource: Resource information to be attached to the spans
//   - cfg: Configuration for the trace provider
//
// Returns:
//   - TracerProvider: Configured trace provider
//   - error: Any error encountered during setup
func NewTracingProvider(ctx context.Context, resource *resource.Resource, cfg *Config) (TracerProvider, error) {
	tracerProviderOptions := make([]sdktrace.TracerProviderOption, 0)
	tracerProviderOptions = append(tracerProviderOptions, sdktrace.WithResource(resource))

	if cfg.Enabled {
		contextWithTimeout, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()

		opt := make([]otlptracehttp.Option, 0)
		if cfg.Retry != nil {
			opt = append(opt, otlptracehttp.WithRetry(otlptracehttp.RetryConfig{
				Enabled:         cfg.Retry.Enabled,
				MaxInterval:     cfg.Retry.MaxInterval,
				InitialInterval: cfg.Retry.InitialInterval,
				MaxElapsedTime:  cfg.Retry.MaxElapsedTime,
			}))
		}

		if cfg.Auth != nil && len(cfg.Auth.User) > 0 && len(cfg.Auth.Password) > 0 {
			opt = append(opt, otlptracehttp.WithHeaders(basicAuthHeaders(cfg.Auth.User, cfg.Auth.Password)))
		}

		if len(cfg.Endpoint) > 0 {
			endpoint, err := url.JoinPath(cfg.Endpoint, "/v1/traces")
			if err != nil {
				return nil, err
			}

			opt = append(opt, otlptracehttp.WithEndpointURL(endpoint))
		}

		if cfg.Insecure {
			opt = append(opt, otlptracehttp.WithInsecure())
		}

		exporter, err := otlptracehttp.New(contextWithTimeout, opt...)
		if err != nil {
			return nil, err
		}

		tracerProviderOptions = append(tracerProviderOptions, sdktrace.WithBatcher(exporter, sdktrace.WithBatchTimeout(cfg.ExportInterval)))
	}

	prov := sdktrace.NewTracerProvider(tracerProviderOptions...)

	return prov, nil
}
