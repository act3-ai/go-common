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
		Env:             flagutil.GetFirstAnnotationOr(f, envAnno, ""),
		Flag:            f.Name,
		FlagShorthand:   f.Shorthand,
		FlagUsage:       flagutil.GetFirstAnnotationOr(f, flagUsageAnno, ""),
		FlagType:        flagutil.GetFirstAnnotationOr(f, flagTypeAnno, f.Value.Type()),
		Short:           flagutil.GetFirstAnnotationOr(f, shortAnno, ""),
		Long:            flagutil.GetFirstAnnotationOr(f, longAnno, ""),
	}
	return opt
}

// ParseEnvOverrides receives a flag set after it has been parsed and
// sets the flag values to environment variables if the flag defines an
// "env" annotation.
//
// Any parsing errors are logged at slog.LevelWarn and are discarded.
func ParseEnvOverrides(flagSet *pflag.FlagSet) {
	flagutil.ParseEnvOverrides(flagSet, envAnno)
}

const (
	typeAnno        = "type"      // annotation for options.Option.Type
	defaultAnno     = "default"   // annotation for options.Option.Default
	jsonAnno        = "json"      // annotation for options.Option.Path
	envAnno         = "env"       // annotation for options.Option.Env
	flagUsageAnno   = "flagUsage" // annotation for options.Option.FlagUsage
	flagTypeAnno    = "flagType"  // annotation for options.Option.FlagType
	shortAnno       = "short"     // annotation for options.Option.Short
	longAnno        = "long"      // annotation for options.Option.Long
	targetGroupAnno = "target"    // annotation for options.Option.TargetGroupName
	groupAnno       = "group"     // used to group flags
)

// withOptionConfig adds sets annotations on the flag from the option definition.
func withOptionConfig(f *pflag.Flag, opt *Option) {
	// Default some fields from the flag
	if opt.FlagUsage == "" {
		opt.FlagUsage = f.Usage
	}
	if opt.FlagType == "" {
		// Use UnqouteUsage function to derive type name from usage string
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
		flagutil.SetAnnotation(f, envAnno, opt.Env)
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
