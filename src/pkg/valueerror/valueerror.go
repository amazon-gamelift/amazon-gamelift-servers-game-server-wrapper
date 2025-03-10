/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package valueerror

import "fmt"

type ValueError struct {
	Value int
	Err   error
}

func New(value int, err error) *ValueError {
	return &ValueError{
		Value: value,
		Err:   err,
	}
}

func (ve *ValueError) Error() string {
	return fmt.Sprintf("value error: %s", ve.Err)
}

func (ve *ValueError) Unwrap() error {
	return ve.Err
}
