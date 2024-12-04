package otel

import (
	"context"
	"time"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Based off of https://github.com/dagger/dagger-go-sdk/blob/v0.14.0/telemetry/live.go

// NearlyImmediate is 100ms, below which has diminishing returns in terms of
// visual perception vs. performance cost.
const NearlyImmediate = 100 * time.Millisecond

// LiveSpanProcessor is a SpanProcessor whose OnStart calls OnEnd on the
// underlying SpanProcessor in order to send live telemetry.
type LiveSpanProcessor struct {
	sdktrace.SpanProcessor
}

func NewLiveSpanProcessor(exp sdktrace.SpanExporter) *LiveSpanProcessor {
	return &LiveSpanProcessor{
		SpanProcessor: sdktrace.NewBatchSpanProcessor(
			exp,
			sdktrace.WithBatchTimeout(NearlyImmediate),
		),
	}
}

// OnStart calls the underlying batch span processor's OnEnd func, which simply
// enqueues spans to be exported.
func (p *LiveSpanProcessor) OnStart(ctx context.Context, span sdktrace.ReadWriteSpan) {
	// Send a read-only snapshot of the live span downstream so it can be
	// filtered out by FilterLiveSpansExporter. Otherwise the span can complete
	// before being exported, resulting in two completed spans being sent, which
	// will confuse traditional OpenTelemetry services.
	p.SpanProcessor.OnEnd(span) // TODO: dagger does some transformation here instead of passing it directly
}

// FilterLiveSpansExporter is a SpanExporter that filters out spans that are
// currently running, as indicated by an end time older than its start time
// (typically year 1753).
type FilterLiveSpansExporter struct {
	sdktrace.SpanExporter
}

// ExportSpans passes each span to the span processor's OnEnd hook so that it
// can be batched and emitted more efficiently.
func (exp FilterLiveSpansExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	filtered := make([]sdktrace.ReadOnlySpan, 0, len(spans))
	for _, span := range spans {
		if span.StartTime().After(span.EndTime()) {
		} else {
			filtered = append(filtered, span)
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	return exp.SpanExporter.ExportSpans(ctx, filtered)
}
