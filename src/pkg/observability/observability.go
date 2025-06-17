/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package observability

import (
	"log/slog"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// Observability manages all observability abilities within an application.
type Observability struct {
	Spanner    Spanner
	Tracer     trace.Tracer
	Meter      metric.Meter
	LogHandler slog.Handler
}
