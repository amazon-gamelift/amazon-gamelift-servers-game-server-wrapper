//go:build windows

/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package process

import (
	"os"
)

func ensureExecutable(fi os.FileInfo, path string) error {
	// check path is executable by running user in windows
	isExecAny(fi.Mode())
	return nil
}

func isExecAny(mode os.FileMode) bool {
	// executable by user, group or any
	return mode&0111 != 0
}
