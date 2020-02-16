package assert

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/projecteru2/aa/errors"
)

// NilError asserts the error is nil.
func NilError(t *testing.T, err error) {
	Nil(t, err, errors.Stack(err))
}

// Error asserts there's an error.
func Error(t *testing.T, err error) {
	NotNil(t, err, errors.Stack(err))
}

// Nil asserts the obj is nil.
func Nil(t *testing.T, obj interface{}, msgAndArgs ...interface{}) {
	require.Nil(t, obj, msgAndArgs...)
}

// NotNil asserts the obj isn't nil.
func NotNil(t *testing.T, obj interface{}, msgAndArgs ...interface{}) {
	require.NotNil(t, obj, msgAndArgs...)
}

// True asserts b is true.
func True(t *testing.T, b bool, msgAndArgs ...interface{}) {
	Equal(t, true, b, msgAndArgs...)
}

// False asserts b is false.
func False(t *testing.T, b bool, msgAndArgs ...interface{}) {
	Equal(t, false, b, msgAndArgs...)
}

// Equal asserts exp and act is equal.
func Equal(t *testing.T, exp, act interface{}, msgAndArgs ...interface{}) {
	require.Equal(t, exp, act, msgAndArgs...)
}
