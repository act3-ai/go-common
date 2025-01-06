package otel

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"

	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
)

// TODO: Currently only supports trace exporters. Much of the plumbing
// for logs and metrics remains, but commented out.

// Config configures the initialization of OpenTelemetry.
type Config struct {
	// Override auto-detect exporters from OTEL_* env variables.
	DisableEnvConfiguration bool

	// SpanProcessors are processors to prepend to the telemetry pipeline.
	SpanProcessors []sdktrace.SpanProcessor

	// BatchedTraceExporters are exporters that receive spans in batches, after
	// the spans have ended.
	BatchedTraceExporters []sdktrace.SpanExporter

	// LiveTraceExporters are exporters that can receive updates for spans at
	// runtime, rather than waiting until the span ends.
	LiveTraceExporters []sdktrace.SpanExporter

	// LogProcessors are processors to prepend to the telemetry pipeline.
	LogProcessors []sdklog.Processor

	// BatchedLogExporters are exporters that receive logs in batches, after
	// the logs have ended.
	BatchedLogExporters []sdklog.Exporter

	// LiveLogExporters are exporters that receive logs in batches of ~100ms.
	LiveLogExporters []sdklog.Exporter

	// LiveMetricExporters are exporters that receive metrics in batches of ~1s.
	// LiveMetricExporters []sdkmetric.Exporter

	// Resource is the resource describing this component and runtime
	// environment.
	Resource *resource.Resource

	traceProvider *sdktrace.TracerProvider
	propagator    propagation.TextMapPropagator
}

// Resource is the globally configured resource, allowing it to be provided
// to dynamically allocated log/trace providers at runtime.
var Resource *resource.Resource

// var LogProcessors = []sdklog.Processor{}
// var MetricExporters = []sdkmetric.Exporter{}

