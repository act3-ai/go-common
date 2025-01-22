// Package flagutil defines utilities for registering and parsing command line flags.
package flagutil

import "github.com/spf13/pflag"

// StringVar creates a [pflag.Flag].
func StringVar(f *pflag.FlagSet, p *string, name string, value string, usage string) *pflag.Flag {
	f.StringVar(p, name, value, usage)
	return f.Lookup(name)
}

// StringVarP creates a [pflag.Flag].
func StringVarP(f *pflag.FlagSet, p *string, name string, shorthand string, value string, usage string) *pflag.Flag {
	f.StringVarP(p, name, shorthand, value, usage)
	return f.Lookup(name)
}

// CountVar creates a [pflag.Flag].
func CountVar(f *pflag.FlagSet, p *int, name string, usage string) *pflag.Flag {
	f.CountVar(p, name, usage)
	return f.Lookup(name)
}

// CountVarP creates a [pflag.Flag].
func CountVarP(f *pflag.FlagSet, p *int, name string, shorthand string, usage string) *pflag.Flag {
	f.CountVarP(p, name, shorthand, usage)
	return f.Lookup(name)
}

// IntVar creates a [pflag.Flag].
func IntVar(f *pflag.FlagSet, p *int, name string, value int, usage string) *pflag.Flag {
	f.IntVar(p, name, value, usage)
	return f.Lookup(name)
}

// IntVarP creates a [pflag.Flag].
func IntVarP(f *pflag.FlagSet, p *int, name string, shorthand string, value int, usage string) *pflag.Flag {
	f.IntVarP(p, name, shorthand, value, usage)
	return f.Lookup(name)
}

// Int64Var creates a [pflag.Flag].
func Int64Var(f *pflag.FlagSet, p *int64, name string, value int64, usage string) *pflag.Flag {
	f.Int64Var(p, name, value, usage)
	return f.Lookup(name)
}

// Int64VarP creates a [pflag.Flag].
func Int64VarP(f *pflag.FlagSet, p *int64, name string, shorthand string, value int64, usage string) *pflag.Flag {
	f.Int64VarP(p, name, shorthand, value, usage)
	return f.Lookup(name)
}

// BoolVar creates a [pflag.Flag].
func BoolVar(f *pflag.FlagSet, p *bool, name string, value bool, usage string) *pflag.Flag {
	f.BoolVar(p, name, value, usage)
	return f.Lookup(name)
}

// BoolVarP creates a [pflag.Flag].
func BoolVarP(f *pflag.FlagSet, p *bool, name string, shorthand string, value bool, usage string) *pflag.Flag {
	f.BoolVarP(p, name, shorthand, value, usage)
	return f.Lookup(name)
}

// StringSliceVar creates a [pflag.Flag].
func StringSliceVar(f *pflag.FlagSet, p *[]string, name string, value []string, usage string) *pflag.Flag {
	f.StringSliceVar(p, name, value, usage)
	return f.Lookup(name)
}

// StringSliceVarP creates a [pflag.Flag].
func StringSliceVarP(f *pflag.FlagSet, p *[]string, name string, shorthand string, value []string, usage string) *pflag.Flag {
	f.StringSliceVarP(p, name, shorthand, value, usage)
	return f.Lookup(name)
}

// StringToIntVar creates a [pflag.Flag].
func StringToIntVar(f *pflag.FlagSet, p *map[string]int, name string, value map[string]int, usage string) *pflag.Flag {
	f.StringToIntVar(p, name, value, usage)
	return f.Lookup(name)
}

// StringToIntVarP creates a [pflag.Flag].
func StringToIntVarP(f *pflag.FlagSet, p *map[string]int, name string, shorthand string, value map[string]int, usage string) *pflag.Flag {
	f.StringToIntVarP(p, name, shorthand, value, usage)
	return f.Lookup(name)
}

// StringToBoolVar creates a [pflag.Flag].
func StringToBoolVar(f *pflag.FlagSet, p *map[string]bool, name string, value map[string]bool, usage string) *pflag.Flag {
	f.VarP(newStringToBoolValue(value, p), name, "", usage)
	return f.Lookup(name)
}

// StringToBoolVarP creates a [pflag.Flag].
func StringToBoolVarP(f *pflag.FlagSet, p *map[string]bool, name, shorthand string, value map[string]bool, usage string) *pflag.Flag {
	f.VarP(newStringToBoolValue(value, p), name, shorthand, usage)
	return f.Lookup(name)
}

// StringToOptStringVar creates a [pflag.Flag].
func StringToOptStringVar(f *pflag.FlagSet, p *map[string]*string, name string, value map[string]*string, usage string) *pflag.Flag {
	f.VarP(newStringToOptStringValue(value, p), name, "", usage)
	return f.Lookup(name)
}

// StringToOptStringVarP creates a [pflag.Flag].
func StringToOptStringVarP(f *pflag.FlagSet, p *map[string]*string, name, shorthand string, value map[string]*string, usage string) *pflag.Flag {
	f.VarP(newStringToOptStringValue(value, p), name, shorthand, usage)
	return f.Lookup(name)
}
