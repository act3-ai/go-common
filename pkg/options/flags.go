package options

import (
	"github.com/spf13/pflag"

	"github.com/act3-ai/go-common/pkg/options/flagutil"
)

/*
This file contains flag creation functions that create a flag matching an Option's definition.

These functions use the flagutil package which combines the flag creation and flag lookup step.

If there are additional flags defined by pflag that you require, please add them in the same fashion as the existing functions.
*/

// FlagFuncP is a function that creates a flag on the given FlagSet and returns it.
type FlagFuncP[T any] = func(f *pflag.FlagSet, p *T, name, shorthand string, value T, usage string) *pflag.Flag

// OptionFlag creates a flag for the option.
func OptionFlag[T any](f *pflag.FlagSet, p *T, value T, opts *Option,
	createVarP FlagFuncP[T],
) *pflag.Flag {
	flag := createVarP(f, p, opts.Flag, opts.FlagShorthand, value, opts.formattedFlagUsage())
	withOptionConfig(flag, opts)
	return flag
}

/* Flag types specific to the flagutil package */

// StringToBoolVar creates a flag for the option.
func StringToBoolVar(f *pflag.FlagSet, p *map[string]bool, value map[string]bool, opts *Option) *pflag.Flag {
	return OptionFlag(f, p, value, opts, flagutil.StringToBoolVarP)
}

// StringToOptStringVar creates a flag for the option.
func StringToOptStringVar(f *pflag.FlagSet, p *map[string]*string, value map[string]*string, opts *Option) *pflag.Flag {
	return OptionFlag(f, p, value, opts, flagutil.StringToOptStringVarP)
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
	return OptionFlag(f, p, value, opts, flagutil.BoolVarP)
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
	return OptionFlag(f, p, value, opts, flagutil.IntVarP)
}

// Int64Var creates a flag for the option.
func Int64Var(f *pflag.FlagSet, p *int64, value int64, opts *Option) *pflag.Flag {
	return OptionFlag(f, p, value, opts, flagutil.Int64VarP)
}

/* String flag types */

// StringVar creates a flag for the option.
func StringVar(f *pflag.FlagSet, p *string, value string, opts *Option) *pflag.Flag {
	return OptionFlag(f, p, value, opts, flagutil.StringVarP)
}

// StringSliceVar creates a flag for the option.
func StringSliceVar(f *pflag.FlagSet, p *[]string, value []string, opts *Option) *pflag.Flag {
	return OptionFlag(f, p, value, opts, flagutil.StringSliceVarP)
}

/* Map flag types */

// StringToIntVar creates a flag for the option.
func StringToIntVar(f *pflag.FlagSet, p *map[string]int, value map[string]int, opts *Option) *pflag.Flag {
	return OptionFlag(f, p, value, opts, flagutil.StringToIntVarP)
}

// StringToInt64Var creates a flag for the option.
func StringToInt64Var(f *pflag.FlagSet, p *map[string]int64, value map[string]int64, opts *Option) *pflag.Flag {
	return OptionFlag(f, p, value, opts, flagutil.StringToInt64VarP)
}

// StringToStringVar creates a flag for the option.
func StringToStringVar(f *pflag.FlagSet, p *map[string]string, value map[string]string, opts *Option) *pflag.Flag {
	return OptionFlag(f, p, value, opts, flagutil.StringToStringVarP)
}

/* Uint flag types */
