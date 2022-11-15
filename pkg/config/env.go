package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// EnvOr grabs the env variable or the default
func EnvOr(name, def string) string {
	if name == "" {
		return def
	}
	envVal, ok := os.LookupEnv(name)
	if !ok {
		return def
	}
	return envVal
}

// EnvIntOr grabs the env variable as an int or the default
func EnvIntOr(name string, def int) int {
	if name == "" {
		return def
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

// EnvBoolOr grabs the env variable as an int or the default
func EnvBoolOr(name string, def bool) bool {
	if name == "" {
		return def
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

// EnvArrayOr grabs the env variable as an array.  Returns an empty array if
func EnvArrayOr(name string, def []string, sep string) []string {
	if name == "" {
		return def
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
