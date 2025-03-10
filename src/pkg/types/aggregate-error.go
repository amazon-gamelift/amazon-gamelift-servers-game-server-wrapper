/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package types

import (
	"fmt"
	"strings"
)

// AggregateError defines a collection of errors that occurred during an operation.
type AggregateError struct {
	errors []error
}

// Errors returns the slice of underlying errors contained in the AggregateError.
//
// Returns:
//   - []error: A slice containing all underlying errors
func (a AggregateError) Errors() []error {
	return a.errors
}

// Error returns a formatted string containing all underlying error messages.
//
// Returns:
//   - string: A formatted error message containing all underlying errors
func (a AggregateError) Error() string {
	errs := make([]string, 0)
	for _, e := range a.errors {
		errs = append(errs, e.Error())
	}

	return fmt.Sprintf("multiple errors: %s", strings.Join(errs, "\n"))
}
