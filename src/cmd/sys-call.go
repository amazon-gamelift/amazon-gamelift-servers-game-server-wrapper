/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func addSyscallInterrupt(ctx context.Context) (context.Context, context.CancelFunc) {
	sysInterrupt := make(chan os.Signal, 1)
	signal.Notify(sysInterrupt, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(ctx)
	go func() {
		sig := <-sysInterrupt
		logger.WarnContext(ctx, "received system signal", "signal", sig)
		cancel()
	}()

	return ctx, cancel
}
