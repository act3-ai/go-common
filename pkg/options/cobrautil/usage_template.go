package cobrautil

import (
	"strings"
	"text/template"

	"github.com/act3-ai/go-common/pkg/options/flagutil"
	"github.com/charmbracelet/x/ansi"
	"github.com/spf13/cobra"
)

// Default values for the flag section headers.
const (
	DefaultLocalFlagHeader  = "Options:"
	DefaultGlobalFlagHeader = "Global options:"
)

// Formatter defines general formatting functions.
type Formatter struct {
	Header         func(s string) string // Formats all header lines
	Command        func(s string) string // Formats all command names
	Args           func(s string) string // Formats all command arg placeholders
	CommandAndArgs func(s string) string // Formats all command+arg snippets (supersedes Command and Args)
	Example        func(s string) string // Formats command examples
}

// SectionFormatter formats entire sections.
type SectionFormatter struct {
	CommandAndArgs func() // Formats all
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
	if f.CommandAndArgs == nil {
		f.CommandAndArgs = noopFormat
	}
	if f.Example == nil {
		f.Example = noopFormat
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
		"formatHeader": func(s string) string {
			return opts.Format.Header(s)
		},
		"formatCommand": func(commandPath string, args ...string) string {
			return formatCommand(opts, commandPath, args...)
		},
		"formatExample": func(s string) string {
			return opts.Format.Example(s)
		},
		"rpadANSI": rpadANSI,
		"formattedUseLine": func(cmd *cobra.Command) string {
			useline := cmd.UseLine()
			commandPath := cmd.CommandPath()
			// commandArgs := []string{}
			if strings.HasPrefix(useline, commandPath) {
				// Get string after the command path
				remainder := strings.TrimPrefix(useline, commandPath+" ")
				// Split on spaces
				commandArgs := strings.Split(remainder, " ")
				return formatCommand(opts, commandPath, commandArgs...)
			}

			// Preserve use line otherwise.
			// commandPath = useline
			return formatCommand(opts, useline)
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
  {{formatCommand .CommandPath "[command]"}}{{end}}{{if gt (len .Aliases) 0}}

{{formatHeader "Aliases:"}}
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

{{formatHeader "Examples:"}}
{{formatExample .Example | indent 2}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

{{formatHeader "Available Commands:"}}{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpadANSI (formatCommand .Name) .NamePadding}} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{formatHeader .Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpadANSI (formatCommand .Name) .NamePadding}} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

{{formatHeader "Additional Commands:"}}{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpadANSI (formatCommand .Name) .NamePadding}} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{with flagUsages .}}

{{ . | trimTrailingWhitespaces }}{{end}}{{if .HasHelpSubCommands}}

{{formatHeader "Additional help topics:"}}{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpadANSI (formatCommand .CommandPath) .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{formatCommand .CommandPath "[command]" "--help"}}" for more information about a command.{{end}}
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

func formatCommand(opts UsageFormatOptions, commandPath string, args ...string) string {
	commandPath = opts.Format.Command(commandPath)
	for i, arg := range args {
		switch {
		case strings.HasPrefix(arg, "--"):
			arg = opts.Format.Command(arg)
		default:
			arg = opts.Format.Args(arg)
		}
		args[i] = arg
	}
	// Assemble full snippet line to format with CommandAndArgs
	snippet := commandPath
	if len(args) > 0 {
		snippet += " " + strings.Join(args, " ")
	}
	return opts.Format.CommandAndArgs(snippet)
}

// rpadANSI adds padding to the right of a string.
//
// based on cobra's version, modified to be ANSI-aware.
func rpadANSI(s string, padding int) string {
	// Cobra implementation:
	// // rpad adds padding to the right of a string.
	// func rpad(s string, padding int) string {
	// 	formattedString := fmt.Sprintf("%%-%ds", padding)
	// 	return fmt.Sprintf(formattedString, s)
	// }
	strlen := ansi.StringWidth(s)
	if strlen < padding {
		return s + strings.Repeat(" ", padding-strlen)
	}
	return s
}
