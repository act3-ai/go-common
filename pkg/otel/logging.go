package otel

import (
	"context"

	sdklog "go.opentelemetry.io/otel/sdk/log"
)

type providerKey struct{}

// WithLoggerProvider returns a new context with the given LoggerProvider.
func WithLoggerProvider(ctx context.Context, provider *sdklog.LoggerProvider) context.Context {
	return context.WithValue(ctx, providerKey{}, provider)
}

// LoggerProvider returns the LoggerProvider from the context.
func LoggerProvider(ctx context.Context) *sdklog.LoggerProvider {
	var loggerProvider *sdklog.LoggerProvider
	if val := ctx.Value(providerKey{}); val != nil {
		loggerProvider = val.(*sdklog.LoggerProvider)
	} else {
		loggerProvider = sdklog.NewLoggerProvider()
	}
	return loggerProvider
}
