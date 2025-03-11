/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package constants

type ContextKey string

const (
	ContextKeySource         ContextKey = "src"
	ContextKeyPackageName    ContextKey = "pkg"
	ContextKeyVersion        ContextKey = "ver"
	ContextKeyAppDir         ContextKey = "appdir"
	ContextKeyRunId          ContextKey = "runid"
	ContextKeyRunLogDir      ContextKey = "runlogdir"
	ContextKeyWrapperLogPath ContextKey = "wrapperlogpath"
)