// Init sets up the global OpenTelemetry providers tracing, logging, and
// someday metrics providers. It is called by the CLI, the engine, and the
// container shim, so it needs to be versatile.
func Init(ctx context.Context, cfg *Config) (context.Context, error) {
	// Do not rely on otel.GetTextMapPropagator() - it's prone to change from a
	// random import.
	cfg.propagator = propagation.NewCompositeTextMapPropagator(
		propagation.Baggage{},
		propagation.TraceContext{},
	)
	otel.SetTextMapPropagator(cfg.propagator)

	// Inherit trace context from env if present.
	ctx = cfg.propagator.Extract(ctx, NewEnvCarrier(true))

	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		logger.FromContext(ctx).ErrorContext(ctx, "failed to emit telemetry", "error", err)
	}))

	if cfg.Resource == nil {
		cfg.Resource = fallbackResource(ctx)
	}

	// Set up the global resource so we can pass it into dynamically allocated
	// log/trace providers at runtime.
	Resource = cfg.Resource

	if !cfg.DisableEnvConfiguration {
		spanExp, err := ConfiguredSpanExporter(ctx)
		if err != nil {
			return nil, fmt.Errorf("configuring span exporter from environment variables: %w", err)
		}
		if spanExp != nil {
			val, exists := os.LookupEnv("OTEL_EXPORTER_OTLP_TRACES_LIVE")
			if exists && val != "" {
				cfg.LiveTraceExporters = append(cfg.LiveTraceExporters, spanExp)
			} else {
				cfg.BatchedTraceExporters = append(cfg.BatchedTraceExporters,
					// Filter out unfinished spans to avoid confusing external systems.
					FilterLiveSpansExporter{spanExp})
			}
		}

		logExp, err := ConfiguredLogExporter(ctx)
		if err != nil {
			return nil, fmt.Errorf("configuring log exporter from environment variables: %w", err)
		}
		if logExp != nil {
			val, exists := os.LookupEnv("OTEL_EXPORTER_OTLP_LOGS_LIVE")
			if exists && val != "" {
				cfg.LiveLogExporters = append(cfg.LiveLogExporters, logExp)
			} else {
				cfg.BatchedLogExporters = append(cfg.BatchedLogExporters, logExp)
			}
		}

		// if exp, ok := ConfiguredMetricExporter(ctx); ok {
		// 	cfg.LiveMetricExporters = append(cfg.LiveMetricExporters, exp)
		// }
	}

	traceOpts := []sdktrace.TracerProviderOption{
		sdktrace.WithResource(cfg.Resource),
	}

	for _, exporter := range cfg.LiveTraceExporters {
		processor := NewLiveSpanProcessor(exporter)
		cfg.SpanProcessors = append(cfg.SpanProcessors, processor)
	}
	for _, exporter := range cfg.BatchedTraceExporters {
		processor := sdktrace.NewBatchSpanProcessor(exporter)
		cfg.SpanProcessors = append(cfg.SpanProcessors, processor)
	}
	for _, proc := range cfg.SpanProcessors {
		traceOpts = append(traceOpts, sdktrace.WithSpanProcessor(proc))
	}

	cfg.traceProvider = sdktrace.NewTracerProvider(traceOpts...)

	// Register our TracerProvider as the global so any imported instrumentation
	// in the future will default to using it.
	//
	// also necessary so that we can establish a root span, otherwise
	// telemetry doesn't work.
	otel.SetTracerProvider(cfg.traceProvider)

	// Set up a log provider if configured.
	if len(cfg.LiveLogExporters) > 0 || len(cfg.BatchedLogExporters) > 0 {
		logOpts := []sdklog.LoggerProviderOption{
			sdklog.WithResource(cfg.Resource),
		}
		for _, exp := range cfg.LiveLogExporters {
			processor := sdklog.NewBatchProcessor(exp,
				sdklog.WithExportInterval(NearlyImmediate))
			cfg.LogProcessors = append(cfg.LogProcessors, processor)
			logOpts = append(logOpts, sdklog.WithProcessor(processor))
		}
		for _, exp := range cfg.BatchedLogExporters {
			processor := sdklog.NewBatchProcessor(exp)
			cfg.LogProcessors = append(cfg.LogProcessors, processor)
		}
		ctx = WithLoggerProvider(ctx, sdklog.NewLoggerProvider(logOpts...))
	}

	// Set up a metric provider if configured.
	// if len(cfg.LiveMetricExporters) > 0 {
	// 	meterOpts := []sdkmetric.Option{
	// 		sdkmetric.WithResource(cfg.Resource),
	// 	}
	// 	const metricsExportInterval = 1 * time.Second
	// 	const metricsExportTimeout = 1 * time.Second
	// 	for _, exp := range cfg.LiveMetricExporters {
	// 		MetricExporters = append(MetricExporters, exp)
	// 		reader := sdkmetric.NewPeriodicReader(exp,
	// 			sdkmetric.WithInterval(metricsExportInterval),
	// 			sdkmetric.WithTimeout(metricsExportTimeout),
	// 		)
	// 		meterOpts = append(meterOpts, sdkmetric.WithReader(reader))
	// 	}
	// 	ctx = WithMeterProvider(ctx, sdkmetric.NewMeterProvider(meterOpts...))
	// }

	return ctx, nil
}

// Close shuts down the global OpenTelemetry providers, flushing any remaining
// data to the configured exporters.
func Close(ctx context.Context, cfg Config) {
	log := logger.FromContext(ctx)

	flushCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
	defer cancel()
	if tracerProvider := otel.GetTracerProvider(); tracerProvider != nil {
		if err := cfg.traceProvider.Shutdown(flushCtx); err != nil {
			log.ErrorContext(ctx, "failed to shut down tracer provider", "error", err)
		}
	}
	if loggerProvider := LoggerProvider(ctx); loggerProvider != nil {
		if err := loggerProvider.Shutdown(flushCtx); err != nil {
			log.ErrorContext(ctx, "failed to shut down logger provider", "error", err)
		}
	}
}

