package options

import (
	"github.com/spf13/pflag"

	"gitlab.com/act3-ai/asce/go-common/pkg/options/flagutil"
)

/*
This file contains flag creation functions that create a flag matching an Option's definition.

These functions use the flagutil package which combines the flag creation and flag lookup step.

If there are additional flags defined by pflag that you require, please add them in the same fashion as the existing functions.
*/

/* Flag types specific to the flagutil package */

// StringToBoolVar creates a flag for the option.
func StringToBoolVar(f *pflag.FlagSet, p *map[string]bool, value map[string]bool, opts *Option) *pflag.Flag {
	return optionFlag(f, p, value, opts, flagutil.StringToBoolVar, flagutil.StringToBoolVarP)
}

// StringToOptStringVar creates a flag for the option.
func StringToOptStringVar(f *pflag.FlagSet, p *map[string]*string, value map[string]*string, opts *Option) *pflag.Flag {
	return optionFlag(f, p, value, opts, flagutil.StringToOptStringVar, flagutil.StringToOptStringVarP)
}

/* Generic value flag types */

// Var creates a flag for the option.
func Var(f *pflag.FlagSet, value pflag.Value, opts *Option) *pflag.Flag {
	var flag *pflag.Flag
	if opts.FlagShorthand == "" {
		flag = flagutil.Var(f, value, opts.Flag, opts.formattedFlagUsage())
	} else {
		flag = flagutil.VarP(f, value, opts.Flag, opts.FlagShorthand, opts.formattedFlagUsage())
	}
	withOptionConfig(flag, opts)
	return flag
}

/* Bool flag types */

// BoolVar creates a flag for the option.
func BoolVar(f *pflag.FlagSet, p *bool, value bool, opts *Option) *pflag.Flag {
	return optionFlag(f, p, value, opts, flagutil.BoolVar, flagutil.BoolVarP)
}

/* Bytes flag types */

/* Count flag types */

// CountVar creates a flag for the option.
func CountVar(f *pflag.FlagSet, p *int, opts *Option) *pflag.Flag {
	var flag *pflag.Flag
	if opts.FlagShorthand == "" {
		flag = flagutil.CountVar(f, p, opts.Flag, opts.formattedFlagUsage())
	} else {
		flag = flagutil.CountVarP(f, p, opts.Flag, opts.FlagShorthand, opts.formattedFlagUsage())
	}
	withOptionConfig(flag, opts)
	return flag
}

/* Duration flag types */

/* Float flag types */

/* IP flag types */

/* Int flag types */

// IntVar creates a flag for the option.
func IntVar(f *pflag.FlagSet, p *int, value int, opts *Option) *pflag.Flag {
	return optionFlag(f, p, value, opts, flagutil.IntVar, flagutil.IntVarP)
}

// Int64Var creates a flag for the option.
func Int64Var(f *pflag.FlagSet, p *int64, value int64, opts *Option) *pflag.Flag {
	return optionFlag(f, p, value, opts, flagutil.Int64Var, flagutil.Int64VarP)
}

/* String flag types */

// StringVar creates a flag for the option.
func StringVar(f *pflag.FlagSet, p *string, value string, opts *Option) *pflag.Flag {
	return optionFlag(f, p, value, opts, flagutil.StringVar, flagutil.StringVarP)
}

// StringSliceVar creates a flag for the option.
func StringSliceVar(f *pflag.FlagSet, p *[]string, value []string, opts *Option) *pflag.Flag {
	return optionFlag(f, p, value, opts, flagutil.StringSliceVar, flagutil.StringSliceVarP)
}

/* Map flag types */

// StringToIntVar creates a flag for the option.
func StringToIntVar(f *pflag.FlagSet, p *map[string]int, value map[string]int, opts *Option) *pflag.Flag {
	return optionFlag(f, p, value, opts, flagutil.StringToIntVar, flagutil.StringToIntVarP)
}

// StringToInt64Var creates a flag for the option.
func StringToInt64Var(f *pflag.FlagSet, p *map[string]int64, value map[string]int64, opts *Option) *pflag.Flag {
	return optionFlag(f, p, value, opts, flagutil.StringToInt64Var, flagutil.StringToInt64VarP)
}

// StringToStringVar creates a flag for the option.
func StringToStringVar(f *pflag.FlagSet, p *map[string]string, value map[string]string, opts *Option) *pflag.Flag {
	return optionFlag(f, p, value, opts, flagutil.StringToStringVar, flagutil.StringToStringVarP)
}

/* Uint flag types */

// optionFlag creates a flag for the option.
func optionFlag[T any](f *pflag.FlagSet, p *T, value T, opts *Option,
	createVar func(*pflag.FlagSet, *T, string, T, string) *pflag.Flag,
	createVarP func(*pflag.FlagSet, *T, string, string, T, string) *pflag.Flag,
) *pflag.Flag {
	var flag *pflag.Flag
	if opts.FlagShorthand == "" {
		flag = createVar(f, p, opts.Flag, value, opts.formattedFlagUsage())
	} else {
		flag = createVarP(f, p, opts.Flag, opts.FlagShorthand, value, opts.formattedFlagUsage())
	}
	withOptionConfig(flag, opts)
	return flag
}
