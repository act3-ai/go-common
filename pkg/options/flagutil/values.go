package flagutil

import (
	"log/slog"
	"maps"
	"time"

	"github.com/spf13/pflag"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ValueOr returns either the flag value or the default.
func ValueOr[T any](f *pflag.Flag, flagValue T, def T) T {
	if !f.Changed {
		return def
	}
	src := valueSource(f)
	slog.Info("overriding config with "+src.Key+" value", //nolint:sloglint
		src,
		slog.Any("value", flagValue),
	)
	return flagValue
}

// DurationOr parses a string flag into a [metav1.Duration], returning def if the flag was not set.
func DurationOr(f *pflag.Flag, flagValue string, def *metav1.Duration) *metav1.Duration {
	if !f.Changed {
		return def
	}
	src := valueSource(f)
	pd, err := time.ParseDuration(flagValue)
	if err != nil {
		slog.Error("parsing duration "+src.Key+" value", //nolint:sloglint
			slog.Any("err", err),
			src,
			slog.String("value", flagValue),
		)
		return def
	}
	slog.Info("overriding config with "+src.Key+" value", //nolint:sloglint
		src,
		slog.String("value", flagValue),
	)
	return &metav1.Duration{Duration: pd}
}

// SetMapKeysOr uses values given to a string slice flag to set the keys of a configuration map.
//
// For keys already set in the configuration map, the current value will be preserved.
//
// For keys not found in the configuration map, they will be set to the empty value of T.
func SetMapKeysOr[T any](f *pflag.Flag, flagValue []string, in map[string]T) map[string]T {
	if !f.Changed {
		return in
	}
	src := valueSource(f)
	slog.Info("overriding config with "+src.Key+" value", //nolint:sloglint
		src,
		slog.Any("value", flagValue),
	)
	return setMapKeys(flagValue, in)
}

func setMapKeys[M ~map[K]V, K comparable, V any](keys []K, in M) M {
	out := make(M, len(keys))
	for _, t := range keys {
		if in == nil {
			var empty V
			out[t] = empty
		} else {
			out[t] = in[t]
		}
	}
	return out
}

// SetFunc describes a function to set map values.
type SetFunc[V1, V2 any] func(flagValue V1, entry *V2, ok bool) (skip bool)

// SetMapValuesOr uses values given to a map-style flag to update a configuration map as defined by the SetFunc.
func SetMapValuesOr[M1 ~map[K]V1, M2 ~map[K]V2, K comparable, V1, V2 any](
	f *pflag.Flag, flagValues M1, in M2, setter SetFunc[V1, V2]) M2 {
	if !f.Changed {
		return in
	}
	src := valueSource(f)
	slog.Info("overriding config with "+src.Key+" value", //nolint:sloglint
		src,
		slog.Any("value", flagValues),
	)
	return setMapValues(flagValues, in, setter)
}

// merges the flag values with the given map. check unit tests for implementation details.
func setMapValues[M1 ~map[K]V1, M2 ~map[K]V2, K comparable, V1, V2 any](
	flagValues M1, in M2, setter SetFunc[V1, V2]) M2 {
	out := maps.Clone(in)
	for k, flagValue := range flagValues {
		entry, ok := in[k]
		skip := setter(flagValue, &entry, ok)
		switch {
		case skip:
			// Skip updating entry if indicated
			continue
		case out == nil:
			out = M2{k: entry}
		default:
			out[k] = entry
		}
	}
	return out
}

// valueSource produces the source of the flag's value.
func valueSource(f *pflag.Flag) slog.Attr {
	envName, ok := GetFirstAnnotation(f, envOverrideAnno)
	if ok {
		return slog.String("env", envName)
	}
	return slog.String("flag", f.Name)
}
