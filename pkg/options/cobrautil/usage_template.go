package cobrautil

import (
	"text/template"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gitlab.com/act3-ai/asce/go-common/pkg/options/flagutil"
)

// Default values for the flag section headers.
const (
	defaultLocalFlagHeader  = "Options:"
	defaultGlobalFlagHeader = "Global options:"
)

// UsageFormatOptions is used to format flag usage output.
type UsageFormatOptions struct {
	FormatHeader   func(s string) string       // Formats all header lines
	FlagOptions    flagutil.UsageFormatOptions // Flag formatting options
	LocalFlags     FlagGroupingOptions         // Flag grouping options (for local flags)
	InheritedFlags FlagGroupingOptions         // Flag grouping options (for inherited flags)
}

// FlagGroupingOptions is used to group flags.
type FlagGroupingOptions struct {
	GroupFlags      bool   // Set true to organize flags by group.
	UngroupedHeader string // Header for section of flags without group.
}

// WithCustomUsage modifies a command's usage function according to the UsageFormatOptions.
func WithCustomUsage(cmd *cobra.Command, opts UsageFormatOptions) {
	if opts.LocalFlags.UngroupedHeader == "" {
		opts.LocalFlags.UngroupedHeader = defaultLocalFlagHeader
	}
	if opts.InheritedFlags.UngroupedHeader == "" {
		opts.InheritedFlags.UngroupedHeader = defaultGlobalFlagHeader
	}
	cobra.AddTemplateFuncs(template.FuncMap{
		"localFlagUsages": func(f *pflag.FlagSet) string {
			return FlagUsages(f, opts.LocalFlags, opts)
		},
		"inheritedFlagUsages": func(f *pflag.FlagSet) string {
			return FlagUsages(f, opts.InheritedFlags, opts)
		},
		"formatHeader": func(s string) string {
			if opts.FormatHeader != nil {
				return opts.FormatHeader(s)
			}
			return s
		},
	})
	cmd.SetUsageTemplate(groupedFlagsUsageTemplate)
}

// This is a modified version of cobra's usage template.
var groupedFlagsUsageTemplate = `{{formatHeader "Aliases:"}}{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

{{formatHeader "Aliases:"}}
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

{{formatHeader "Examples:"}}
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

{{formatHeader "Available Commands:"}}{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

{{formatHeader "Additional Commands:"}}{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}
{{localFlagUsages .LocalFlags | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}
{{inheritedFlagUsages .InheritedFlags | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

{{formatHeader "Additional help topics:"}}{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

// This is the default usage template from cobra.
//
//nolint:unused
var cobraUsageTemplate = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

Available Commands:{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

Additional Commands:{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
