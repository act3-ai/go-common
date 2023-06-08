package config

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// define errors for config
var (
	// ErrEnvVarNotFound is returned when an environment variable is not found (os.LookupEnv error)
	ErrEnvVarNotFound = errors.New("environment variable not found")
	// ErrParseEnvVar is returned when an environment variable is found but cannot be parsed
	ErrParseEnvVar = errors.New("error parsing environment variable")
)

// EnvOr grabs the env variable or the default
func EnvOr(name, def string) string {
	if name == "" {
		panic("name must not be empty")
	}
	envVal, ok := os.LookupEnv(name)
	if !ok {
		return def
	}
	return envVal
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

// EnvIntOr grabs the env variable as an int or the default
func EnvIntOr(name string, def int) int {
	if name == "" {
		panic("name must not be empty")
	}
	envVal, ok := os.LookupEnv(name)
	if !ok {
		return def
	}
	ret, err := strconv.Atoi(envVal)
	if err != nil {
		return def
	}
	return ret
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

// EnvBoolOr grabs the env variable as an int or the default
func EnvBoolOr(name string, def bool) bool {
	if name == "" {
		panic("name must not be empty")
	}
	envVal, ok := os.LookupEnv(name)
	if !ok {
		return def
	}
	ret, err := strconv.ParseBool(envVal)
	if err != nil {
		return def
	}
	return ret
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

// EnvArrayOr grabs the env variable as an array.  Returns an empty array if
func EnvArrayOr(name string, def []string, sep string) []string {
	if name == "" {
		panic("name must not be empty")
	}
	envVal, ok := os.LookupEnv(name)
	if !ok || envVal == "" {
		return def
	}
	return strings.Split(envVal, sep)
}

// EnvPathOr grabs the env variable as an array splitting on the default (OS specific) path list separator
func EnvPathOr(name string, def []string) []string {
	return EnvArrayOr(name, def, string(filepath.ListSeparator))
}

// EnvDurationOr grabs the env variable as a Duration or the default
func EnvDurationOr(name string, def time.Duration) time.Duration {
	if name == "" {
		panic("name must not be empty")
	}
	envVal, ok := os.LookupEnv(name)
	if !ok {
		return def
	}
	ret, err := time.ParseDuration(envVal)
	if err != nil {
		return def
	}
	return ret
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
