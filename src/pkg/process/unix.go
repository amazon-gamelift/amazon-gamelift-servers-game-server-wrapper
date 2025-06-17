//go:build unix

/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package process

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
)

func ensureExecutable(fi os.FileInfo, path string) error {
	m := fi.Mode()

	if !((m.IsRegular()) || (uint32(m&fs.ModeSymlink) == 0)) {
		return errors.Errorf("file '%s' is not a normal file or symlink", path)
	}

	pathParts := strings.Split(path, "/")
	parentPath := strings.TrimSuffix(path, pathParts[len(pathParts)-1])

	chmodString := fmt.Sprintf("chmod -R 755 %v", parentPath)
	cmd := exec.Command("sh", "-c", chmodString)
	var stderr bytes.Buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return errors.Errorf("error setting permissions for folder '%s', error: %v, stderr: %v", parentPath, err, stderr.String())
	}

	if unix.Access(path, unix.X_OK) != nil {
		return errors.Errorf("file '%s' cannot be executed by this user", path)
	}

	return nil
}
