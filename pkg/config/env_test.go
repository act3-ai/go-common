package config

import (
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestStringEnv(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name        string
		envName     string
		envValue    string
		expectedErr error
	}{
		{
			name:        "env variable exists",
			envName:     "MY_STRING",
			envValue:    "hello",
			expectedErr: nil,
		},
		{
			name:        "env variable does not exist",
			envName:     "MY_OTHER_STRING",
			envValue:    "",
			expectedErr: ErrEnvVarNotFound,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up environment variable
			if !errors.Is(tc.expectedErr, ErrEnvVarNotFound) {
				t.Setenv(tc.envName, tc.envValue)
			}

			// Test Env
			val, err := Env(tc.envName)
			assert.Equal(t, tc.expectedErr, err)
			assert.Equal(t, tc.envValue, val)

			// Test EnvOr
			val = EnvOr(tc.envName, "default")
			if errors.Is(tc.expectedErr, ErrEnvVarNotFound) {
				assert.Equal(t, "default", val)
			} else {
				assert.Equal(t, tc.envValue, val)
			}
		})
	}
}

func TestQuantityEnv(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name        string
		envName     string
		envValue    resource.Quantity
		expectedErr error
	}{
		{
			name:        "env variable exists",
			envName:     "MY_QUANTITY",
			envValue:    resource.MustParse("1Gi"),
			expectedErr: nil,
		},
		{
			name:        "env variable does not exist",
			envName:     "MY_OTHER_QUANTITY",
			envValue:    resource.Quantity{},
			expectedErr: ErrEnvVarNotFound,
		},
		{
			name:        "parse error",
			envName:     "MY_WRONG_QUANTITY",
			envValue:    resource.Quantity{},
			expectedErr: ErrParseEnvVar,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up environment variable
			if !errors.Is(tc.expectedErr, ErrEnvVarNotFound) {
				if errors.Is(tc.expectedErr, ErrParseEnvVar) {
					t.Setenv(tc.envName, "wrong")
				} else {
					t.Setenv(tc.envName, tc.envValue.String())
				}
			}

			// Test EnvQuantity
			val, err := EnvQuantity(tc.envName)
			assert.Equal(t, tc.expectedErr, err)
			assert.Equal(t, tc.envValue, val)

			// Test EnvQuantityOr
			val = EnvQuantityOr(tc.envName, resource.MustParse("5Gi"))
			if err != nil {
				assert.Equal(t, resource.MustParse("5Gi"), val)
			} else {
				assert.Equal(t, tc.envValue, val)
			}
		})
	}
}

func TestIntEnv(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name        string
		envName     string
		envValue    int
		expectedErr error
	}{
		{
			name:        "env variable exists",
			envName:     "MY_INT",
			envValue:    5,
			expectedErr: nil,
		},
		{
			name:        "env variable does not exist",
			envName:     "MY_OTHER_INT",
			envValue:    0,
			expectedErr: ErrEnvVarNotFound,
		},
		{
			name:        "parse error",
			envName:     "MY_WRONG_INT",
			envValue:    0,
			expectedErr: ErrParseEnvVar,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up environment variable
			if !errors.Is(tc.expectedErr, ErrEnvVarNotFound) {
				if errors.Is(tc.expectedErr, ErrParseEnvVar) {
					t.Setenv(tc.envName, "wrong")
				} else {
					t.Setenv(tc.envName, strconv.Itoa(tc.envValue))
				}
			}

			// Test EnvInt
			val, err := EnvInt(tc.envName)
			assert.Equal(t, tc.expectedErr, err)
			assert.Equal(t, tc.envValue, val)

			// Test EnvIntOr
			val = EnvIntOr(tc.envName, 6)
			if err != nil {
				assert.Equal(t, 6, val)
			} else {
				assert.Equal(t, tc.envValue, val)
			}
		})
	}
}

