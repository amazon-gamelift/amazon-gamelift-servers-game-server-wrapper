/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package observability

import "encoding/base64"

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func basicAuthHeaders(username, password string) map[string]string {
	return map[string]string{
		"Authorization": "Basic " + basicAuth(username, password),
	}
}
