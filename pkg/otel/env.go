package otel

import (
	"os"
	"strings"

	"go.opentelemetry.io/otel/propagation"
)

// EnvCarrier is used to inherit trace context from the environment.
type EnvCarrier struct {
	System bool
	Env    []string
}

// NewEnvCarrier initializes an EnvCarrier, optionally fetching from
// the system.
func NewEnvCarrier(system bool) *EnvCarrier {
	return &EnvCarrier{
		System: system,
	}
}

var _ propagation.TextMapCarrier = (*EnvCarrier)(nil)

// Get returns the value for a key.
func (c *EnvCarrier) Get(key string) string {
	envName := strings.ToUpper(key)
	for _, env := range c.Env {
		env, val, ok := strings.Cut(env, "=")
		if ok && env == envName {
			return val
		}
	}
	if c.System {
		if envVal := os.Getenv(envName); envVal != "" {
			return envVal
		}
	}
	return ""
}

// Set adds a key value pair to the environment.
func (c *EnvCarrier) Set(key, val string) {
	c.Env = append(c.Env, strings.ToUpper(key)+"="+val)
}

// Keys returns all keys in the environment.
func (c *EnvCarrier) Keys() []string {
	keys := make([]string, 0, len(c.Env))
	for _, env := range c.Env {
		env, _, ok := strings.Cut(env, "=")
		if ok {
			keys = append(keys, env)
		}
	}
	return keys
}
