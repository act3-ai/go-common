package options

import (
	"github.com/spf13/pflag"

	"github.com/act3-ai/go-common/pkg/options/flagutil"
)

// FromFlag produces an Option from annotations on a flag.
func FromFlag(f *pflag.Flag) *Option {
	opt := &Option{
		Type:            Type(flagutil.GetFirstAnnotationOr(f, typeAnno, "")),
		ValueType:       Type(flagutil.GetFirstAnnotationOr(f, valueTypeAnno, "")),
		TargetGroupName: flagutil.GetFirstAnnotationOr(f, targetGroupAnno, ""),
		Default:         flagutil.GetFirstAnnotationOr(f, defaultAnno, ""),
		Name:            flagutil.GetFirstAnnotationOr(f, nameAnno, ""),
		JSON:            flagutil.GetFirstAnnotationOr(f, jsonAnno, ""),
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
	defaultAnno     = "options_option_default"   // annotation for [Option.Default]
	typeAnno        = "options_option_type"      // annotation for [Option.Type]
	valueTypeAnno   = "options_option_valueType" // annotation for [Option.ValueType]
	nameAnno        = "options_option_name"      // annotation for [Option.Name]
	jsonAnno        = "options_option_json"      // annotation for [Option.JSON]
	flagUsageAnno   = "options_option_flagUsage" // annotation for [Option.FlagUsage]
	flagTypeAnno    = "options_option_flagType"  // annotation for [Option.FlagType]
	shortAnno       = "options_option_short"     // annotation for [Option.Short]
	longAnno        = "options_option_long"      // annotation for [Option.Long]
	targetGroupAnno = "options_option_target"    // annotation for [Option.TargetGroupName]
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
	setAnnoIfNotEmpty(f, typeAnno, opt.Type)
	setAnnoIfNotEmpty(f, valueTypeAnno, opt.ValueType)
	setAnnoIfNotEmpty(f, targetGroupAnno, opt.TargetGroupName)
	setAnnoIfNotEmpty(f, defaultAnno, opt.Default)
	setAnnoIfNotEmpty(f, nameAnno, opt.Name)
	setAnnoIfNotEmpty(f, jsonAnno, opt.JSON)
	if opt.Env != "" {
		flagutil.SetEnvName(f, opt.Env)
	}
	setAnnoIfNotEmpty(f, flagUsageAnno, opt.FlagUsage)
	setAnnoIfNotEmpty(f, flagTypeAnno, opt.FlagType)
	setAnnoIfNotEmpty(f, shortAnno, opt.Short)
	setAnnoIfNotEmpty(f, longAnno, opt.Long)
}

func setAnnoIfNotEmpty[T ~string](f *pflag.Flag, key string, value T) {
	if value != "" {
		flagutil.SetAnnotation(f, key, string(value))
	}
}
