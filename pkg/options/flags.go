package options

import (
	"github.com/spf13/pflag"

	"gitlab.com/act3-ai/asce/go-common/pkg/options/flagutil"
)

// StringVar creates a flag for the option.
func StringVar(f *pflag.FlagSet, p *string, value string, opts *Option) *pflag.Flag {
	var flag *pflag.Flag
	if opts.FlagShorthand == "" {
		flag = flagutil.StringVar(f, p, opts.Flag, value, opts.formattedFlagUsage())
	} else {
		flag = flagutil.StringVarP(f, p, opts.Flag, opts.FlagShorthand, value, opts.formattedFlagUsage())
	}
	withOptionConfig(flag, opts)
	return flag
}

// IntVar creates a flag for the option.
func IntVar(f *pflag.FlagSet, p *int, value int, opts *Option) *pflag.Flag {
	var flag *pflag.Flag
	if opts.FlagShorthand == "" {
		flag = flagutil.IntVar(f, p, opts.Flag, value, opts.formattedFlagUsage())
	} else {
		flag = flagutil.IntVarP(f, p, opts.Flag, opts.FlagShorthand, value, opts.formattedFlagUsage())
	}
	withOptionConfig(flag, opts)
	return flag
}

// Int64Var creates a flag for the option.
func Int64Var(f *pflag.FlagSet, p *int64, value int64, opts *Option) *pflag.Flag {
	var flag *pflag.Flag
	if opts.FlagShorthand == "" {
		flag = flagutil.Int64Var(f, p, opts.Flag, value, opts.formattedFlagUsage())
	} else {
		flag = flagutil.Int64VarP(f, p, opts.Flag, opts.FlagShorthand, value, opts.formattedFlagUsage())
	}
	withOptionConfig(flag, opts)
	return flag
}

// BoolVar creates a flag for the option.
func BoolVar(f *pflag.FlagSet, p *bool, value bool, opts *Option) *pflag.Flag {
	var flag *pflag.Flag
	if opts.FlagShorthand == "" {
		flag = flagutil.BoolVar(f, p, opts.Flag, value, opts.formattedFlagUsage())
	} else {
		flag = flagutil.BoolVarP(f, p, opts.Flag, opts.FlagShorthand, value, opts.formattedFlagUsage())
	}
	withOptionConfig(flag, opts)
	return flag
}

// StringSliceVar creates a flag for the option.
func StringSliceVar(f *pflag.FlagSet, p *[]string, value []string, opts *Option) *pflag.Flag {
	var flag *pflag.Flag
	if opts.FlagShorthand == "" {
		flag = flagutil.StringSliceVar(f, p, opts.Flag, value, opts.formattedFlagUsage())
	} else {
		flag = flagutil.StringSliceVarP(f, p, opts.Flag, opts.FlagShorthand, value, opts.formattedFlagUsage())
	}
	withOptionConfig(flag, opts)
	return flag
}

// StringToIntVar creates a flag for the option.
func StringToIntVar(f *pflag.FlagSet, p *map[string]int, value map[string]int, opts *Option) *pflag.Flag {
	var flag *pflag.Flag
	if opts.FlagShorthand == "" {
		flag = flagutil.StringToIntVar(f, p, opts.Flag, value, opts.formattedFlagUsage())
	} else {
		flag = flagutil.StringToIntVarP(f, p, opts.Flag, opts.FlagShorthand, value, opts.formattedFlagUsage())
	}
	withOptionConfig(flag, opts)
	return flag
}

// StringToBoolVar creates a flag for the option.
func StringToBoolVar(f *pflag.FlagSet, p *map[string]bool, value map[string]bool, opts *Option) *pflag.Flag {
	var flag *pflag.Flag
	if opts.FlagShorthand == "" {
		flag = flagutil.StringToBoolVar(f, p, opts.Flag, value, opts.formattedFlagUsage())
	} else {
		flag = flagutil.StringToBoolVarP(f, p, opts.Flag, opts.FlagShorthand, value, opts.formattedFlagUsage())
	}
	withOptionConfig(flag, opts)
	return flag
}

// StringToOptStringVar creates a flag for the option.
func StringToOptStringVar(f *pflag.FlagSet, p *map[string]*string, value map[string]*string, opts *Option) *pflag.Flag {
	var flag *pflag.Flag
	if opts.FlagShorthand == "" {
		flag = flagutil.StringToOptStringVar(f, p, opts.Flag, value, opts.formattedFlagUsage())
	} else {
		flag = flagutil.StringToOptStringVarP(f, p, opts.Flag, opts.FlagShorthand, value, opts.formattedFlagUsage())
	}
	withOptionConfig(flag, opts)
	return flag
}
