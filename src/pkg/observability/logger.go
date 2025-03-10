/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package observability

import (
	"context"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/log"
	logsdk "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/trace"
	"log/slog"
	"net/url"
	"strconv"
	"time"
)

// LoggerProvider defines the interface for managing log collection.
// It provides methods for creating loggers, flushing pending logs, and shutdown.
type LoggerProvider interface {
	ForceFlush(ctx context.Context) error
	Shutdown(ctx context.Context) error
	Logger(name string, opts ...log.LoggerOption) log.Logger
}

// NewLoggingProvider creates a new LoggerProvider configured with the provided
// resource and configuration options.
//
// Parameters:
//   - ctx: Context for setup operations
//   - resource: Resource information to be attached to logs
//   - cfg: Configuration for the logger provider
//
// Returns:
//   - LoggerProvider: Configured logger provider
//   - error: Any error encountered during setup
func NewLoggingProvider(ctx context.Context, resource *resource.Resource, cfg *Config) (LoggerProvider, error) {
	loggerProviderOptions := make([]logsdk.LoggerProviderOption, 0)
	loggerProviderOptions = append(loggerProviderOptions, logsdk.WithResource(resource))

	if cfg.Enabled {
		timeoutContext, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()

		opt := make([]otlploghttp.Option, 0)
		if cfg.Retry != nil {
			opt = append(opt, otlploghttp.WithRetry(otlploghttp.RetryConfig{
				Enabled:         cfg.Retry.Enabled,
				MaxInterval:     cfg.Retry.MaxInterval,
				InitialInterval: cfg.Retry.InitialInterval,
				MaxElapsedTime:  cfg.Retry.MaxElapsedTime,
			}))
		}

		if cfg.Auth != nil && len(cfg.Auth.User) > 0 && len(cfg.Auth.Password) > 0 {
			opt = append(opt, otlploghttp.WithHeaders(basicAuthHeaders(cfg.Auth.User, cfg.Auth.Password)))
		}

		if len(cfg.Endpoint) > 0 {
			endpoint, err := url.JoinPath(cfg.Endpoint, "/v1/logs")
			if err != nil {
				return nil, err
			}

			opt = append(opt, otlploghttp.WithEndpointURL(endpoint))
		}

		if cfg.Insecure {
			opt = append(opt, otlploghttp.WithInsecure())
		}
		exporter, err := otlploghttp.New(timeoutContext, opt...)
		if err != nil {
			return nil, err
		}

		proc := logsdk.NewBatchProcessor(exporter, logsdk.WithExportInterval(cfg.ExportInterval))
		loggerProviderOptions = append(loggerProviderOptions, logsdk.WithProcessor(proc))
	}

	prov := logsdk.NewLoggerProvider(loggerProviderOptions...)

	return prov, nil
}

// Handler implements slog.Handler interface to provide OpenTelemetry
// integration with Go's structured logging slog.
type Handler struct {
	logger log.Logger
	next   slog.Handler
}

func (handler *Handler) record(level slog.Level) *log.Record {
	r := &log.Record{}

	switch level {
	case slog.LevelError:
		r.SetSeverity(log.SeverityError)
	case slog.LevelWarn:
		r.SetSeverity(log.SeverityWarn)
	case slog.LevelInfo:
		r.SetSeverity(log.SeverityInfo)
	case slog.LevelDebug:
		r.SetSeverity(log.SeverityDebug)
	default:
		r.SetSeverity(log.SeverityInfo)
	}

	return r
}

func Slogger(logger log.Logger, next slog.Handler) *slog.Logger {
	return slog.New(&Handler{
		logger: logger,
		next:   next,
	})
}

func (handler *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	r := handler.record(level)
	otelEnabled := handler.logger.Enabled(ctx, *r)
	if otelEnabled {
		return true
	}

	if handler.next == nil {
		return false
	}

	return handler.next.Enabled(ctx, level)
}

func (handler *Handler) sendNext(ctx context.Context, record slog.Record) error {
	if handler.next != nil {
		return handler.next.Handle(ctx, record)
	}
	return nil
}

func (handler *Handler) Handle(ctx context.Context, record slog.Record) error {
	if ctx == nil {
		return handler.sendNext(ctx, record)
	}

	span := trace.SpanFromContext(ctx)
	if span == nil || !span.IsRecording() {
		return handler.sendNext(ctx, record)
	}

	r := handler.record(record.Level)
	r.SetBody(log.StringValue(record.Message))
	r.SetTimestamp(record.Time)

	record.Attrs(func(attr slog.Attr) bool {
		var kv log.KeyValue

		switch attr.Value.Kind() {
		case slog.KindBool:
			kv = log.Bool(attr.Key, attr.Value.Bool())
			break
		case slog.KindDuration:
			kv = log.String(attr.Key, attr.Value.Duration().String())
			break
		case slog.KindFloat64:
			kv = log.Float64(attr.Key, attr.Value.Float64())
			break
		case slog.KindTime:
			kv = log.String(attr.Key, attr.Value.Time().Format(time.RFC3339Nano))
			break
		case slog.KindUint64:
			kv = log.String(attr.Key, strconv.FormatUint(attr.Value.Uint64(), 10))
			break
		case slog.KindString:
			fallthrough
		default:
			kv = log.String(attr.Key, attr.Value.String())
		}

		r.AddAttributes(kv)
		return true
	})

	handler.logger.Emit(ctx, *r)

	return handler.sendNext(ctx, record)
}

func (handler *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return handler
}

func (handler *Handler) WithGroup(name string) slog.Handler {
	return handler
}
