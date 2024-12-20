package otel

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

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

	fmt.Printf("%s\n", rsrc)
	// Output:
	// service.name=Example_Service,telemetry.sdk.language=go,telemetry.sdk.name=opentelemetry,telemetry.sdk.version=1.32.0
}
