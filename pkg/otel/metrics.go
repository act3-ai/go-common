package otel

import (
	"context"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

type meterProviderKey struct{}

// WithMeterProvider returns a new context with the given MeterProvider.
func WithMeterProvider(ctx context.Context, provider *sdkmetric.MeterProvider) context.Context {
	return context.WithValue(ctx, meterProviderKey{}, provider)
}

// MeterProvider returns the MeterProviders from the context.
func MeterProvider(ctx context.Context) *sdkmetric.MeterProvider {
	var meterProvider *sdkmetric.MeterProvider
	if val := ctx.Value(meterProviderKey{}); val != nil {
		meterProvider = val.(*sdkmetric.MeterProvider)
	} else {
		meterProvider = sdkmetric.NewMeterProvider()
	}
	return meterProvider
}
