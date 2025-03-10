/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package observability

import (
	"context"
	"encoding/hex"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Spanner provides an interface for creating and managing trace spans.
type Spanner interface {
	NewSpan(ctx context.Context, name string, meta map[string]string) (context.Context, trace.Span, error)
	NewSpanWithTraceId(ctx context.Context, name string, traceId uuid.UUID, meta map[string]string) (context.Context, trace.Span, error)
}

type spanner struct {
	tracer trace.Tracer
}

func (spanner *spanner) NewSpanWithTraceId(ctx context.Context, name string, traceId uuid.UUID, meta map[string]string) (context.Context, trace.Span, error) {
	spanStartOptions := make([]trace.SpanStartOption, 0)

	l := trace.LinkFromContext(ctx)
	spanStartOptions = append(spanStartOptions, trace.WithLinks(l))
	ctx = context.Background() // Make new context - this is to break the span parent on purpose

	spanContextConfig := trace.SpanContextConfig{
		TraceFlags: trace.FlagsSampled,
		Remote:     false,
	}

	tid, err := trace.TraceIDFromHex(hex.EncodeToString(traceId[:]))
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to create span")
	}
	spanContextConfig.TraceID = tid

	spanContext := trace.NewSpanContext(spanContextConfig)

	ctx = trace.ContextWithSpanContext(ctx, spanContext)

	ctx, span := spanner.tracer.Start(ctx, name, spanStartOptions...)

	attributes := make([]attribute.KeyValue, 0)
	for k, v := range meta {
		a := attribute.String(k, v)
		attributes = append(attributes, a)
	}

	span.SetAttributes(attributes...)

	return ctx, span, nil
}

func (spanner *spanner) NewSpan(ctx context.Context, name string, meta map[string]string) (context.Context, trace.Span, error) {
	ctx, span := spanner.tracer.Start(ctx, name)

	attr := make([]attribute.KeyValue, 0)
	for k, v := range meta {
		a := attribute.String(k, v)
		attr = append(attr, a)
	}

	span.SetAttributes(attr...)
	return ctx, span, nil
}

// NewSpanner creates a new Spanner instance with the provided tracer.
//
// Parameters:
//   - tracer: OpenTelemetry tracer instance to use for span creation
//
// Returns:
//   - Spanner: Configured spanner instance
func NewSpanner(tracer trace.Tracer) Spanner {
	return &spanner{
		tracer: tracer,
	}
}
