package otel

import (
	"context"
	"fmt"
	"os"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
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
		// resource.WithFromEnv(),      // Discover and provide attributes from OTEL_RESOURCE_ATTRIBUTES and OTEL_SERVICE_NAME environment variables.
		resource.WithTelemetrySDK(), // Discover and provide information about the OpenTelemetry SDK used.
		// resource.WithProcess(),      // Discover and provide process information.
		// resource.WithOS(), // Discover and provide OS information.
		// resource.WithContainer(),    // Discover and provide container information.
		// resource.WithHost(), // Discover and provide host information.
		resource.WithAttributes(
			// if no service name is provided, OpenTelemetry will default to ("unknown_service");
			// if no resouce is provided to the config, the fallback resource created by this package is "ACT3_ASCE".
			semconv.ServiceName("Example_Service"),
		),
	)
	if err != nil {
		panic(fmt.Sprintf("insufficient resource information: error = %v", err))
	}

	fmt.Fprintf(os.Stderr, "%s\n", rsrc)
}

// TestExampleConfig_simple wraps ExampleConfig_simple as test func to simply display the
// output. Without this, we would need a deterministic output of the example func.
func TestExampleConfig_simple(t *testing.T) {
	if !testing.Verbose() {
		return
	}
	ExampleConfig_simple()
}

// ExampleConfig_simple demonstrates configuration setup for exporting telemetry
// spans in batches when they finish.
func ExampleConfig_simple() {
	ctx := context.Background()

	rsrc, err := resource.New(ctx,
		resource.WithTelemetrySDK(),
		resource.WithAttributes(semconv.ServiceName("Example_Service")),
	)
	if err != nil {
		panic(fmt.Sprintf("insufficient resource information: error = %v", err))
	}

	// create an exporter of your choosing
	exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint(), stdouttrace.WithWriter(os.Stderr))
	if err != nil {
		panic(fmt.Sprintf("initializing stdout exporter: error = %v", err))
	}

	cfg := Config{
		// LiveTraceExporters: []sdktrace.SpanExporter{}, // export when spans start and finish
		BatchedTraceExporters: []sdktrace.SpanExporter{exp}, // export wehn spans finish
		Resource:              rsrc,
	}

	ctx, err = Init(ctx, &cfg)
	if err != nil {
		panic(fmt.Sprintf("initializing global OpenTelemetry provider: error = %v", err))
	}
	defer Close(ctx, cfg) // ensure to shutdown, flushing remaining data to exporters

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
