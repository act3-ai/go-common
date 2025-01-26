package flagutil

import (
	"errors"
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

// ParseEnvOverrides receives a flag set after it has been parsed and
// sets the flag values to environment variables if the flag defines an
// "env" annotation.
//
// Any parsing errors are returned.
func ParseEnvOverrides(flagSet *pflag.FlagSet) error {
	var parseErrs []error
	flagSet.VisitAll(func(f *pflag.Flag) {
		// Do not load env if this flag was changed, flag values should win
		if f.Changed {
			return
		}
		// Skip flags without an "env" annotation
		envName, ok := GetFirstAnnotation(f, envAnno)
		if !ok {
			return
		}
		// Lookup environment variable, skip if unset.
		envString, ok := os.LookupEnv(envName)
		if !ok {
			return
		}
		err := f.Value.Set(envString)
		if err != nil {
			parseErrs = append(parseErrs, fmt.Errorf("invalid value %q for %q env variable: %w", envString, envName, err))
			return
		}
		// Set changed to true to signal that the flag value should be used.
		f.Changed = true
		SetAnnotation(f, envOverrideAnno, envName)
	})
	if len(parseErrs) > 0 {
		return errors.Join(parseErrs...) //nolint:wrapcheck // the errors being joined are already wrapped
	}
	return nil
}
