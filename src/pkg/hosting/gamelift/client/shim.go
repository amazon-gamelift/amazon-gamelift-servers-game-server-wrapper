/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/gamelift"
)

type GameLift interface {
	ListCompute(ctx context.Context, params *gamelift.ListComputeInput, optFns ...func(*gamelift.Options)) (*gamelift.ListComputeOutput, error)
	RegisterCompute(ctx context.Context, params *gamelift.RegisterComputeInput, optFns ...func(*gamelift.Options)) (*gamelift.RegisterComputeOutput, error)
	DeregisterCompute(ctx context.Context, params *gamelift.DeregisterComputeInput, optFns ...func(*gamelift.Options)) (*gamelift.DeregisterComputeOutput, error)
	GetComputeAuthToken(ctx context.Context, params *gamelift.GetComputeAuthTokenInput, optFns ...func(*gamelift.Options)) (*gamelift.GetComputeAuthTokenOutput, error)
}
