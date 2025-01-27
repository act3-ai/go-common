// Package logutil defines basic logging utilities.
package logutil

import "log/slog"

// errKey is the key used for errors in [log/slog] attributes.
var errKey = "err"

// ErrKey produces the key used when logging errors.
func ErrKey() string {
	return errKey
}

// SetErrKey overrides the default key used when logging errors.
func SetErrKey(key string) {
	errKey = key
}

// Err produces a [log/slog.Attr] for errors.
func Err(err error) slog.Attr {
	return slog.Any(errKey, err)
}
