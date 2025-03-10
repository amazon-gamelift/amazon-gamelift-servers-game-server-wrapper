/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mocks

import "context"

type InitMock struct {
	InitCtx context.Context
	InitErr error

	CloseCtx context.Context
	CloseErr error
}

func (i *InitMock) InitSdk(ctx context.Context) error {
	i.InitCtx = ctx
	return i.InitErr
}

func (i *InitMock) Close(ctx context.Context) error {
	i.CloseCtx = ctx
	return i.CloseErr
}
