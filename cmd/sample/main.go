// Package main is a sample CLI tool to demonstrate how these libraries are properly composed.
package main

import (
	"context"
	"embed"
	"os"

	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"

	commands "gitlab.com/act3-ai/asce/go-common/pkg/cmd"
	"gitlab.com/act3-ai/asce/go-common/pkg/config"
	"gitlab.com/act3-ai/asce/go-common/pkg/embedutil"
	"gitlab.com/act3-ai/asce/go-common/pkg/otel"
	"gitlab.com/act3-ai/asce/go-common/pkg/runner"
	vv "gitlab.com/act3-ai/asce/go-common/pkg/version"
)

// manpages and schema definitions are embedded here for use in the gendocs and genschema commands
//
//go:embed schemas/*
var schemas embed.FS

// an example quick start guide is embedded here
// for use in the "gendocs" and "info" commands
//
//go:embed docs/*.md
var docs embed.FS

// getVersionInfo retreives the proper version information for this executable
func getVersionInfo() vv.Info {
	info := vv.Get()
	if version != "" {
		info.Version = version
	}
	return info
}

const verbosityEnvName = "ACE_SAMPLE_VERBOSITY"

func mainE(args []string) error {
	ctx := context.Background()

	info := getVersionInfo()        // Load the version info from the build
	root := newSample(info.Version) // Create the root command

	schemaAssociations := []commands.SchemaAssociation{
		{
			Definition: "configuration-schema.json",
			FileMatch:  config.DefaultConfigValidatePath("ace", "sample", "config.yaml"),
		},
	}

	docs := &embedutil.Documentation{
		Title:   "Sample command showing the use of go-common's utilities for CLI development",
		Command: root,
		Categories: []*embedutil.Category{
			embedutil.NewCategory(
				"docs", "General Documentation", root.Name(), 7,
				embedutil.LoadMarkdown("quick-start-guide", "Example Quick Start Guide", "docs/quick-start-guide.md", docs),
			),
		},
	}

	commands.AddGroupedCommands(root,
		&cobra.Group{
			ID:    "utils",
			Title: "Utility commands",
		},
		commands.NewVersionCmd(info),
		commands.NewInfoCmd(docs),
		commands.NewGendocsCmd(docs),
		commands.NewGenschemaCmd(schemas, schemaAssociations),
	)

	root.SetArgs(args)

	if v := os.Getenv("OTEL_INSTRUMENTATION_ENABLED"); v == "true" {
		r, _ := resource.New(
			ctx,
			resource.WithAttributes(
				semconv.ServiceName("sample"),
				semconv.ServiceVersion(info.Version),
			),
			resource.WithFromEnv(),
			resource.WithTelemetrySDK(),
			resource.WithOS(),
		)

		otelCfg := &otel.Config{
			Resource: r,
		}

		// Run root command with OTel instrumentation enabled.
		return otel.Run(ctx, root, otelCfg, verbosityEnvName)
	}
	return runner.Run(ctx, root, verbosityEnvName)
}

func main() {
	if err := mainE(os.Args[1:]); err != nil {
		os.Exit(1)
	}
}
