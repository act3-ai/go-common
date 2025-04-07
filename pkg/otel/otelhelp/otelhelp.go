// Package otelhelp defines CLI help commands with OTel configuration docs.
package otelhelp

import (
	_ "embed"

	"github.com/act3-ai/go-common/pkg/termdoc"
	"github.com/spf13/cobra"
)

// Fetch the otel docs
// //go:generate ./fetch-otel-docs.sh otel-general.md otel-otlp-exporter.md

// Embed the otel docs
var (
	//go:embed otel-general.md
	otelGeneral string

	//go:embed otel-otlp-exporter.md
	otelExporter string
)

// GeneralHelpCmd creates a help command for general OpenTelemetry configuration.
func GeneralHelpCmd() *cobra.Command {
	return termdoc.AdditionalHelpTopic(
		"otel-config",
		"Help for general OpenTelemetry configuration.",
		otelGeneral,
		termdoc.AutoMarkdownFormat(),
	)
}

// ExporterHelpCmd creates a help command for OpenTelemetry Protocol Exporter (OTLP) configuration.
func ExporterHelpCmd() *cobra.Command {
	return termdoc.AdditionalHelpTopic(
		"otlp-config",
		"Help for OpenTelemetry Protocol Exporter (OTLP) configuration.",
		otelExporter,
		termdoc.AutoMarkdownFormat(),
	)
}

// GeneralDoc returns the general OpenTelemetry configuration document for caller use.
//
// Use this to create your own "Additional Help Topic" command if you want different configuration.
func GeneralDoc() string {
	return otelGeneral
}

// ExporterDoc returns the OpenTelemetry Protocol Exporter (OTLP) configuration document for caller use.
//
// Use this to create your own "Additional Help Topic" command if you want different configuration.
func ExporterDoc() string {
	return otelExporter
}
