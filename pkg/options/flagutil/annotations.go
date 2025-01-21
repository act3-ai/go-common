package flagutil

import (
	"log/slog"
	"os"

	"github.com/spf13/pflag"
)

// SetAnnotation sets the flag's annotations for the given key.
func SetAnnotation(f *pflag.Flag, key string, values ...string) {
	if f.Annotations == nil {
		f.Annotations = map[string][]string{key: values}
	} else {
		f.Annotations[key] = values
	}
}

// GetFirstAnnotation returns the first annotation for the key, if it exists,
// and a boolean indicating if the annotation was found.
func GetFirstAnnotation(f *pflag.Flag, key string) (string, bool) {
	if f == nil || f.Annotations == nil || len(f.Annotations[key]) == 0 {
		return "", false
	}
	return f.Annotations[key][0], true
}

// GetFirstAnnotationOr returns the first annotation for the key, if it exists,
// or the default value given.
func GetFirstAnnotationOr(f *pflag.Flag, key string, def string) string {
	v, ok := GetFirstAnnotation(f, key)
	if !ok {
		return def
	}
	return v
}

// ParseEnvOverrides receives a flag set after it has been parsed and
// sets the flag values to environment variables if the flag defines an
// "env" annotation.
//
// Any parsing errors are logged at slog.LevelWarn and are discarded.
func ParseEnvOverrides(flagSet *pflag.FlagSet, envAnnoKey string) {
	flagSet.VisitAll(func(f *pflag.Flag) {
		// Do not load env if this flag was changed, flag values should win
		if f.Changed {
			return
		}
		// Skip flags without an "env" annotation
		envName, ok := GetFirstAnnotation(f, envAnnoKey)
		if !ok {
			return
		}
		// Lookup environment variable, skip if unset.
		envString, ok := os.LookupEnv(envName)
		if !ok {
			return
		}
		slog.Info("setting flag from env variable", slog.String("env", envName), slog.String("flag", f.Name))
		err := f.Value.Set(envString)
		if err != nil {
			slog.Warn("parsing env variable",
				slog.String("env", envName),
				slog.Any("err", err),
			)
		}
		// Set changed to true to signal that the flag value should be used.
		f.Changed = true
	})
}
