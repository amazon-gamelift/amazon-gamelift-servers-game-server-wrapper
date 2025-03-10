/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package validate

import (
	"errors"
	"reflect"

	"github.com/go-playground/validator/v10"
)

var myValidator = validator.New()

// Validate performs validation on a given struct.
// It ensures the input is non-nil and validates its fields based on
// defined validation tags.
//
// Parameters:
//   - structure: interface{} - The struct to be validated
//
// Returns:
//   - error: Validation errors if any occur
func Validate(structure interface{}) error {
	if structure == nil {
		return errors.New("validation failed: input is nil")
	}
	if reflect.ValueOf(structure).Kind() != reflect.Struct {
		return errors.New("validation failed: input must be a struct")
	}
	return myValidator.Struct(structure)
}
