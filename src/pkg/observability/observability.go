/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package observability

import (
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"log/slog"
)

// Observability manages all observability abilities within an application.
type Observability struct {
	Spanner    Spanner
	Tracer     trace.Tracer
	Meter      metric.Meter
	LogHandler slog.Handler
}
