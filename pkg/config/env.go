package config

import (
	"os"
	"strconv"
)

// EnvOr grabs the env variable or the default
func EnvOr(name, def string) string {
	if v, ok := os.LookupEnv(name); ok {
		return v
	}
	return def
}

// EnvIntOr grabs the env variable as an int or the default
func EnvIntOr(name string, def int) int {
	if name == "" {
		return def
	}
	envVal := EnvOr(name, strconv.Itoa(def))
	ret, err := strconv.Atoi(envVal)
	if err != nil {
		return def
	}
	return ret
}

// EnvBoolOr grabs the env variable as an int or the default
func EnvBoolOr(name string, def bool) bool {
	if name == "" {
		return def
	}
	envVal := EnvOr(name, strconv.FormatBool(def))
	ret, err := strconv.ParseBool(envVal)
	if err != nil {
		return def
	}
	return ret
}
