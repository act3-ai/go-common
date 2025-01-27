package flagutil

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
)

const (
	// envAnno is the key for the environment variable annotation.
	envAnno = "flagutil_env_name"

	// envOverrideAnno signals that the flag's value came from an environment variable.
	envOverrideAnno = "flagutil_value_from_env"
)

// SetEnvName sets the name of an environment variable used to override the flag's value
// in the ParseEnvOverrides function.
func SetEnvName(f *pflag.Flag, envName string) {
	if envName == "" {
		panic("empty envName")
	}
	SetAnnotation(f, envAnno, envName)
}

// GetEnvName gets the name of the environment variable used to override the flag's value
// in the ParseEnvOverrides function.
//
// An empty string means the annotation is not set.
func GetEnvName(f *pflag.Flag) string {
	return GetFirstAnnotationOr(f, envAnno, "")
}

// ParseEnvOverrides overrides the flag from an environment variable,
// if it has a defined environment variable and the flag was not already set.
//
// ParseEnvOverride should be called flag parsing, otherwise the
// environment variable will take precedence over the flag's value.
//
// Flag environment variables can be set with [SetEnvName].
// The flag creation functions in pkg/options/flags.go set an
// environment variable for the flag if Option.Env is set.
//
// If the environment variable cannot be parsed, an error will returned.
// Errors will be of type [EnvParseError] which allows the calling function to access
// the name of the environment variable, its value, and the underlying parse error
// if needed for error handling.
func ParseEnvOverrides(f *pflag.Flag) error {
	// Do not load env if this flag was changed, flag values should win
	if f.Changed {
		return nil
	}
	// Skip flags without an "env" annotation
	envName, ok := GetFirstAnnotation(f, envAnno)
	if !ok {
		return nil
	}
	// Lookup environment variable, skip if unset.
	envString, ok := os.LookupEnv(envName)
	if !ok {
		return nil
	}
	err := f.Value.Set(envString)
	if err != nil {
		return NewEnvParseError(envName, envString, err)
	}
	// Set changed to true to signal that the flag value should be used.
	f.Changed = true
	SetAnnotation(f, envOverrideAnno, envName)
	return nil
}

// EnvParseError represents an environment variable parsing error.
type EnvParseError interface {
	error
	// EnvName produces the invalid environment variable's name.
	EnvName() string
	// EnvValue produces the invalid environment variable's value.
	EnvValue() string
}

// NewEnvParseError creates an environment variable parsing error.
func NewEnvParseError(envName, envValue string, cause error) EnvParseError {
	if cause == nil {
		return nil
	}
	return envParseError{
		envName:  envName,
		envValue: envValue,
		cause:    cause,
	}
}

// envParseError represents an environment variable parsing error.
type envParseError struct {
	envName  string // environment variable name
	envValue string // environment variable value
	cause    error  // underlying error
}

// Error implements [error].
func (err envParseError) Error() string {
	// matches the format used in pflag.FlagSet.Set.
	return fmt.Sprintf("invalid value %q for %q env variable: %s",
		err.envValue, err.envName, err.cause.Error())
}

// Unwrap implements [error].
func (err envParseError) Unwrap() error {
	return err.cause
}

// EnvName produces the invalid environment variable's name.
func (err envParseError) EnvName() string {
	return err.envName
}

// EnvValue produces the invalid environment variable's value.
func (err envParseError) EnvValue() string {
	return err.envValue
}
