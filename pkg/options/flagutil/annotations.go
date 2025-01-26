package flagutil

import (
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