// ConfiguredSpanExporter examines environment variables to build a sdktrace.SpanExporter.
func ConfiguredSpanExporter(ctx context.Context) (sdktrace.SpanExporter, error) {
	// derived from https://github.com/dagger/dagger-go-sdk/blob/v0.14.0/telemetry/init.go#L35
	ctx = context.WithoutCancel(ctx) // TODO: Why?

	var configuredSpanExporter sdktrace.SpanExporter
	var err error

	// handle protocol first so we can guess the full uri from a top-level OTLP endpoint
	var proto string
	if v := os.Getenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL"); v != "" {
		proto = v
	} else if v := os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL"); v != "" {
		proto = v
	} else {
		// https://github.com/open-telemetry/opentelemetry-specification/blob/v1.33.0/specification/protocol/exporter.md#specify-protocol
		proto = "http/protobuf"
	}

	var endpoint string
	if v := os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"); v != "" {
		endpoint = v
	} else if v := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"); v != "" {
		if proto == "http/protobuf" {
			endpoint, err = url.JoinPath(v, "v1", "traces")
			if err != nil {
				return nil, fmt.Errorf("joining OTEL_EXPORTER_OTLP_ENDPOINT traces path: %w", err)
			}
		} else {
			endpoint = v
		}
	}
	if endpoint == "" {
		return nil, nil
	}

	switch proto {
	case "http/protobuf", "http":
		headers := map[string]string{}
		if hs := os.Getenv("OTEL_EXPORTER_OTLP_HEADERS"); hs != "" {
			for _, header := range strings.Split(hs, ",") {
				name, value, _ := strings.Cut(header, "=")
				headers[name] = value
			}
		}
		configuredSpanExporter, err = otlptracehttp.New(ctx,
			otlptracehttp.WithEndpointURL(endpoint),
			otlptracehttp.WithHeaders(headers))
		if err != nil {
			return nil, fmt.Errorf("creating http/protobuf span exporter: %w", err)
		}
	case "grpc":
		return nil, fmt.Errorf("OTLP grpc protocol not supported")
	default:
		return nil, fmt.Errorf("unknown OTLP protocol: %s", proto)
	}

	return configuredSpanExporter, nil
}

// ConfiguredSpanExporter examines environment variables to build a sdklog.Exporter.
func ConfiguredLogExporter(ctx context.Context) (sdklog.Exporter, error) {
	ctx = context.WithoutCancel(ctx)

	var configuredLogExporter sdklog.Exporter
	var err error

	var proto string
	if v := os.Getenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL"); v != "" {
		proto = v
	} else if v := os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL"); v != "" {
		proto = v
	} else {
		// https://github.com/open-telemetry/opentelemetry-specification/blob/v1.33.0/specification/protocol/exporter.md#specify-protocol
		proto = "http/protobuf"
	}

	var endpoint string
	if v := os.Getenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT"); v != "" {
		endpoint = v
	} else if v := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"); v != "" {
		if proto == "http/protobuf" {
			endpoint, err = url.JoinPath(v, "v1", "logs")
			if err != nil {
				return nil, fmt.Errorf("joining OTEL_EXPORTER_OTLP_ENDPOINT logs path: %w", err)
			}
		} else {
			endpoint = v
		}
	}
	if endpoint == "" {
		return nil, nil
	}

	switch proto {
	case "http/protobuf", "http":
		headers := map[string]string{}
		if hs := os.Getenv("OTEL_EXPORTER_OTLP_HEADERS"); hs != "" {
			for _, header := range strings.Split(hs, ",") {
				name, value, _ := strings.Cut(header, "=")
				headers[name] = value
			}
		}
		configuredLogExporter, err = otlploghttp.New(ctx,
			otlploghttp.WithEndpointURL(endpoint),
			otlploghttp.WithHeaders(headers))
		if err != nil {
			return nil, fmt.Errorf("creating http/protobuf log exporter: %w", err)
		}

	case "grpc":
		return nil, fmt.Errorf("OTLP grpc protocol not supported")
	default:
		return nil, fmt.Errorf("unknown OTLP protocol: %s", proto)
	}

	return configuredLogExporter, nil
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
		resource.WithAttributes(
			semconv.ServiceName("ACT3_ASCE"), // default value is "unknown_service"
		),
	)
	return r
}
