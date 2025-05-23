package otel

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"

	slogmulti "github.com/samber/slog-multi"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/act3-ai/go-common/pkg/logger"
	tlog "github.com/act3-ai/go-common/pkg/test"
)

// TestExampleResource wraps ExampleResource as test func to simply display the
// output. Without this, we would need a deterministic output of the example func.
func TestExampleResource(t *testing.T) {
	if !testing.Verbose() {
		return
	}
	ExampleResource()
}

// ExampleResource demonstrates the minimum resouce configuration, while showing
// optional resource information. Although a default may be used, it is recommended
// to define a custom resource configuration to uniquely identify the service
// being integrated with OpenTelemetry.
func ExampleResource() {
	ctx := context.Background()

	rsrc, err := resource.New(ctx,
		resource.WithAttributes(
			// if no service name is provided, OpenTelemetry will default to ("unknown_service"), defining a custom service before WithFromEnv() allows users to overwrite it.
			semconv.ServiceName("example.service"),
			semconv.ServiceVersion(fmt.Sprintf("%d.%d.%d", 0, 0, 1)),
		),
		resource.WithFromEnv(),      // Discover and provide attributes from OTEL_RESOURCE_ATTRIBUTES and OTEL_SERVICE_NAME environment variables.
		resource.WithTelemetrySDK(), // Discover and provide information about the OpenTelemetry SDK used.
		// resource.WithProcess(),      // Discover and provide process information.
		// resource.WithOS(), // Discover and provide OS information.
		// resource.WithContainer(),    // Discover and provide container information.
		// resource.WithHost(), // Discover and provide host information.

	)
	if err != nil {
		panic(fmt.Sprintf("insufficient resource information: error = %v", err))
	}

	fmt.Fprintf(os.Stderr, "%s\n", rsrc)
}

// TestExampleConfig_spans wraps ExampleConfig_spans as a test function since
// our example isn't runnable without a deterministic output.
func TestExampleConfig_spans(t *testing.T) {
	if !testing.Verbose() {
		return
	}
	ExampleConfig_spans()
}

// ExampleConfig_simple demonstrates configuration setup for exporting telemetry
// spans in batches when they finish.
func ExampleConfig_spans() {
	ctx := context.Background()

	rsrc, err := resource.New(ctx,
		resource.WithTelemetrySDK(),
		resource.WithAttributes(
			semconv.ServiceName("example.service"),
			semconv.ServiceVersion(fmt.Sprintf("%d.%d.%d", 0, 0, 1)),
		),
	)
	if err != nil {
		panic(fmt.Sprintf("insufficient resource information: error = %v", err))
	}

	// add options here, else they will be discovered through env vars
	exp, err := otlptracehttp.New(ctx)
	if err != nil {
		panic(fmt.Sprintf("initializing trace exporter: error = %v", err))
	}
	sp := sdktrace.NewBatchSpanProcessor(exp)

	cfg := Config{
		SpanProcessors: []sdktrace.SpanProcessor{sp},
		Resource:       rsrc,
	}

	ctx, err = cfg.Init(ctx)
	if err != nil {
		panic(fmt.Sprintf("initializing OpenTelemetry: error = %v", err))
	}
	defer cfg.Shutdown(ctx) // ensure to shutdown, flushing remaining data to exporters

	// start a tracer
	t := otel.GetTracerProvider().Tracer("ExampleTracer")

	fn := func(ctx context.Context) {
		//	start spans at the beginning of a function...
		ctx, span := t.Start(ctx, "ExampleSpan", trace.WithAttributes(attribute.String("Key", "Value")))
		defer span.End()

		// doing calulations...

		// something happened
		span.AddEvent("ExampleEvent")

		// ...
	}

	fn(ctx)
}

// TestExampleConfig_logs wraps ExampleConfig_logs as a test function since
// our example isn't runnable without a deterministic output.
func TestExampleConfig_logs(t *testing.T) {
	if !testing.Verbose() {
		return
	}
	ExampleConfig_logs()
}

