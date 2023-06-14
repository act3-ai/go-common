package config

import (
	"errors"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/resource"
)

// TestStringEnv tests the StringEnv helper struct and its methods
func TestNewEnvStruct(t *testing.T) {
	testEnvStruct := NewEnvStruct()

	// string env
	testStringName := "MY_STRING"
	testStringVal := "hello"
	t.Setenv(testStringName, testStringVal)
	testString := new(string)
	testEnvStruct.AddString(testString, testStringName)

	// int env
	testIntName := "MY_INT"
	testIntVal := 1
	t.Setenv(testIntName, strconv.Itoa(testIntVal))
	testInt := new(int)
	testEnvStruct.AddInt(testInt, testIntName)

	// bool env
	testBoolName := "MY_BOOL"
	testBoolVal := true
	t.Setenv(testBoolName, strconv.FormatBool(testBoolVal))
	testBool := new(bool)
	testEnvStruct.AddBool(testBool, testBoolName)

	// duration env
	testDurationName := "MY_DURATION"
	testDurationVal := 1 * time.Second
	t.Setenv(testDurationName, testDurationVal.String())
	testDuration := new(time.Duration)
	testEnvStruct.AddDuration(testDuration, testDurationName)

	// quantity env
	testQuantityName := "MY_QUANTITY"
	testQuantityVal := resource.MustParse("1Gi")
	t.Setenv(testQuantityName, testQuantityVal.String())
	testQuantity := new(resource.Quantity)
	testEnvStruct.AddQuantity(testQuantity, testQuantityName)

	// string array env
	testStringArrayName := "MY_STRING_ARRAY"
	testStringArrayVal := []string{"hello", "world"}
	testSep := ","
	t.Setenv(testStringArrayName, strings.Join(testStringArrayVal, testSep))
	testStringArray := new([]string)
	testEnvStruct.AddStringArray(testStringArray, testStringArrayName, testSep)

	// path env
	testPathName := "MY_PATH"
	tmpVal := []string{"hello", "world"}
	testPathVal := strings.Join(tmpVal, string(filepath.ListSeparator))
	t.Setenv(testPathName, testPathVal)
	testPath := new([]string)
	testEnvStruct.AddPath(testPath, testPathName)

	// Test lookup
	err := testEnvStruct.EnvOverrides()
	assert.NoError(t, err)
}

func TestEnvStructHandlers(t *testing.T) {
	testEnvStruct := NewEnvStruct()

	handleSuccessFunc := func() {
		t.Log("handleSuccess")
	}
	testEnvStruct.SetHandleSuccess(handleSuccessFunc)

	handleLookupFunc := func(err error) error {
		t.Log("handleLookupErr")
		assert.Equal(t, ErrEnvVarNotFound, err)
		// changing this to nil allows us to verify that we changed the internal handler func
		return nil
	}
	testEnvStruct.SetHandleLookupErr(handleLookupFunc)

	handleParseFunc := func(err error) error {
		t.Log("handleParseErr")
		assert.Equal(t, ErrParseEnvVar, err)
		// changing this to nil allows us to verify that we changed the internal handler func
		return nil
	}
	testEnvStruct.SetHandleParseErr(handleParseFunc)

	// add string that fails on lookup
	testEnvStruct.AddString(new(string), "MY_STRING")

	// add int that fails on parse
	testIntName := "MY_INT"
	t.Setenv(testIntName, "NaN")
	testInt := new(int)
	testEnvStruct.AddInt(testInt, testIntName)

	err := testEnvStruct.EnvOverrides()
	assert.NoError(t, err)

}

func TestLookupString(t *testing.T) {
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

			// Create helper
			h := &helper{
				name: tc.envName,
				pntr: new(string),
			}

			// Test lookupString
			err := h.lookupString()
			assert.Equal(t, tc.expectedErr, err)
			if err == nil {
				assert.Equal(t, tc.envValue, *(h.pntr.(*string)))
			}
		})
	}
}

func TestLookupQuantity(t *testing.T) {
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

			// Create helper
			h := &helper{
				name: tc.envName,
				pntr: new(resource.Quantity),
			}

			// Test lookupQuantity
			err := h.lookupQuantity()
			assert.Equal(t, tc.expectedErr, err)
			if err == nil {
				assert.Equal(t, tc.envValue, *(h.pntr.(*resource.Quantity)))
			}
		})
	}
}

func TestLookupInt(t *testing.T) {
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

			// Create helper
			h := &helper{
				name: tc.envName,
				pntr: new(int),
			}

			// Test lookupInt
			err := h.lookupInt()
			assert.Equal(t, tc.expectedErr, err)
			if err == nil {
				assert.Equal(t, tc.envValue, *(h.pntr.(*int)))
			}
		})
	}
}

func TestLookupBool(t *testing.T) {
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

			// Create helper
			h := &helper{
				name: tc.envName,
				pntr: new(bool),
			}

			// Test lookupBool
			err := h.lookupBool()
			assert.Equal(t, tc.expectedErr, err)
			if err == nil {
				assert.Equal(t, tc.envValue, *(h.pntr.(*bool)))
			}
		})
	}
}

func TestLookupDuration(t *testing.T) {
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

			// Create helper
			h := &helper{
				name: tc.envName,
				pntr: new(time.Duration),
			}

			// Test lookupDuration
			err := h.lookupDuration()
			assert.Equal(t, tc.expectedErr, err)
			if err == nil {
				assert.Equal(t, tc.envValue, *(h.pntr.(*time.Duration)))
			}
		})
	}
}

func TestLookupStringArray(t *testing.T) {
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

			// Create helper
			h := &helper{
				name: tc.envName,
				pntr: new([]string),
				sep:  ",",
			}

			// Test lookupArray
			err := h.lookupStringArray()
			assert.Equal(t, tc.expectedErr, err)
			if err == nil {
				assert.Equal(t, tc.envValue, *(h.pntr.(*[]string)))
			}
		})
	}
}

func TestLookupPath(t *testing.T) {
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
				t.Setenv(tc.envName, strings.Join(tc.envValue, string(filepath.ListSeparator)))
			}

			// Create helper
			h := &helper{
				name: tc.envName,
				pntr: new([]string),
				sep:  string(filepath.ListSeparator),
			}

			// Test lookupPath
			err := h.lookupPath()
			assert.Equal(t, tc.expectedErr, err)
			if err == nil {
				assert.Equal(t, tc.envValue, *(h.pntr.(*[]string)))
			}
		})
	}
}
