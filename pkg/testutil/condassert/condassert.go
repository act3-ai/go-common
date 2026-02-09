package condassert

import (
	"github.com/stretchr/testify/assert"
)

// // Assertions is implemented by assert.Assertions and require.Assertions.
// type Assertions interface {
// 	// ErrorAssertions
// 	// PanicsAssertions
// 	// NilAssertions
// 	// ZeroAssertions
//
// 	Error(err error, msgAndArgs ...any) bool
// 	NoError(err error, msgAndArgs ...any) bool
//
// 	Panics(f assert.PanicTestFunc, msgAndArgs ...any) bool
// 	NotPanics(f assert.PanicTestFunc, msgAndArgs ...any) bool
//
// 	Nil(object any, msgAndArgs ...any) bool
// 	NotNil(object any, msgAndArgs ...any) bool
//
// 	Zero(i any, msgAndArgs ...any) bool
// 	NotZero(i any, msgAndArgs ...any) bool
// }

// ErrorAssertions checks errors.
type ErrorAssertions interface {
	Error(err error, msgAndArgs ...any) bool
	NoError(err error, msgAndArgs ...any) bool
}

// ErrorIf asserts that err is not nil if wantErr is true, or nil if wantErr is false.
func ErrorIf(a ErrorAssertions, wantErr bool, err error, msgAndArgs ...any) bool {
	if wantErr {
		return a.Error(err, msgAndArgs...)
	}
	return a.NoError(err, msgAndArgs...)
}

// PanicsAssertions checks panics.
type PanicsAssertions interface {
	Panics(f assert.PanicTestFunc, msgAndArgs ...any) bool
	NotPanics(f assert.PanicTestFunc, msgAndArgs ...any) bool
}

// PanicsIf asserts that function panics if wantPanic is true, or does not panic if wantPanic is false.
func PanicsIf(a PanicsAssertions, wantPanic bool, f assert.PanicTestFunc, msgAndArgs ...any) bool {
	if wantPanic {
		return a.Panics(f, msgAndArgs...)
	}
	return a.NotPanics(f, msgAndArgs...)
}

// NilAssertions checks nilness.
type NilAssertions interface {
	Nil(object any, msgAndArgs ...any) bool
	NotNil(object any, msgAndArgs ...any) bool
}

// NilIf asserts that the specified object is nil if wantNil is true, or not nil if wantNil is false.
func NilIf(a NilAssertions, wantNil bool, object any, msgAndArgs ...any) bool {
	if wantNil {
		return a.Nil(object, msgAndArgs...)
	}
	return a.NotNil(object, msgAndArgs...)
}

// ZeroAssertions checks against zero values.
type ZeroAssertions interface {
	Zero(i any, msgAndArgs ...any) bool
	NotZero(i any, msgAndArgs ...any) bool
}

// ZeroIf asserts that i is the zero value for its type if wantZero is true, or not if wantZero is true.
func ZeroIf(a ZeroAssertions, wantNil bool, i any, msgAndArgs ...any) bool {
	if wantNil {
		return a.Zero(i, msgAndArgs...)
	}
	return a.NotZero(i, msgAndArgs...)
}
