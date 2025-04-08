package cobrautil

import (
	"bytes"
	"strings"
	"unicode"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/act3-ai/go-common/pkg/options"
	"github.com/act3-ai/go-common/pkg/options/flagutil"
)

// CommandFlagUsages returns flag usage description for all flags of a command.
/*
{{if .HasAvailableLocalFlags}}

{{localFlagUsages .LocalFlags | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

{{inheritedFlagUsages .InheritedFlags | trimTrailingWhitespaces}}{{end}}
*/
func CommandFlagUsages(cmd *cobra.Command, opts UsageFormatOptions) string {
	if opts.LocalFlags.UngroupedHeader == "" {
		opts.LocalFlags.UngroupedHeader = DefaultLocalFlagHeader
	}
	if opts.InheritedFlags.UngroupedHeader == "" {
		opts.InheritedFlags.UngroupedHeader = DefaultGlobalFlagHeader
	}

	buf := new(bytes.Buffer)

	if cmd.HasAvailableLocalFlags() {
		usage := LocalFlagUsages(cmd, opts)
		usage = strings.TrimRightFunc(usage, unicode.IsSpace) // trimTrailingWhitespaces
		buf.WriteString(usage + "\n")
	}

	// Additional separator if needed
	if cmd.HasAvailableLocalFlags() && cmd.HasAvailableInheritedFlags() {
		buf.WriteString("\n")
	}

	if cmd.HasAvailableInheritedFlags() {
		usage := InheritedFlagUsages(cmd, opts)
		usage = strings.TrimRightFunc(usage, unicode.IsSpace) // trimTrailingWhitespaces
		buf.WriteString(usage + "\n")
	}

	return buf.String()
}

// LocalFlagUsages returns flag usage for a command's local flags.
func LocalFlagUsages(cmd *cobra.Command, opts UsageFormatOptions) string {
	if opts.LocalFlags.UngroupedHeader == "" {
		opts.LocalFlags.UngroupedHeader = DefaultLocalFlagHeader
	}

	if !cmd.HasAvailableLocalFlags() {
		return ""
	}

	return GroupedFlagUsages(cmd.LocalFlags(), opts.LocalFlags, opts.Format, opts.FlagOptions)
}

// InheritedFlagUsages returns flag usage for a command's inherited flags.
func InheritedFlagUsages(cmd *cobra.Command, opts UsageFormatOptions) string {
	if opts.InheritedFlags.UngroupedHeader == "" {
		opts.InheritedFlags.UngroupedHeader = DefaultGlobalFlagHeader
	}

	if !cmd.HasAvailableInheritedFlags() {
		return ""
	}

	return GroupedFlagUsages(cmd.InheritedFlags(), opts.InheritedFlags, opts.Format, opts.FlagOptions)
}

// GroupedFlagUsages returns a string containing the usage information
// for all flags in the FlagSet. Wrapped to `cols` columns (0 for no
// wrapping)
func GroupedFlagUsages(f *pflag.FlagSet, gopts FlagGroupingOptions, format Formatter, opts flagutil.UsageFormatOptions) string {
	format.Default() // default formatter funcs

	buf := new(strings.Builder)

	if !gopts.GroupFlags {
		header := gopts.UngroupedHeader
		header = format.Header(header)
		if header != "" {
			_, _ = buf.WriteString(header + "\n")
		}
		_, _ = buf.WriteString(flagutil.FlagUsages(f, opts))
		return buf.String()
	}

	groups, ungrouped := options.ToGroupFlagSets(f)

	// Write ungrouped flags
	if ungrouped.FlagSet.HasAvailableFlags() {
		header := gopts.UngroupedHeader
		header = format.Header(header)
		_, _ = buf.WriteString(header + "\n")
		_, _ = buf.WriteString(flagutil.FlagUsages(ungrouped.FlagSet, opts))

		// Add newline to separate first group from the ungrouped flags,
		// but only if there are groups to be separated
		if len(groups) > 0 {
			_, _ = buf.WriteString("\n")
		}
	}

	// Write each group of flags
	for i, group := range groups {
		if !group.FlagSet.HasAvailableFlags() {
			// Skip empty groups (unsure how this could happen)
			continue
		}
		header := strings.TrimRight(group.Description, ".:") + ":"
		header = format.Header(header)
		if header != "" {
			if i != 0 {
				// Separate from previous group
				_, _ = buf.WriteString("\n")
			}
			_, _ = buf.WriteString(header + "\n")
		}
		_, _ = buf.WriteString(flagutil.FlagUsages(group.FlagSet, opts))
	}

	return buf.String()
}
