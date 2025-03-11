/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/gamelift"
	"golang.org/x/net/context"
)

type ClientProviderMock struct {
	InitError            error
	GetAwsConfigError    error
	GetGameLiftError     error
	GetAwsConfigResponse aws.Config
	GetGameLiftResponse  *ClientGameLiftMock
	InitCalled           bool
	GetAwsConfigCalled   bool
	GetGameLiftCalled    bool
}

func (clientProviderMock *ClientProviderMock) Init(ctx context.Context) error {
	clientProviderMock.InitCalled = true
	return clientProviderMock.InitError
}

func (clientProviderMock *ClientProviderMock) GetAwsConfig(ctx context.Context) (aws.Config, error) {
	clientProviderMock.GetAwsConfigCalled = true
	return clientProviderMock.GetAwsConfigResponse, clientProviderMock.GetAwsConfigError
}

func (clientProviderMock *ClientProviderMock) GetGameLift(ctx context.Context) (GameLift, error) {
	clientProviderMock.GetGameLiftCalled = true
	return clientProviderMock.GetGameLiftResponse, clientProviderMock.GetGameLiftError
}

type ClientGameLiftMock struct {
	ListComputeInput  *gamelift.ListComputeInput
	ListComputeResult *gamelift.ListComputeOutput
	ListComputeError  error

	RegisterComputeInput  *gamelift.RegisterComputeInput
	RegisterComputeResult *gamelift.RegisterComputeOutput
	RegisterComputeError  error

	DeregisterComputeInput  *gamelift.DeregisterComputeInput
	DeregisterComputeResult *gamelift.DeregisterComputeOutput
	DeregisterComputeError  error

	GetComputeAuthTokenInput  *gamelift.GetComputeAuthTokenInput
	GetComputeAuthTokenResult *gamelift.GetComputeAuthTokenOutput
	GetComputeAuthTokenError  error
}

func (clientGameLiftMock *ClientGameLiftMock) ListCompute(ctx context.Context, params *gamelift.ListComputeInput, optFns ...func(*gamelift.Options)) (*gamelift.ListComputeOutput, error) {
	clientGameLiftMock.ListComputeInput = params
	return clientGameLiftMock.ListComputeResult, clientGameLiftMock.ListComputeError
}

func (clientGameLiftMock *ClientGameLiftMock) RegisterCompute(ctx context.Context, params *gamelift.RegisterComputeInput, optFns ...func(*gamelift.Options)) (*gamelift.RegisterComputeOutput, error) {
	clientGameLiftMock.RegisterComputeInput = params
	return clientGameLiftMock.RegisterComputeResult, clientGameLiftMock.RegisterComputeError
}

func (clientGameLiftMock *ClientGameLiftMock) DeregisterCompute(ctx context.Context, params *gamelift.DeregisterComputeInput, optFns ...func(*gamelift.Options)) (*gamelift.DeregisterComputeOutput, error) {
	clientGameLiftMock.DeregisterComputeInput = params
	return clientGameLiftMock.DeregisterComputeResult, clientGameLiftMock.DeregisterComputeError
}

func (clientGameLiftMock *ClientGameLiftMock) GetComputeAuthToken(ctx context.Context, params *gamelift.GetComputeAuthTokenInput, optFns ...func(*gamelift.Options)) (*gamelift.GetComputeAuthTokenOutput, error) {
	clientGameLiftMock.GetComputeAuthTokenInput = params
	return clientGameLiftMock.GetComputeAuthTokenResult, clientGameLiftMock.GetComputeAuthTokenError
}

type MockReadCloser struct {
	ExpectedData []byte
	ExpectedErr  error
}

func (mockReadCloser *MockReadCloser) Read(p []byte) (int, error) {
	copy(p, mockReadCloser.ExpectedData)
	return len(mockReadCloser.ExpectedData), mockReadCloser.ExpectedErr
}
func (mockReadCloser *MockReadCloser) Close() error { return nil }

type MockWriteCloser struct{}

func (mockWriteCloser *MockWriteCloser) Close() error { return nil }
func (mockWriteCloser *MockWriteCloser) Write(b []byte) (n int, err error) {
	return 0, nil
}