// ExampleConfig_logs demonstrates configuration setup for exporting logs,
// bridged with slog, in batches.
func ExampleConfig_logs() {
	ctx := context.Background()

	rsrc, err := resource.New(ctx,
		resource.WithTelemetrySDK(),
		resource.WithAttributes(
			semconv.ServiceName("example.service"),
			semconv.ServiceVersion(fmt.Sprintf("%d.%d.%d", 0, 0, 1)),
		),
	)
	if err != nil {
		panic(fmt.Sprintf("insufficient resource information: error = %v", err))
	}

	exp, err := otlploghttp.New(ctx)
	if err != nil {
		panic(fmt.Sprintf("initializing log exporter: error = %v", err))
	}
	lp := sdklog.NewBatchProcessor(exp)

	cfg := Config{
		LogProcessors: []sdklog.Processor{lp},
		Resource:      rsrc,
	}

	ctx, err = cfg.Init(ctx)
	if err != nil {
		panic(fmt.Sprintf("initializing OpenTelemetry: error = %v", err))
	}
	defer cfg.Shutdown(ctx) // ensure to shutdown, flushing remaining data to exporters

	// Multi-logger setup is handled by otel.Run().
	level := new(slog.LevelVar)
	level.Set(slog.LevelDebug)
	stdErrHandler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	otelHandler := otelslog.NewHandler("Example", otelslog.WithLoggerProvider(cfg.logProvider))
	multiHandler := slogmulti.Fanout(stdErrHandler, otelHandler)
	log := slog.New(multiHandler)

	fn := func(ctx context.Context) {
		// conventional logging sent to stderr as well as otel exporters
		log.InfoContext(ctx, "Starting function")
		log.ErrorContext(ctx, "Something bad has happened")
		log.DebugContext(ctx, "Debug me please", "foo", "bar")
	}
	fn(ctx)
}

func TestSpans(t *testing.T) {
	ctx := context.Background()
	log := tlog.Logger(t, 0)
	ctx = logger.NewContext(ctx, log)

	rsrc, err := resource.New(ctx,
		resource.WithTelemetrySDK(),
		resource.WithAttributes(
			semconv.ServiceName("example.service"),
			semconv.ServiceVersion(fmt.Sprintf("%d.%d.%d", 0, 0, 1)),
		),
	)
	if err != nil {
		t.Fatalf("insufficient resource information: error = %v", err)
	}

	exp := tracetest.NewInMemoryExporter()
	sp := sdktrace.NewSimpleSpanProcessor(exp) // simple only recommended for testing

	cfg := Config{
		DisableEnvConfiguration: true,
		SpanProcessors:          []sdktrace.SpanProcessor{sp},
		Resource:                rsrc,
	}

	ctx, err = cfg.Init(ctx)
	if err != nil {
		panic(fmt.Sprintf("initializing OpenTelemetry: error = %v", err))
	}
	defer cfg.Shutdown(ctx) // ensure to shutdown, flushing remaining data to exporters

	// start a tracer
	tp := otel.GetTracerProvider().Tracer("ExampleTracer")

	fn := func(ctx context.Context) {
		//	start spans at the beginning of a function...
		_, span := tp.Start(ctx, "ExampleSpan", trace.WithAttributes(attribute.String("Key", "Value")))
		defer span.End()

		// doing calulations...

		// something happened
		span.AddEvent("ExampleEvent")

		// ...
	}
	fn(ctx)

	// typically a force flush is not necessary as shutdown, or if earlier contitions
	// such as timeout or queue limts are met, kicks off the telemetry pipeline.
	// However, doing so would not allow us to ensure the exporter recieved
	// what we expected it to.
	if err := cfg.traceProvider.ForceFlush(ctx); err != nil {
		t.Fatalf("force flushing traces: error = %v", err)
	}

	spans := exp.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("invalid span count: want %d, got %d", 1, len(spans))
	}
}

func TestEmpty(t *testing.T) {
	ctx := context.Background()
	cfg := Config{}
	defer cfg.Shutdown(ctx)

	var err error
	ctx, err = cfg.Init(ctx)
	if err != nil {
		t.Fatalf("initializing with empty configuration: error = %v", err)
	}

	if ctx == nil {
		t.Fatalf("got nil context from empty configuration: error = %v", err)
	}
}
