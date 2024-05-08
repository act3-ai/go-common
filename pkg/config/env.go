package config

import (
	"gitlab.com/act3-ai/asce/go-common/pkg/config/env"
)

// define errors for config
var (
	// ErrEnvVarNotFound is returned when an environment variable is not found (os.LookupEnv error)
	ErrEnvVarNotFound = env.ErrEnvVarNotFound
	// ErrParseEnvVar is returned when an environment variable is found but cannot be parsed
	ErrParseEnvVar = env.ErrParseEnvVar
)

// Redirect functions for backwards compatibility
var (
	// EnvOr grabs the env variable or the default
	EnvOr = env.Or

	// Env returns the named env variable if it exists,
	// otherwise returns empty string and an ErrEnvVarNotFound error.
	Env = env.OrError

	// EnvIntOr grabs the env variable as an int or the default
	EnvIntOr = env.IntOr
	// EnvInt returns the named env variable if it exists,
	// otherwise returns 0 and either an ErrEnvVarNotFound or an ErrParseEnvVar error.
	EnvInt = env.IntOrError

	// EnvBoolOr grabs the env variable as an int or the default
	EnvBoolOr = env.BoolOr

	// EnvBool returns the named env variable if it exists,
	// otherwise returns false and either an ErrEnvVarNotFound or an ErrParseEnvVar error.
	EnvBool = env.BoolOrError

	// EnvArrayOr grabs the env variable as an array.  Returns an empty array if
	EnvArrayOr = env.ArrayOr

	// EnvPathOr grabs the env variable as an array splitting on the default (OS specific) path list separator
	EnvPathOr = env.PathOr

	// EnvDurationOr grabs the env variable as a Duration or the default
	EnvDurationOr = env.DurationOr

	// EnvDuration returns the named env variable if it exists,
	// otherwise returns 0 and either an ErrEnvVarNotFound or an ErrParseEnvVar error.
	EnvDuration = env.DurationOrError
)
