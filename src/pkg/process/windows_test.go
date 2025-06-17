/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package process

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetByIdRedisCacheHit(t *testing.T) {
	// Arrange
	filename := "./testrun.sh"
	fi, err := os.Stat(filename)
	if err != nil {
		assert.Fail(t, "Error getting file stats", "err", err)
	}

	// Act
	err = ensureExecutable(fi, filename)
	if err != nil {
		assert.Fail(t, "EnsureExecutable Errored", "err", err)
	}

	//Assert
	assert.Nil(t, err)
}
