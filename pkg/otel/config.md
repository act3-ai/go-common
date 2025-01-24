# Configuring OpenTelemetry Signal Processing

OpenTelemetry offers a wide range of configuration options that may burden the new developer when instrumenting OTel signals into their application. The `otel` package is intended to reduce this initial cognitive burden, while supporting more advanced use cases as needed.

Third-party packages may also aid in adding instrumentation, refer to the [OTel Registry](https://opentelemetry.io/ecosystem/registry/).

## Table of Contents

- [Configuring OpenTelemetry Signal Processing](#configuring-opentelemetry-signal-processing)
  - [Table of Contents](#table-of-contents)
  - [OpenTelemetry Export Errors](#opentelemetry-export-errors)
  - [Quick-Start](#quick-start)
    - [Add Default Configuration](#add-default-configuration)
    - [Instrument OTel Signals](#instrument-otel-signals)
      - [Logs](#logs)
      - [Traces and Metrics](#traces-and-metrics)
      - [HTTP Clients and Servers](#http-clients-and-servers)
        - [Clients](#clients)
        - [Servers](#servers)
  - [Configuration Through Environment Variables](#configuration-through-environment-variables)
  - [Live Exporting](#live-exporting)
  - [Hardcoded](#hardcoded)

## OpenTelemetry Export Errors

- Any errors encountered when initializing components, exporting signals, and shutting down processors are never fatal, and logged at the error level.
  - It is recommended for custom OTel configurations to adhere to this non-fatal error handling practice, e.g. when creating exporters or resources.
- Logged OTel errors are only sent to stderr, not to the OTel logging endpoint(s).
- If an error occurs during initialization or if an OTel endpoint is not configured, all signal intrumentation is a noop.

## Quick-Start

The minimum steps to add OTel instumentation are based on existing `go-common` practices or projects initialized with the [act3-project-tool](https://gitlab.com/act3-ai/asce/pt#act3-project-tool).

### Add Default Configuration

More configuration examples are available in [config_test.go](./config_test.go) or in the [go pkg registry](https://pkg.go.dev/go.opentelemetry.io/otel/sdk@v1.33.0/resource#example-New). However, the following provides insight on where to place the configuration in the context of a project structured with `go-common` practices or created with the [act3-project-tool](https://gitlab.com/act3-ai/asce/pt#act3-project-tool).

- In `main.go` add a custom resource, which provides identifying information added to OTel signals; e.g. which service generated the signals.
  - See [OTel Semantic Conventions](https://opentelemetry.io/docs/specs/semconv/resource/#service) for service naming best practices.
- Any errors encountered during OTel configuration should be logged within `root.PersistentPreRun`, see [OTel Export Errors](#opentelemetry-export-errors) for more info.
- Replace `runner.RunWithContext()` with `otel.RunWithContext()`, and add the `otel.Config` as an argument.

```go
import (
   "context"
   "log/slog"

   "go.opentelemetry.io/otel/sdk/resource"
   semconv "go.opentelemetry.io/otel/semconv/v1.17.0"

   "gitlab.com/act3-ai/asce/go-common/pkg/otel"
   "gitlab.com/act3-ai/asce/go-common/pkg/logger"
)

func main() {
   info := getVersionInfo()
   root := cli.NewCLI(info.Version)
   root.SilenceUsage = true

   ctx := context.Background()

   // Define custom resource. Order matters, by defining a service name before WithFromEnv()
   // we can use "my.service.name" as the default while allowing users to override via
   // OTEL_SERVICE_NAME.
   r, err := resource.New(
      ctx,
      resource.WithAttributes(
         semconv.ServiceName("my.service.name"),
         semconv.ServiceVersion(info.Version),
      ),
      resource.WithFromEnv(),
      resource.WithTelemetrySDK(),
      resource.WithOS(),
   )

   // Optionally, create hardcoded exporters here with errors logged within root.PersistentPreRun

   // Add resource to config
   otelCfg := otel.Config{
      Resource: r,
      // Hardcoded exporters may be added here...
   }

   root.PersistentPreRun = func(cmd *cobra.Command, args []string) {
      ctx := cmd.Context()
      // OTel errors should not be fatal, but we must wait for the logger to be
      // initialized in otel.RunWithContext(); a convention established by pkg runner.RunWithContext()
      log := logger.FromContext(ctx)
      if err != nil {
         log.ErrorContext(ctx, "insufficient resource information", "error", err)
      }
   }

   // ...

   // Run root command with OTel, replaces runner.RunWithContext().
   // Initializes OTel providers and shuts them down appropriately.
   if err := otel.RunWithContext(ctx, root, &otelCfg, "MY_SERVICE_VERBOSITY"); err != nil {
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
   Tracer = otel.GetTracerProvider().Tracer("act3.asce.otel-demo.hello")
   Meter = otel.GetMeterProvider().Meter("act3.asce.otel-demo.hello")

   // the first tracer.Start() creates a root span
   ctx, span := Tracer.Start(ctx, "RootSpanName")
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
   counter, err := Meter.Float64Counter("ExampleCounter", metric.WithDescription("We're counting something."), metric.WithUnit("ExampleUnit"))
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
   ctx, span := Tracer.Start(ctx, "NestedSpan", trace.WithAttributes(attribute.String("foo", "bar")))
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
   ctx, span := Tracer.Start(ctx, "SecondRootSpan", trace.WithNewRoot())
   defer span.End()
}
```

#### HTTP Clients and Servers

Clients and Servers may use the `otelhttp` [package](https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp@v0.58.0#section-documentation), developed by OTel maintainers, or third-party packages available in the [OTel Registry](https://opentelemetry.io/ecosystem/registry/).

##### Clients

Client instrumentation wraps an existing `http.RoundTripper`, starting a span for each outbound request. The span context is propagated through HTTP headers, which may be inherited by the server.

Before:

```go
func example() {
   &http.Client{
      Transport: http.DefaultTransport,
   }
}
```

After:

```go
func example() {
   &http.Client{
      Transport: otelhttp.NewTransport(http.DefaultTransport),
   }
}
```

##### Servers

Server instrumentation wraps existing handlers with a named span. Optionally add [filters](https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp@v0.58.0#WithFilter) to exclude span creation based on information available in an `http.Request`.

Before:

```go
type Serve struct {
   *Tool
   Listen string // port
}

func (action *Serve) Run(ctx context.Context) {
   servMux := http.NewServeMux()
   servMux.HandleFunc("/hello", handleHello)

   // ...

   // Run server
}

func (action *Serve) handleHello(w http.ResponseWriter, r *http.Request) {
   defer r.Body.Close()

   log := logger.FromContext(r.Context())
   log.InfoContext(r.Context(), "received hello request")

   // ...

   _, err := io.WriteString(w, "Hello, from otel-demo-server!\n")
   if err != nil {
      panic(fmt.Sprintf("writing response: error = %v", err))
   }
}
```

After:

```go
type Serve struct {
   *Tool
   Listen string

   requestDurHistogram metric.Float64Histogram // http request duration
}

func (action *Serve) Run(ctx context.Context) error {
   servMux := http.NewServeMux()
   servMux.Handle("/hello",
      otelhttp.NewHandler(http.HandlerFunc(action.getHello), "HelloResponse"),
   )

   // Initialize meters, making them available to handlers
   meter := otel.GetMeterProvider().Meter("otel-server-meter")
   action.requestDurHistogram, err = meter.Float64Histogram(
      "http.request.duration", // meter name
      metric.WithDescription("The duration of an HTTP request."),
      metric.WithUnit("s"),
   )
   if err != nil {
      return fmt.Errorf("initializing request duration histogram: %w", err)
   }

   // Run server
}

func (action *Serve) handleHello(w http.ResponseWriter, r *http.Request) {
   defer r.Body.Close()

   // Use request context to inherit the trace and span, if propagated.
   span := trace.SpanFromContext(r.Context())
   defer span.End()

   // log after span is created, ensuring to correlate logs to the trace
   log := logger.FromContext(r.Context())
   log.InfoContext(r.Context(), "received hello request")

   start := time.Now() // start timing request duration

   // things are happening
   span.AddEvent("Found one o in foo.", trace.WithTimestamp(time.Now()))
   span.AddEvent("Found two o's in foo.", trace.WithTimestamp(time.Now()))

   // record request duration
   action.requestDurHistogram.Record(r.Context(), time.Since(start).Seconds())
   _, err := io.WriteString(w, "Hello, from otel-demo-server!\n")
   if err != nil {
   panic(fmt.Sprintf("writing response: error = %v", err))
   }
}
```

## Configuration Through Environment Variables

Without a hardcoded configuration, the default behavior of an OTel-instrumented application is *opt-in*. As such, users must define a receiver endpoint to send OTel signals to. It is important to note that the value of receiver related environment variables is dependent on the deployment of the receiver(s). Intuitively, the "blanket" approach is the lowest barrier for users to opt-in by requiring a single environment variable to be set.

A "blanket" approach:

- User sets a single endpoint `OTEL_EXPORTER_OTLP_ENDPOINT="http://localhost:4318"`
- Traces are sent to `http://localhost:4318/v1/traces`
- Logs are sent to `http://localhost:4318/v1/logs`
- Metrics are sent to `http://localhost:4318/v1/metrics`

A "fine-grained" approach:

- User sets a custom endpoint for each signal, which are used directly without extra path joining
  - `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT="http://localhost:4318/trace/endpoint"`
  - `OTEL_EXPORTER_OTLP_LOGS_ENDPOINT="http://localhost:4318/logs/endpoint"`
  - `OTEL_EXPORTER_OTLP_METRICS_PROTOCOL="http://localhost:4318/metrics/endpoint"`

Advanced configuration may be done through environment variables standardized across implementations. They include export timings, queue sizes, attribute limits, and more. See [OpenTelemetry docs](https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables/) for more information.

## Live Exporting

Live exporting, inspired by dagger's implementation, is custom to the `go-common` `otel` package and serves as a method to quickly export and process OTel signals. For standard uses cases, configuring live exporting may not be recommended as it is typically sufficient to view the OTel signals after the default export intervals. However, some use cases may be heavily dependent on the processing speed of OTel signals; such as the refresh rate of information displayed in a TUI.

Live exporting may be enabled for traces, logs, and metrics independently; through hardcoded exporters and processors or through designated environment variables. Although this functionality may be achieved through well-defined exporter and processor options, the *Live* abstraction is intended for developers who wish to quickly configure sufficiently fast OTel signal processing with minimal cognitive burden.

While live traces and logs use a 100ms export interval, metrics uses a 1s interval to limit the impact of the exports on the metrics themselves.

- `OTEL_EXPORTER_OTLP_TRACES_LIVE`
  - Export interval: 100ms.
- `OTEL_EXPORTER_OTLP_LOGS_LIVE`
  - Export interval: 100ms.
- `OTEL_EXPORTER_OTLP_METRICS_LIVE`
  - Export interval: 1s.
  
## Hardcoded

TODO
