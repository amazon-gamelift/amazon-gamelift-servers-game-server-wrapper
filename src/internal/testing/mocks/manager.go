/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mocks

import (
	"context"
	"os/exec"
)

type ManagerMock struct {
	RunCtx context.Context
	RunErr error

	SetCommandCmd *exec.Cmd
}

func (i *ManagerMock) Run(ctx context.Context) error {
	i.RunCtx = ctx
	return i.RunErr
}

func (i *ManagerMock) SetCommand(cmd *exec.Cmd) {
	i.SetCommandCmd = cmd
}
