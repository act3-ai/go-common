package otel

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Config configures the initialization of OpenTelemetry. Typically configuration
// through environment variables is sufficient. If not, define the appropriate
// exporters and processors.
type Config struct {
	// Override auto-detect exporters from OTEL_* env variables.
	DisableEnvConfiguration bool

	// SpanProcessors are processors to prepend to the telemetry pipeline.
	SpanProcessors []sdktrace.SpanProcessor

	// LogProcessors are processors to prepend to the telemetry pipeline.
	LogProcessors []sdklog.Processor

	// MetricReaders are readers that collect metric data.
	MetricReaders []sdkmetric.Reader

	// Resource is the resource describing this component and runtime
	// environment.
	Resource *resource.Resource

	traceProvider *sdktrace.TracerProvider
	logProvider   *sdklog.LoggerProvider
	meterProvider *sdkmetric.MeterProvider
	propagator    propagation.TextMapPropagator
}

// Init sets up the global OpenTelemetry providers for tracing, logging, and
// metrics. It does not setup handling of telemetry errors, use otel.SetErrorHandler
// to do so.
func (c *Config) Init(ctx context.Context) (context.Context, error) {
	// Do not rely on otel.GetTextMapPropagator() - it's prone to change from a
	// random import.
	c.propagator = propagation.NewCompositeTextMapPropagator(
		propagation.Baggage{},
		propagation.TraceContext{},
	)
	otel.SetTextMapPropagator(c.propagator)

	// Inherit trace context from env if present.
	ctx = c.propagator.Extract(ctx, NewEnvCarrier(true))

	if c.Resource == nil {
		slog.WarnContext(ctx, "No OpenTelemetry resource defined, using fallback resource")
		c.Resource = fallbackResource(ctx)
	}

	if !c.DisableEnvConfiguration {
		if err := c.configureFromEnvironment(ctx); err != nil {
			return nil, fmt.Errorf("configuring exporters from environment: %w", err)
		}
	}

	if len(c.SpanProcessors) > 0 {
		traceOpts := make([]sdktrace.TracerProviderOption, 0, 1+len(c.SpanProcessors))
		traceOpts = append(traceOpts, sdktrace.WithResource(c.Resource))

		for _, sp := range c.SpanProcessors {
			traceOpts = append(traceOpts, sdktrace.WithSpanProcessor(sp))
		}

		c.traceProvider = sdktrace.NewTracerProvider(traceOpts...)
		otel.SetTracerProvider(c.traceProvider)
	}

	if len(c.LogProcessors) > 0 {
		logOpts := make([]sdklog.LoggerProviderOption, 0, 1+len(c.LogProcessors))
		logOpts = append(logOpts, sdklog.WithResource(c.Resource))

		for _, lp := range c.LogProcessors {
			logOpts = append(logOpts, sdklog.WithProcessor(lp))
		}

		c.logProvider = sdklog.NewLoggerProvider(logOpts...)
		// unlike traces and metrics we don't need to set a global logger provider,
		// not only  does otel not provide this but we use a slog bridge
		// and we're still able to shut down properly.
	}

	if len(c.MetricReaders) > 0 {
		meterOpts := make([]sdkmetric.Option, 0, 1+len(c.MetricReaders))
		meterOpts = append(meterOpts, sdkmetric.WithResource(c.Resource))

		for _, mr := range c.MetricReaders {
			meterOpts = append(meterOpts, sdkmetric.WithReader(mr))
		}

		c.meterProvider = sdkmetric.NewMeterProvider(meterOpts...)
		otel.SetMeterProvider(c.meterProvider)
	}

	return ctx, nil
}

// Shutdown shuts down the global OpenTelemetry providers, flushing any remaining
// data to the configured exporters.
func (c *Config) Shutdown(ctx context.Context) error {
	var errs []error
	flushCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
	defer cancel()
	if c.traceProvider != nil {
		if err := c.traceProvider.Shutdown(flushCtx); err != nil {
			errs = append(errs, fmt.Errorf("shutting down trace provider: %w", err))
		}
	}
	if c.logProvider != nil {
		if err := c.logProvider.Shutdown(flushCtx); err != nil {
			errs = append(errs, fmt.Errorf("shutting down log provider: %w", err))
		}
	}
	if c.meterProvider != nil {
		if err := c.meterProvider.Shutdown(flushCtx); err != nil {
			errs = append(errs, fmt.Errorf("shutting down metric provider: %w", err))
		}
	}

	return errors.Join(errs...) //nolint:wrapcheck
}

// configureFromEnvironment creates trace exporters, log exporters, and metric readers
// configured through environment variables.
func (c *Config) configureFromEnvironment(ctx context.Context) error {
	// span exporter from environment
	spanExp, err := autoexport.NewSpanExporter(ctx)
	if err != nil {
		return fmt.Errorf("configuring span exporter from environment variables: %w", err)
	}
	if spanExp != nil {
		// span processor from environment
		sp := sdktrace.NewBatchSpanProcessor(spanExp)
		c.SpanProcessors = append(c.SpanProcessors, sp)
	}

	// log exporter from environment
	logExp, err := autoexport.NewLogExporter(ctx)
	if err != nil {
		return fmt.Errorf("configuring log exporter from environment variables: %w", err)
	}
	if logExp != nil {
		// log processor from environment
		lp := sdklog.NewBatchProcessor(logExp)
		c.LogProcessors = append(c.LogProcessors, lp)
	}

	// metric exporter and reader from environment
	mr, err := autoexport.NewMetricReader(ctx)
	if err != nil {
		return fmt.Errorf("configuring metric exporter from environment variables: %w", err)
	}
	if mr != nil {
		c.MetricReaders = append(c.MetricReaders, mr)
	}

	return nil
}

// fallbackResouce is used by Init() if one is not explcitly provided in the Config.
func fallbackResource(ctx context.Context) *resource.Resource {
	r, _ := resource.New(
		ctx,
		resource.WithFromEnv(),      // Discover and provide attributes from OTEL_RESOURCE_ATTRIBUTES and OTEL_SERVICE_NAME environment variables.
		resource.WithTelemetrySDK(), // Discover and provide information about the OpenTelemetry SDK used.
		// resource.WithProcess(),      // Discover and provide process information.
		resource.WithOS(), // Discover and provide OS information.
		// resource.WithContainer(),    // Discover and provide container information.
		// resource.WithHost(), // Discover and provide host information.
	)
	return r
}
