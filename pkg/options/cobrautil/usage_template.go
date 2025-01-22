package cobrautil

import (
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"gitlab.com/act3-ai/asce/go-common/pkg/options/flagutil"
)

// Default values for the flag section headers.
const (
	DefaultLocalFlagHeader  = "Options:"
	DefaultGlobalFlagHeader = "Global options:"
)

// Formatter defines general formatting functions.
type Formatter struct {
	Header  func(s string) string // Formats all header lines
	Command func(s string) string // Formats all command names
	Args    func(s string) string // Formats all command arg placeholders
}

// Default initializes the formatter so it is safe to call all of its functions without nil checks.
func (f *Formatter) Default() {
	if f == nil {
		f = &Formatter{}
	}
	if f.Header == nil {
		f.Header = noopFormat
	}
	if f.Command == nil {
		f.Command = noopFormat
	}
	if f.Args == nil {
		f.Args = noopFormat
	}
}

// UsageFormatOptions is used to format flag usage output.
type UsageFormatOptions struct {
	Format         Formatter                   // General formatting functions
	FlagOptions    flagutil.UsageFormatOptions // Flag formatting options
	LocalFlags     FlagGroupingOptions         // Flag grouping options (for local flags)
	InheritedFlags FlagGroupingOptions         // Flag grouping options (for inherited flags)
}

// FlagGroupingOptions is used to group flags.
type FlagGroupingOptions struct {
	GroupFlags      bool   // Set true to organize flags by group.
	SortFlags       bool   // Set true to sort flags alphabetically.
	UngroupedHeader string // Header for section of flags without group.
}

func noopFormat(s string) string { return s }

// WithCustomUsage modifies a command's usage function according to the UsageFormatOptions.
func WithCustomUsage(cmd *cobra.Command, opts UsageFormatOptions) {
	opts.Format.Default() // default formatter funcs
	if opts.LocalFlags.UngroupedHeader == "" {
		opts.LocalFlags.UngroupedHeader = DefaultLocalFlagHeader
	}
	if opts.InheritedFlags.UngroupedHeader == "" {
		opts.InheritedFlags.UngroupedHeader = DefaultGlobalFlagHeader
	}
	cobra.AddTemplateFuncs(template.FuncMap{
		"flagUsages": func(cmd *cobra.Command) string {
			return CommandFlagUsages(cmd, opts)
		},
		// "localFlagUsages": func(f *pflag.FlagSet) string {
		// 	return GroupedFlagUsages(f, opts.LocalFlags, opts.Format, opts.FlagOptions)
		// },
		// "inheritedFlagUsages": func(f *pflag.FlagSet) string {
		// 	return GroupedFlagUsages(f, opts.InheritedFlags, opts.Format, opts.FlagOptions)
		// },
		"formatHeader": func(s string) string {
			return opts.Format.Header(s)
		},
		"formatCommand": func(s string) string {
			return opts.Format.Command(s)
		},
		"formatArgs": func(s string) string {
			return opts.Format.Args(s)
		},
		"formattedUseLine": func(cmd *cobra.Command) string {
			useline := cmd.UseLine()

			commandPath := cmd.CommandPath()
			commandArgs := ""
			if strings.HasPrefix(useline, commandPath) {
				commandArgs = strings.TrimPrefix(useline, commandPath+" ")
			} else {
				// give up
				return useline
			}

			commandPath = opts.Format.Command(commandPath)
			// useline = opts.FormatCommand(cmd.CommandPath()) + strings.TrimPrefix(useline, cmd.CommandPath())
			commandArgs = opts.Format.Args(commandArgs)
			return commandPath + " " + commandArgs
		},
		// Indent s by indent spaces (including the first line)
		"indent": func(indent int, s string) string {
			linePrefix := strings.Repeat(" ", indent)
			lines := strings.Split(s, "\n")
			for i := range lines {
				if lines[i] == "" {
					// Do not pad empty lines
					continue
				}
				lines[i] = linePrefix + lines[i]
			}
			return strings.Join(lines, "\n")
		},
	})
	cmd.SetUsageTemplate(groupedFlagsUsageTemplate)
}

// This is a modified version of cobra's usage template.
var groupedFlagsUsageTemplate = `{{formatHeader "Usage:"}}{{if .Runnable}}
  {{formattedUseLine .}}{{end}}{{if .HasAvailableSubCommands}}
  {{formatCommand .CommandPath}} {{formatArgs "[command]"}}{{end}}{{if gt (len .Aliases) 0}}

{{formatHeader "Aliases:"}}
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

{{formatHeader "Examples:"}}
{{.Example | indent 2}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

{{formatHeader "Available Commands:"}}{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{formatCommand (rpad .Name .NamePadding) }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{formatHeader .Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{formatCommand (rpad .Name .NamePadding) }} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

{{formatHeader "Additional Commands:"}}{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  {{formatCommand (rpad .Name .NamePadding) }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{with flagUsages .}}

{{ . | trimTrailingWhitespaces }}{{end}}{{if .HasHelpSubCommands}}

{{formatHeader "Additional help topics:"}}{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{formatCommand (rpad .CommandPath .CommandPathPadding)}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

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
