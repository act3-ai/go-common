// Package testutil contains utilities for writing tests.
package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// AssertErrorIf asserts that err is not nil if wantErr is true, or nil if wantErr is false.
func AssertErrorIf(t *testing.T, wantErr bool, err error, msgAndArgs ...any) bool {
	t.Helper()
	if wantErr {
		return assert.Error(t, err, msgAndArgs...)
	}
	return assert.NoError(t, err, msgAndArgs...)
}

// AssertNilIf asserts that expected and actual are either both nil or both non-nil.
func AssertNilIf(t *testing.T, wantNil bool, actual any, msgAndArgs ...any) bool {
	t.Helper()
	if wantNil {
		return assert.Nil(t, actual, msgAndArgs...)
	}
	return assert.NotNil(t, actual, msgAndArgs...)
}