func TestBoolEnv(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name        string
		envName     string
		envValue    bool
		expectedErr error
	}{
		{
			name:        "env variable exists and is true",
			envName:     "MY_BOOL",
			envValue:    true,
			expectedErr: nil,
		},
		{
			name:        "env variable exists and is false",
			envName:     "MY_OTHER_BOOL",
			envValue:    false,
			expectedErr: nil,
		},
		{
			name:        "env variable does not exist",
			envName:     "MY_NONEXISTENT_BOOL",
			envValue:    false,
			expectedErr: ErrEnvVarNotFound,
		},
		{
			name:        "parse error",
			envName:     "MY_WRONG_BOOL",
			envValue:    false,
			expectedErr: ErrParseEnvVar,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up environment variable
			if !errors.Is(tc.expectedErr, ErrEnvVarNotFound) {
				if errors.Is(tc.expectedErr, ErrParseEnvVar) {
					t.Setenv(tc.envName, "wrong")
				} else {
					t.Setenv(tc.envName, strconv.FormatBool(tc.envValue))
				}
			}

			// Test EnvBool
			val, err := EnvBool(tc.envName)
			assert.Equal(t, tc.expectedErr, err)
			assert.Equal(t, tc.envValue, val)

			// Test EnvBoolOr
			val = EnvBoolOr(tc.envName, true)
			if err != nil {
				assert.Equal(t, true, val)
			} else {
				assert.Equal(t, tc.envValue, val)
			}
		})
	}
}

func TestDurationEnv(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name        string
		envName     string
		envValue    time.Duration
		expectedErr error
	}{
		{
			name:        "env variable exists",
			envName:     "MY_DURATION",
			envValue:    5 * time.Second,
			expectedErr: nil,
		},
		{
			name:        "env variable does not exist",
			envName:     "MY_OTHER_DURATION",
			envValue:    0,
			expectedErr: ErrEnvVarNotFound,
		},
		{
			name:        "parse error",
			envName:     "MY_WRONG_DURATION",
			envValue:    0,
			expectedErr: ErrParseEnvVar,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up environment variable
			if !errors.Is(tc.expectedErr, ErrEnvVarNotFound) {
				if errors.Is(tc.expectedErr, ErrParseEnvVar) {
					t.Setenv(tc.envName, "wrong")
				} else {
					t.Setenv(tc.envName, tc.envValue.String())
				}
			}

			// Test EnvDuration
			val, err := EnvDuration(tc.envName)
			assert.Equal(t, tc.expectedErr, err)
			assert.Equal(t, tc.envValue, val)

			// Test EnvDurationOr
			val = EnvDurationOr(tc.envName, 6*time.Second)
			if err != nil {
				assert.Equal(t, 6*time.Second, val)
			} else {
				assert.Equal(t, tc.envValue, val)
			}
		})
	}
}

func TestArrayEnv(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name        string
		envName     string
		envValue    []string
		expectedErr error
	}{
		{
			name:        "env variable exists",
			envName:     "MY_ARRAY",
			envValue:    []string{"value1", "value2", "value3"},
			expectedErr: nil,
		},
		{
			name:        "env variable does not exist",
			envName:     "MY_OTHER_ARRAY",
			envValue:    nil,
			expectedErr: ErrEnvVarNotFound,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up environment variable
			if !errors.Is(tc.expectedErr, ErrEnvVarNotFound) {
				t.Setenv(tc.envName, strings.Join(tc.envValue, ","))
			}

			// Test EnvArray
			val, err := EnvArray(tc.envName, ",")
			assert.Equal(t, tc.expectedErr, err)
			assert.Equal(t, tc.envValue, val)

			// Test EnvArrayOr
			val = EnvArrayOr(tc.envName, []string{"default1", "default2"}, ",")
			if err != nil {
				assert.Equal(t, []string{"default1", "default2"}, val)
			} else {
				assert.Equal(t, tc.envValue, val)
			}
		})
	}
}
