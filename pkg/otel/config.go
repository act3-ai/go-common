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
// through environment variables is sufficient. If not, define processors and exporters
// with the appropriate options as needed.
type Config struct {
	// Override auto-detect exporters from OTEL_* env variables.
	DisableEnvConfiguration bool

	// SpanProcessors are processors to prepend to the telemetry pipeline.
	SpanProcessors []sdktrace.SpanProcessor

	// BatchedTraceExporters are exporters that receive spans in batches, after
	// the spans have ended.
	BatchedTraceExporters []sdktrace.SpanExporter

	// LogProcessors are processors to prepend to the telemetry pipeline.
	LogProcessors []sdklog.Processor

	// BatchedLogExporters are exporters that receive logs in batches.
	BatchedLogExporters []sdklog.Exporter

	// MetricReaders are readers that collect metric data.
	MetricReaders []sdkmetric.Reader

	// BatchedMetricExporters are exporters that receive metrics in batches.
	BatchedMetricExporters []sdkmetric.Exporter

	// Resource is the resource describing this component and runtime
	// environment.
	Resource *resource.Resource

	traceProvider *sdktrace.TracerProvider
	logProvider   *sdklog.LoggerProvider
	meterProvider *sdkmetric.MeterProvider
	propagator    propagation.TextMapPropagator
}

// Resource is the globally configured resource, allowing it to be provided
// to dynamically allocated log/trace providers at runtime.
var Resource *resource.Resource

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

	// Set up the global resource so we can pass it into dynamically allocated
	// log/trace providers at runtime.
	Resource = c.Resource

	if !c.DisableEnvConfiguration {
		if err := c.configureFromEnvironment(ctx); err != nil {
			return nil, fmt.Errorf("configuring exporters from environment: %w", err)
		}
	}

	// Set up trace provider if configured.
	if len(c.BatchedTraceExporters) > 0 {
		traceOpts := []sdktrace.TracerProviderOption{
			sdktrace.WithResource(c.Resource),
		}

		for _, exporter := range c.BatchedTraceExporters {
			processor := sdktrace.NewBatchSpanProcessor(exporter)
			c.SpanProcessors = append(c.SpanProcessors, processor)
			traceOpts = append(traceOpts, sdktrace.WithSpanProcessor(processor))
		}
		c.traceProvider = sdktrace.NewTracerProvider(traceOpts...)
		otel.SetTracerProvider(c.traceProvider)
	}

	// Set up a log provider if configured.
	if len(c.BatchedLogExporters) > 0 {
		logOpts := []sdklog.LoggerProviderOption{
			sdklog.WithResource(c.Resource),
		}

		for _, exp := range c.BatchedLogExporters {
			processor := sdklog.NewBatchProcessor(exp)
			c.LogProcessors = append(c.LogProcessors, processor)
			logOpts = append(logOpts, sdklog.WithProcessor(processor))
		}
		c.logProvider = sdklog.NewLoggerProvider(logOpts...)
		// unlike traces and metrics we don't need to set a global logger provider,
		// not only  does otel not provide this but we use a slog bridge
		// and we're still able to shut down properly.
	}

	// Set up a metric provider if configured.
	if len(c.BatchedMetricExporters) > 0 {
		meterOpts := []sdkmetric.Option{
			sdkmetric.WithResource(c.Resource),
		}

		for _, exp := range c.BatchedMetricExporters {
			reader := sdkmetric.NewPeriodicReader(exp)
			c.MetricReaders = append(c.MetricReaders, reader)
			meterOpts = append(meterOpts, sdkmetric.WithReader(reader))
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

	return errors.Join(errs...)
}

// configureFromEnvironment creates trace exporters, log exporters, and metric readers
// configured through environment variables.
func (c *Config) configureFromEnvironment(ctx context.Context) error {
	// Spans
	spanExp, err := autoexport.NewSpanExporter(ctx)
	if err != nil {
		return fmt.Errorf("configuring span exporter from environment variables: %w", err)
	}
	if spanExp != nil {
		c.BatchedTraceExporters = append(c.BatchedTraceExporters,
			// Filter out unfinished spans to avoid confusing external systems.
			FilterLiveSpansExporter{spanExp})
	}

	// Logs
	logExp, err := autoexport.NewLogExporter(ctx)
	if err != nil {
		return fmt.Errorf("configuring log exporter from environment variables: %w", err)
	}
	if logExp != nil {
		c.BatchedLogExporters = append(c.BatchedLogExporters, logExp)
	}

	// Metrics
	metricReader, err := autoexport.NewMetricReader(ctx)
	if err != nil {
		return fmt.Errorf("configuring metric exporter from environment variables: %w", err)
	}
	if metricReader != nil {
		c.MetricReaders = append(c.MetricReaders, metricReader)
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
