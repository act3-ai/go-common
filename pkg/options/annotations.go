package options

import (
	"github.com/spf13/pflag"

	"gitlab.com/act3-ai/asce/go-common/pkg/options/flagutil"
)

// FromFlag produces an Option from annotations on a flag.
func FromFlag(f *pflag.Flag) *Option {
	opt := &Option{
		Type:            Type(flagutil.GetFirstAnnotationOr(f, typeAnno, "")),
		TargetGroupName: flagutil.GetFirstAnnotationOr(f, targetGroupAnno, ""),
		Default:         flagutil.GetFirstAnnotationOr(f, defaultAnno, ""),
		Path:            flagutil.GetFirstAnnotationOr(f, jsonAnno, ""),
		Env:             flagutil.GetEnvName(f),
		Flag:            f.Name,
		FlagShorthand:   f.Shorthand,
		FlagUsage:       flagutil.GetFirstAnnotationOr(f, flagUsageAnno, ""),
		FlagType:        flagutil.GetFirstAnnotationOr(f, flagTypeAnno, f.Value.Type()),
		Short:           flagutil.GetFirstAnnotationOr(f, shortAnno, ""),
		Long:            flagutil.GetFirstAnnotationOr(f, longAnno, ""),
	}
	return opt
}

// Defined annotations used to store [Option] fields in [pflag.Flag] annotations.
// Used to round-trip an Option through a [pflag.Flag].
const (
	defaultAnno     = "options_option_default"   // annotation for options.Option.Default
	typeAnno        = "options_option_type"      // annotation for options.Option.Type
	jsonAnno        = "options_option_json"      // annotation for options.Option.Path
	flagUsageAnno   = "options_option_flagUsage" // annotation for options.Option.FlagUsage
	flagTypeAnno    = "options_option_flagType"  // annotation for options.Option.FlagType
	shortAnno       = "options_option_short"     // annotation for options.Option.Short
	longAnno        = "options_option_long"      // annotation for options.Option.Long
	targetGroupAnno = "options_option_target"    // annotation for options.Option.TargetGroupName
	groupAnno       = "options_option_group"     // used to group flags
)

// withOptionConfig adds sets annotations on the flag from the option definition.
func withOptionConfig(f *pflag.Flag, opt *Option) {
	// Default some fields from the flag
	if opt.FlagUsage == "" {
		opt.FlagUsage = f.Usage
	}
	if opt.FlagType == "" {
		// Use UnqouteUsage function to derive type name from usage string
		// or from the flag value if not found in usage.
		if varname, _ := pflag.UnquoteUsage(f); varname != "" {
			opt.FlagType = varname
		}
	}
	if opt.Type != "" {
		flagutil.SetAnnotation(f, typeAnno, string(opt.Type))
	}
	if opt.TargetGroupName != "" {
		flagutil.SetAnnotation(f, targetGroupAnno, opt.TargetGroupName)
	}
	if opt.Default != "" {
		flagutil.SetAnnotation(f, defaultAnno, opt.Default)
	}
	if opt.Path != "" {
		flagutil.SetAnnotation(f, jsonAnno, opt.Path)
	}
	if opt.Env != "" {
		flagutil.SetEnvName(f, opt.Env)
	}
	if opt.FlagUsage != "" {
		flagutil.SetAnnotation(f, flagUsageAnno, opt.FlagUsage)
	}
	if opt.FlagType != "" {
		flagutil.SetAnnotation(f, flagTypeAnno, opt.FlagType)
	}
	if opt.Short != "" {
		flagutil.SetAnnotation(f, shortAnno, opt.Short)
	}
	if opt.Long != "" {
		flagutil.SetAnnotation(f, longAnno, opt.Long)
	}
}
