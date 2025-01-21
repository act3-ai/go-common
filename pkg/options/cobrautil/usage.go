package cobrautil

import (
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gitlab.com/act3-ai/asce/go-common/pkg/options"
)

// WithGroupedFlagUsage modifies a command's usage function to show flags as grouped by the options package.
func WithGroupedFlagUsage(cmd *cobra.Command) {
	cobra.AddTemplateFuncs(template.FuncMap{
		"listFlagGroups": func(flagSet *pflag.FlagSet) []*options.Group {
			groups, _ := options.ToGroups(flagSet)
			return groups
		},
		"getGroupFlagSet":   options.GetGroupFlagSet,
		"getNoGroupFlagSet": options.GetNoGroupFlagSet,
		"groupHeader": func(group *options.Group) string {
			return strings.TrimRight(group.Description, ".:")
		},
	})
	cmd.SetUsageTemplate(groupedFlagsUsageTemplate)
}

// This is a modified version of cobra's usage template.
var groupedFlagsUsageTemplate = `Usage:{{if .Runnable}}
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

{{- with getNoGroupFlagSet .LocalFlags }}

Flags:
{{.FlagUsages | trimTrailingWhitespaces}}
{{- end }}
{{- $flags := .LocalFlags }}
{{- $groups := listFlagGroups $flags }}
{{- range $group := $groups }}
{{- $groupFlags := getGroupFlagSet $flags $group }}

{{ groupHeader $group }}:
{{ $groupFlags.FlagUsages | trimTrailingWhitespaces}}
{{- end }}
{{- with getNoGroupFlagSet .InheritedFlags }}

Global Flags:
{{.FlagUsages | trimTrailingWhitespaces}}
{{- end }}
{{- $flags := .InheritedFlags }}
{{- $groups := listFlagGroups $flags }}
{{- range $group := $groups }}
{{- $groupFlags := getGroupFlagSet $flags $group }}

{{ groupHeader $group }}:
{{ $groupFlags.FlagUsages | trimTrailingWhitespaces}}
{{- end }}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
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
