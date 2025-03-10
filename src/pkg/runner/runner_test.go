/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runner

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Runner_Run_HappyPath(t *testing.T) {
	//arrange
	gameLiftMockHelper := CreateRunnerMockHelper()

	//act
	err := gameLiftMockHelper.Runner.Run(gameLiftMockHelper.Ctx)

	//assert
	assert.Nil(t, err)
	assert.True(t, gameLiftMockHelper.ManagerService.InitCalled)
	assert.True(t, gameLiftMockHelper.ManagerService.RunCalled)
	assert.True(t, gameLiftMockHelper.ManagerService.CloseCalled)
	assert.Equal(t, 1, gameLiftMockHelper.ManagerService.InitCount)
	assert.Equal(t, 1, gameLiftMockHelper.ManagerService.RunCount)
	assert.Equal(t, 1, gameLiftMockHelper.ManagerService.CloseCount)
	assert.Equal(t, gameLiftMockHelper.ManagerService.InitRunId, gameLiftMockHelper.ManagerService.RunRunId)
	logString := gameLiftMockHelper.LogBuffer.String()
	assert.Contains(t, logString, "Starting a new run")
	assert.Contains(t, logString, "Executing the run")
	assert.Contains(t, logString, "Starting to clean up the run")
	assert.Contains(t, logString, "Run cleaned up")
}

func Test_Runner_Run_Manager_Init_Error(t *testing.T) {
	//arrange
	gameLiftMockHelper := CreateRunnerMockHelper()
	gameLiftMockHelper.ManagerService.InitError = errors.New("Unit Test")

	//act
	err := gameLiftMockHelper.Runner.Run(gameLiftMockHelper.Ctx)

	//assert
	assert.Errorf(t, err, "%s failed to init")
	assert.True(t, gameLiftMockHelper.ManagerService.InitCalled)
	assert.False(t, gameLiftMockHelper.ManagerService.RunCalled)
	assert.False(t, gameLiftMockHelper.ManagerService.CloseCalled)
	assert.Equal(t, 1, gameLiftMockHelper.ManagerService.InitCount)
	assert.Equal(t, 0, gameLiftMockHelper.ManagerService.RunCount)
	assert.Equal(t, 0, gameLiftMockHelper.ManagerService.CloseCount)
	logString := gameLiftMockHelper.LogBuffer.String()
	assert.Contains(t, logString, "Starting a new run")
	assert.NotContains(t, logString, "Executing the run")
	assert.NotContains(t, logString, "Starting to clean up the run")
	assert.NotContains(t, logString, "Run cleaned up")
}

func Test_Runner_Run_Manager_Run_Error(t *testing.T) {
	//arrange
	gameLiftMockHelper := CreateRunnerMockHelper()
	gameLiftMockHelper.ManagerService.RunError = errors.New("Unit Test")

	//act
	err := gameLiftMockHelper.Runner.Run(gameLiftMockHelper.Ctx)

	//assert
	assert.Errorf(t, err, "%s failed to init")
	assert.True(t, gameLiftMockHelper.ManagerService.InitCalled)
	assert.True(t, gameLiftMockHelper.ManagerService.RunCalled)
	assert.False(t, gameLiftMockHelper.ManagerService.CloseCalled)
	assert.Equal(t, 1, gameLiftMockHelper.ManagerService.InitCount)
	assert.Equal(t, 1, gameLiftMockHelper.ManagerService.RunCount)
	assert.Equal(t, 0, gameLiftMockHelper.ManagerService.CloseCount)
	logString := gameLiftMockHelper.LogBuffer.String()
	assert.Contains(t, logString, "Starting a new run")
	assert.Contains(t, logString, "Executing the run")
	assert.NotContains(t, logString, "Starting to clean up the run")
	assert.NotContains(t, logString, "Run cleaned up")
}

func Test_Runner_Run_Manager_Close_Error(t *testing.T) {
	//arrange
	gameLiftMockHelper := CreateRunnerMockHelper()
	gameLiftMockHelper.ManagerService.CloseError = errors.New("Unit Test")

	//act
	err := gameLiftMockHelper.Runner.Run(gameLiftMockHelper.Ctx)

	//assert
	assert.Errorf(t, err, "%s failed to close")
	assert.True(t, gameLiftMockHelper.ManagerService.InitCalled)
	assert.True(t, gameLiftMockHelper.ManagerService.RunCalled)
	assert.True(t, gameLiftMockHelper.ManagerService.CloseCalled)
	assert.Equal(t, 1, gameLiftMockHelper.ManagerService.InitCount)
	assert.Equal(t, 1, gameLiftMockHelper.ManagerService.RunCount)
	assert.Equal(t, 1, gameLiftMockHelper.ManagerService.CloseCount)
	logString := gameLiftMockHelper.LogBuffer.String()
	assert.Contains(t, logString, "Starting a new run")
	assert.Contains(t, logString, "Executing the run")
	assert.Contains(t, logString, "Starting to clean up the run")
	assert.NotContains(t, logString, "Run cleaned up")
}
