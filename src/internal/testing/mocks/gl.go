/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mocks

import (
	"runtime"
	"strings"

	"github.com/amazon-gamelift/amazon-gamelift-servers-go-server-sdk/v5/server"
)

type GlMock struct {
	MethodsCalled []string
	InitSDKParams *server.ServerParameters
	InitSDKError  error

	InitSDKFromEnvironmentError error

	ProcessReadyParams *server.ProcessParameters
	ProcessReadyError  error

	ProcessEndingError error

	ActivateGameSessionError error

	DestroyError error
}

func (g *GlMock) RegisterCall() {
	pc := make([]uintptr, 10) // at least 1 entry needed
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	name := f.Name()
	parts := strings.Split(name, ".")
	name = parts[len(parts)-1]
	g.MethodsCalled = append(g.MethodsCalled, name)
}

func (g *GlMock) InitSDK(params server.ServerParameters) error {
	g.RegisterCall()
	g.InitSDKParams = &params
	return g.InitSDKError
}

func (g *GlMock) InitSDKFromEnvironment() error {
	g.RegisterCall()
	return g.InitSDKFromEnvironmentError
}

func (g *GlMock) ProcessReady(params server.ProcessParameters) error {
	g.RegisterCall()
	g.ProcessReadyParams = &params
	return g.ProcessReadyError
}

func (g *GlMock) ProcessEnding() error {
	g.RegisterCall()
	return g.ProcessEndingError
}

func (g *GlMock) ActivateGameSession() error {
	g.RegisterCall()
	return g.ActivateGameSessionError
}

func (g *GlMock) Destroy() error {
	g.RegisterCall()
	return g.DestroyError
}

func NewGlMock() *GlMock {
	m := &GlMock{
		MethodsCalled: make([]string, 0),
	}

	return m
}
