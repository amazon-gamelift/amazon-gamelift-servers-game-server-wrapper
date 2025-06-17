/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package observability

import (
	"context"
	"log/slog"
	"time"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/trace"
)

// Provider manages the lifecycle and accessibility of the observability
// components. It includes logging, metrics, and tracing providers.
type Provider struct {
	resource *resource.Resource
	logger   LoggerProvider
	metrics  MeterProvider
	tracer   TracerProvider
}

// Close shuts down the observability components with a timeout.
// It ensures all pending data is flushed before shutdown.
//
// Returns:
//   - error: Combined error if any component fails to shut down
func (provider *Provider) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	errs := make([]error, 0)
	if provider.logger != nil {
		if err := provider.logger.ForceFlush(ctx); err != nil {
			errs = append(errs, err)
		}
		if err := provider.logger.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if provider.metrics != nil {
		if err := provider.metrics.ForceFlush(ctx); err != nil {
			errs = append(errs, err)
		}
		if err := provider.metrics.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if provider.tracer != nil {
		if err := provider.tracer.ForceFlush(ctx); err != nil {
			errs = append(errs, err)
		}
		if err := provider.tracer.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		return nil
	}

	if len(errs) == 1 {
		return errs[0]
	}

	var err error
	for i := len(errs) - 1; i >= 0; i-- {
		e := errs[i]
		if err == nil {
			err = e
		} else {
			err = errors.Wrap(err, e.Error())
		}
	}

	err = errors.Wrapf(err, "multiple errors")

	return err
}

func (provider *Provider) GetLogger(name string, next slog.Handler) slog.Handler {
	otelLogger := provider.logger.Logger(name)

	otelHandler := &Handler{
		logger: otelLogger,
		next:   next,
	}

	return otelHandler
}

func (provider *Provider) GetTracer(name string) trace.Tracer {
	return provider.tracer.Tracer(name)
}

func (provider *Provider) GetMetrics(name string) metric.Meter {
	return provider.metrics.Meter(name)
}

func (provider *Provider) Stack(log, trace, metric string, nextLogger slog.Handler) *Observability {
	o := &Observability{
		Meter:      provider.GetMetrics(metric),
		Tracer:     provider.GetTracer(trace),
		LogHandler: provider.GetLogger(log, nextLogger),
	}

	o.Spanner = NewSpanner(o.Tracer)

	return o
}

// NewProvider creates a new Provider instance with all observability
// components configured based on the provided configuration.
//
// Parameters:
//   - ctx: Context for provider initialization
//   - cfg: Configuration for observability components
//
// Returns:
//   - *Provider: Configured provider instance
//   - error: Any error encountered during setup
func NewProvider(ctx context.Context, cfg *Config) (*Provider, error) {
	provider := &Provider{}

	resourceInstance, err := NewResource(cfg.Name)
	if err != nil {
		return nil, err
	}
	provider.resource = resourceInstance

	logger, err := NewLoggingProvider(ctx, resourceInstance, cfg)
	if err != nil {
		return nil, err
	}
	provider.logger = logger

	metrics, err := NewMetricsProvider(ctx, resourceInstance, cfg)
	if err != nil {
		return nil, err
	}
	provider.metrics = metrics

	tracer, err := NewTracingProvider(ctx, resourceInstance, cfg)
	if err != nil {
		return nil, err
	}
	provider.tracer = tracer

	return provider, nil
}
