package config

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
)

// define errors for config
var (
	// ErrEnvVarNotFound is returned when an environment variable is not found (os.LookupEnv error)
	ErrEnvVarNotFound = errors.New("environment variable not found")
	// ErrParseEnvVar is returned when an environment variable is found but cannot be parsed
	ErrParseEnvVar = errors.New("error parsing environment variable")
)

// StringEnv is a helper struct for organizing environment variables to be parsed as strings.
// Range over a slice of these to populate your config struct.
//
// Example:
//
//	var envs = []StringEnv{
//		{&config.MyString, "MY_STRING"},
//		{&config.MyOtherString, "MY_OTHER_STRING"},
//	}
type StringEnv struct {
	Data *string
	Name string
}

// EnvOr grabs the env variable or the default
func EnvOr(name, def string) string {
	if val, err := Env(name); err == nil {
		return val
	}
	return def
}

// Env returns the named env variable if it exists,
// otherwise returns empty string and an ErrEnvVarNotFound error.
func Env(name string) (string, error) {
	if name == "" {
		panic("name must not be empty")
	}
	envVal, ok := os.LookupEnv(name)
	if !ok {
		return "", ErrEnvVarNotFound
	}
	return envVal, nil
}

// QuantityEnv is a helper struct for organizing environment variables to be parsed as resource.Quantity.
// Range over a slice of these to populate your config struct.
//
// Example:
//
//	var envs = []QuantityEnv{
//		{&config.MyQuantity, "MY_QUANTITY"},
//		{&config.MyOtherQuantity, "MY_OTHER_QUANTITY"},
//	}
type QuantityEnv struct {
	Data *resource.Quantity
	Name string
}

// EnvQuantityOr grabs the env variable as a resource.Quantity or the default
func EnvQuantityOr(name string, def resource.Quantity) resource.Quantity {
	if val, err := EnvQuantity(name); err == nil {
		return val
	}
	return def
}

// EnvQuantity returns the named env variable if it exists,
// otherwise returns the default Quantity{} and an ErrEnvVarNotFound error.
func EnvQuantity(name string) (resource.Quantity, error) {
	if name == "" {
		panic("name must not be empty")
	}
	envVal, ok := os.LookupEnv(name)
	if !ok {
		return resource.Quantity{}, ErrEnvVarNotFound
	}
	parsedVal, err := resource.ParseQuantity(envVal)
	if err != nil {
		return resource.Quantity{}, ErrParseEnvVar
	}
	return parsedVal, nil
}

// IntEnv is a helper struct for organizing environment variables to be parsed as ints.
// Range over a slice of these to populate your config struct.
//
// Example:
//
//	var envs = []IntEnv{
//		{&config.MyInt, "MY_INT"},
//		{&config.MyOtherInt, "MY_OTHER_INT"},
//	}
type IntEnv struct {
	Data *int
	Name string
}

// EnvIntOr grabs the env variable as an int or the default
func EnvIntOr(name string, def int) int {
	if val, err := EnvInt(name); err == nil {
		return val
	}
	return def
}

// EnvInt returns the named env variable if it exists,
// otherwise returns 0 and either an ErrEnvVarNotFound or an ErrParseEnvVar error.
func EnvInt(name string) (int, error) {
	if name == "" {
		panic("name must not be empty")
	}
	envVal, ok := os.LookupEnv(name)
	if !ok {
		return 0, ErrEnvVarNotFound
	}
	parsedVal, err := strconv.Atoi(envVal)
	if err != nil {
		return 0, ErrParseEnvVar
	}
	return parsedVal, nil
}

// BoolEnv is a helper struct for organizing environment variables to be parsed as bools.
// Range over a slice of these to populate your config struct.
//
// Example:
//
//	var envs = []BoolEnv{
//		{&config.MyBool, "MY_BOOL"},
//		{&config.MyOtherBool, "MY_OTHER_BOOL"},
//	}
type BoolEnv struct {
	Data *bool
	Name string
}

// EnvBoolOr grabs the env variable as an int or the default
func EnvBoolOr(name string, def bool) bool {
	if val, err := EnvBool(name); err == nil {
		return val
	}
	return def
}

// EnvBool returns the named env variable if it exists,
// otherwise returns false and either an ErrEnvVarNotFound or an ErrParseEnvVar error.
func EnvBool(name string) (bool, error) {
	if name == "" {
		panic("name must not be empty")
	}
	envVal, ok := os.LookupEnv(name)
	if !ok {
		return false, ErrEnvVarNotFound
	}
	parsedVal, err := strconv.ParseBool(envVal)
	if err != nil {
		return false, ErrParseEnvVar
	}
	return parsedVal, nil
}

// StringArrayEnv is a helper struct for organizing environment variables to be parsed as string arrays.
// Range over a slice of these to populate your config struct.
// This struct is also used for EnvPath and EnvPathOr
//
// Example:
//
//	var envs = []StringArrayEnv{
//		{&config.MyStringArray, "MY_STRING_ARRAY"},
//		{&config.MyOtherStringArray, "MY_OTHER_STRING_ARRAY"},
//	}
type StringArrayEnv struct {
	Data *[]string
	Name string
}

// EnvArrayOr grabs the env variable as an array.  Returns an empty array if
func EnvArrayOr(name string, def []string, sep string) []string {
	if val, err := EnvArray(name, sep); err == nil {
		return val
	}
	return def
}

// EnvArray returns the named env variable if it exists,
// otherwise returns nil and an ErrEnvVarNotFound error.
func EnvArray(name string, sep string) ([]string, error) {
	if name == "" {
		panic("name must not be empty")
	}
	envVal, ok := os.LookupEnv(name)
	if !ok {
		return nil, ErrEnvVarNotFound
	}
	return strings.Split(envVal, sep), nil
}

// EnvPathOr grabs the env variable as an array splitting on the default (OS specific) path list separator
func EnvPathOr(name string, def []string) []string {
	return EnvArrayOr(name, def, string(filepath.ListSeparator))
}

// EnvPath returns the named env variable if it exists,
// otherwise returns nil and an ErrEnvVarNotFound error.
func EnvPath(name string) ([]string, error) {
	return EnvArray(name, string(filepath.ListSeparator))
}

// DurationEnv is a helper struct for organizing environment variables to be parsed as time.Durations.
// Range over a slice of these to populate your config struct.
//
// Example:
//
//	var envs = []DurationEnv{
//		{&config.MyDuration, "MY_DURATION"},
//		{&config.MyOtherDuration, "MY_OTHER_DURATION"},
//	}
type DurationEnv struct {
	Data *time.Duration
	Name string
}

// EnvDurationOr grabs the env variable as a Duration or the default
func EnvDurationOr(name string, def time.Duration) time.Duration {
	if val, err := EnvDuration(name); err == nil {
		return val
	}
	return def
}

// EnvDuration returns the named env variable if it exists,
// otherwise returns 0 and either an ErrEnvVarNotFound or an ErrParseEnvVar error.
func EnvDuration(name string) (time.Duration, error) {
	if name == "" {
		panic("name must not be empty")
	}
	envVal, ok := os.LookupEnv(name)
	if !ok {
		return 0, ErrEnvVarNotFound
	}
	parsedVal, err := time.ParseDuration(envVal)
	if err != nil {
		return 0, ErrParseEnvVar
	}
	return parsedVal, nil
}
