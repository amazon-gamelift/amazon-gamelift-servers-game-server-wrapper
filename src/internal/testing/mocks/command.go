/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mocks

import (
	"context"
	"os/exec"
)

type CommandMock struct {
	GetCtx context.Context
	GetCmd *exec.Cmd
	GetErr error
}

func (i *CommandMock) Get(ctx context.Context) (*exec.Cmd, error) {
	i.GetCtx = ctx
	return i.GetCmd, i.GetErr
}
