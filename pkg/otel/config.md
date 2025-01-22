# Configuring OpenTelemetry Signal Processing

OpenTelemetry offers a wide range of configuration options that may burden the new developer when instrumenting OTel signals into their application. The `otel` package is intended to reduce this initial cognitive burden, while supporting more advanced use cases as needed.

## OpenTelemetry Export Errors

* Any errors encountered when initializing components, exporting signals, and shutting down processors are never fatal, and logged at the error level.
  * It is recommended for custom OTel configurations to adhere to this non-fatal error handling practice, e.g. when creating exporters or resources.
* Logged OTel errors are only sent to stderr, not to the OTel logging endpoint(s).
* If an error occurs during initialization or if an OTel endpoint is not configured, all signal intrumentation is a noop.

## Quick-Start

The minimum steps to add OTel instumentation are based on existing `go-common` practices or projects initialized with the [act3-project-tool](https://gitlab.com/act3-ai/asce/pt#act3-project-tool).

### Add Default Configuration

In `main.go` add a custom resource, which serves to provide identifying information added to OTel signals. More examples are available in [config_test.go](./config_test.go) or in the [go pkg registry](https://pkg.go.dev/go.opentelemetry.io/otel/sdk@v1.33.0/resource#example-New).

```go
import (
    "gitlab.com/act3-ai/asce/go-common/pkg/otel"
)

func main() {
   info := getVersionInfo()
   root := cli.NewCLI(info.Version)
   root.SilenceUsage = true

   // Define custom resource. Order matters, by defining service name before WithFromEnv()
   // we can use "my_service_name" as the default while allowing users to override via
   // OTEL_SERVICE_NAME.
   r, err := resource.New(
      root.Context(),
      resource.WithAttributes(
         semconv.ServiceName("my_service_name"),
   ),
   resource.WithFromEnv(),
   resource.WithTelemetrySDK(),
   resource.WithOS(),
)
if err != nil {
   panic(fmt.Sprintf("insufficient resource information: error = %w", err))
}

// Add resource to config
otelCfg := otel.Config{
   Resource: r,
   // Hardcoded exporters may be added here...
}

// ...

// Run root command with OTel, replaces runner.RunWithContext().
// Initializes OTel providers and shuts them down appropriately.
if err := otel.RunWithContext(context.Background(), root, &otelCfg, "MY_SERVICE_VERBOSITY"); err != nil {
   os.Exit(1)
}
}
```

### Instrument OTel Signals

Documentation on OTel naming conventions are available on [github](https://github.com/open-telemetry/opentelemetry-specification/blob/v1.39.0/specification/glossary.md#instrumentation-library). In summary, it is advised to use names that identify the instrumentation scope or library.

#### Logs

`otel.RunWithContext()` configures logging for you. Logs at the configured verbosity will be output to stderr and all logs will be exported as an OTel signal, if an endpoint is defined (e.g. setting `OTEL_EXPORTER_OTLP_ENDPOINT`). Logs are written to stderr, at the configured verbosity level, regardless of OTel logging configuration.

#### Traces and Metrics

Any instrumentation for traces or metrics need access to a `trace.Tracer` or `metric.Meter` respectively. They are created from the global tracer and meter providers, which are initialized in `otel.RunWithContext()`. For this reason, setting a global `Tracer` and `Meter` is a common practice. Following `go-common` practices, it is advised to initialize a trace or metric itself in a `Run` action.

```go
import (
   "go.opentelemetry.io/otel"
   "go.opentelemetry.io/otel/attribute"
   "go.opentelemetry.io/otel/metric"
   "go.opentelemetry.io/otel/trace"
)

// Tracer and Meter must be available to scopes being instrumented.
var Tracer trace.Tracer
var Meter metric.Meter

func (action *Hello) Run(ctx context.Context) error {
   // logs before starting a root span will not include a traceid or spanid.
   // it will not be directly correlated to a trace or span when exported, but
   // be present in the log itself.
   log := logger.FromContext(ctx)
   log.InfoContext(ctx, "This log will write to stderr and export to OTel endpoint but without trace-identifying metadata.")

   // alternatively, the tracer and meter may be set elsewhere; such as in cobra command's cmd.PersistentPreRun
   tracer = otel.GetTracerProvider().Tracer("act3.asce.otel-demo.hello")
   meter = otel.GetMeterProvider().Meter("act3.asce.otel-demo.hello")

   // the first tracer.Start() creates a root span
   ctx, span := tracer.Start(ctx, "RootSpanName")
   defer span.End() // always end the span to release resources

   // correlated log
   log.InfoContext(ctx, "This log will include a traceid and spanid for exported logs only, logs to stderr will not include these values.")

   // events indicate something has occurred within a span that is not have
   span.AddEvent("Something has happened.")

   // create child span, the typical use case
   nestedSpans(ctx)

   // inherit the span created above
   nonNestedSpans(ctx)

   // create a new root span
   newRootSpan(ctx)

   // different types of metric data may be collected, use "Observable" versions for async instrumentation.
   // see https://pkg.go.dev/go.opentelemetry.io/otel/metric@v1.33.0#Meter
   counter, err := meter.Float64Counter("ExampleCounter", metric.WithDescription("We're counting something."), metric.WithUnit("ExampleUnit"))
   if err != nil {
      log.ErrorContext(ctx, "initializing counter metric", "error", err)
   }

   for i := range 5 {
      counter.Add(ctx, 1, metric.WithAttributes(attribute.String("foo", "bar")))
   }

   return nil
}

func nestedSpans(ctx context.Context) {
   // subsequent calls to tracer.Start() create a child span.
   ctx, span := tracer.Start(ctx, "NestedSpan", trace.WithAttributes(attribute.String("foo", "bar")))
   defer span.End() // always end the span to release resources

   log := logger.FromContext(ctx)
   log.InfoContext(ctx, "Again, this log will write to stderr and export to OTel endpoint.")
}

func nonNestedSpans(ctx context.Context) {
   // inherit an existing span from the context, does not create a child span
   span := trace.SpanFromContext(ctx)
   // for clarity, don't end the span here in favor of keeping it paired with the span creator.
}

func newRootSpan(ctx context.Context) {
   // root spans do not have a parent span
   ctx, span := tracer.Start(ctx, "SecondRootSpan", trace.WithNewRoot())
   defer span.End()
}
```

## Configuration Through Environment Variables

Without a hardcoded configuration, the default behavior of an OTel-instrumented application is *opt-in*. As such, users must define a receiver endpoint to send OTel signals to. It is important to note that the value of receiver related environment variables is dependent on the deployment of the receiver(s). Intuitively, the "blanket" approach is the lowest barrier for users to opt-in by requiring a single environment variable to be set.

A "blanket" approach:

* User sets a single endpoint `OTEL_EXPORTER_OTLP_ENDPOINT="http://localhost:4318"`
* Traces are sent to `http://localhost:4318/v1/traces`
* Logs are sent to `http://localhost:4318/v1/logs`
* Metrics are sent to `http://localhost:4318/v1/metrics`

A "fine-grained" approach:

* User sets a custom endpoint for each signal, which are used directly without extra path joining
  * `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT="http://localhost:4318/trace/endpoint"`
  * `OTEL_EXPORTER_OTLP_LOGS_ENDPOINT="http://localhost:4318/logs/endpoint"`
  * `OTEL_EXPORTER_OTLP_METRICS_PROTOCOL="http://localhost:4318/metrics/endpoint"`

Advanced configuration may be done through environment variables standardized across implementations. They include export timings, queue sizes, attribute limits, and more. See [OpenTelemetry docs](https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables/) for more information.

## Live Exporting

Live exporting, inspired by dagger's implementation, is custom to the `go-common` `otel` package and serves as a method to quickly export and process OTel signals. For standard uses cases, configuring live exporting may not be recommended as it is typically sufficient to view the OTel signals after the default export intervals. However, some use cases may be heavily dependent on the processing speed of OTel signals; such as the refresh rate of information displayed in a TUI.

Live exporting may be enabled for traces, logs, and metrics independently; through hardcoded exporters and processors or through designated environment variables. Although this functionality may be achieved through well-defined exporter and processor options, the *Live* abstraction is intended for developers who wish to quickly configure sufficiently fast OTel signal processing with minimal cognitive burden.

While live traces and logs use a 100ms export interval, metrics uses a 1s interval to limit the impact of the exports on the metrics themselves.

* `OTEL_EXPORTER_OTLP_TRACES_LIVE`
  * Export interval: 100ms.
* `OTEL_EXPORTER_OTLP_LOGS_LIVE`
  * Export interval: 100ms.
* `OTEL_EXPORTER_OTLP_METRICS_LIVE`
  * Export interval: 1s.
  
## Hardcoded

TODO
