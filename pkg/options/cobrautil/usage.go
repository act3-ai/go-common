package cobrautil

import (
	"strings"

	"github.com/spf13/pflag"
	"gitlab.com/act3-ai/asce/go-common/pkg/options"
	"gitlab.com/act3-ai/asce/go-common/pkg/options/flagutil"
)

// FlagUsages returns a string containing the usage information
// for all flags in the FlagSet. Wrapped to `cols` columns (0 for no
// wrapping)
func FlagUsages(f *pflag.FlagSet, gopts FlagGroupingOptions, opts UsageFormatOptions) string {
	buf := new(strings.Builder)

	if !gopts.GroupFlags {
		header := gopts.UngroupedHeader
		if opts.FormatHeader != nil {
			header = opts.FormatHeader(header)
		}
		_, _ = buf.WriteString("\n" + header + "\n")
		_, _ = buf.WriteString(flagutil.FlagUsages(f, opts.FlagOptions))
		return buf.String()
	}

	// Write ungrouped flags
	if ungroupedFlags := options.GetNoGroupFlagSet(f); ungroupedFlags != nil {
		header := gopts.UngroupedHeader
		if opts.FormatHeader != nil {
			header = opts.FormatHeader(header)
		}
		_, _ = buf.WriteString("\n" + header + "\n")
		_, _ = buf.WriteString(flagutil.FlagUsages(ungroupedFlags, opts.FlagOptions))
	}

	// Write each group of flags
	groups, _ := options.ToGroups(f)
	for _, group := range groups {
		groupFlags := options.GetGroupFlagSet(f, group)
		if groupFlags == nil {
			// Skip empty groups (unsure how this could happen)
			continue
		}

		header := strings.TrimRight(group.Description, ".:") + ":"
		if opts.FormatHeader != nil {
			header = opts.FormatHeader(header)
		}
		_, _ = buf.WriteString("\n" + header + "\n")
		_, _ = buf.WriteString(flagutil.FlagUsages(groupFlags, opts.FlagOptions))
	}

	return buf.String()
}
