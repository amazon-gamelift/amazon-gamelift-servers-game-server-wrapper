/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package observability

import (
	"fmt"
	"log/slog"

	"github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/pkg/constants"
	"github.com/go-logr/logr"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"
)

func SetLogger(handler slog.Handler) {
	srcAttr := slog.Attr{Key: string(constants.ContextKeySource), Value: slog.StringValue("otel")}

	logHandler := logr.FromSlogHandler(handler.WithAttrs([]slog.Attr{srcAttr}))

	otel.SetLogger(logHandler)
}

func NewResource(serviceName string) (*resource.Resource, error) {
	resourceInstance, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("could not merge resources: %w", err)
	}

	return resourceInstance, nil
}
