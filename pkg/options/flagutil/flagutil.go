// Package flagutil defines utilities for registering and parsing command line flags.
package flagutil

import "github.com/spf13/pflag"

/*
This file contains flag creation functions that combine the flag creation and flag lookup step.

If there are additional flags defined by pflag that you require, please add them in the same fashion as the existing functions.
*/

/*
type flagVarFunc[T any] func(p *T, name string, value T, usage string)

func flagVar[T any](flagFunc flagVarFunc[T], f *pflag.FlagSet, p *T, name string, value T, usage string) *pflag.Flag {
	flagFunc(p, name, value, usage)
	return f.Lookup(name)
}

// flagVarPFunc represents a function that creates a flag.
type flagVarPFunc[T any] = func(p *T, name, shorthand string, value T, usage string)

func flagVarP[T any](flagFunc flagVarPFunc[T], f *pflag.FlagSet, p *T, name, shorthand string, value T, usage string) *pflag.Flag {
	flagFunc(p, name, shorthand, value, usage)
	return f.Lookup(name)
}
*/

/* Flag types specific to this package */

// StringToBoolVar creates a [pflag.Flag].
func StringToBoolVar(f *pflag.FlagSet, p *map[string]bool, name string, value map[string]bool, usage string) *pflag.Flag {
	return Var(f, newStringToBoolValue(value, p), name, usage)
}

// StringToBoolVarP creates a [pflag.Flag].
func StringToBoolVarP(f *pflag.FlagSet, p *map[string]bool, name, shorthand string, value map[string]bool, usage string) *pflag.Flag {
	return VarP(f, newStringToBoolValue(value, p), name, shorthand, usage)
}

// StringToOptStringVar creates a [pflag.Flag].
func StringToOptStringVar(f *pflag.FlagSet, p *map[string]*string, name string, value map[string]*string, usage string) *pflag.Flag {
	return Var(f, newStringToOptStringValue(value, p), name, usage)
}

// StringToOptStringVarP creates a [pflag.Flag].
func StringToOptStringVarP(f *pflag.FlagSet, p *map[string]*string, name, shorthand string, value map[string]*string, usage string) *pflag.Flag {
	return VarP(f, newStringToOptStringValue(value, p), name, shorthand, usage)
}

/* Generic value flag types */

// Var creates a [pflag.Flag].
func Var(f *pflag.FlagSet, value pflag.Value, name string, usage string) *pflag.Flag {
	f.Var(value, name, usage)
	return f.Lookup(name)
}

// VarP creates a [pflag.Flag].
func VarP(f *pflag.FlagSet, value pflag.Value, name, shorthand string, usage string) *pflag.Flag {
	f.VarP(value, name, shorthand, usage)
	return f.Lookup(name)
}

/* Bool flag types */
// BoolSlice

// BoolVar creates a [pflag.Flag].
func BoolVar(f *pflag.FlagSet, p *bool, name string, value bool, usage string) *pflag.Flag {
	f.BoolVar(p, name, value, usage)
	return f.Lookup(name)
}

// BoolVarP creates a [pflag.Flag].
func BoolVarP(f *pflag.FlagSet, p *bool, name, shorthand string, value bool, usage string) *pflag.Flag {
	f.BoolVarP(p, name, shorthand, value, usage)
	return f.Lookup(name)
}

/* Bytes flag types */
// BytesBase64
// BytesHex

/* Count flag types */

// CountVar creates a [pflag.Flag].
func CountVar(f *pflag.FlagSet, p *int, name string, usage string) *pflag.Flag {
	f.CountVar(p, name, usage)
	return f.Lookup(name)
}

// CountVarP creates a [pflag.Flag].
func CountVarP(f *pflag.FlagSet, p *int, name, shorthand string, usage string) *pflag.Flag {
	f.CountVarP(p, name, shorthand, usage)
	return f.Lookup(name)
}

/* Duration flag types */
// Duration
// DurationSlice

/* Float flag types */

// Float32Var creates a [pflag.Flag].
func Float32Var(f *pflag.FlagSet, p *float32, name string, value float32, usage string) *pflag.Flag {
	f.Float32Var(p, name, value, usage)
	return f.Lookup(name)
}

// Float32VarP creates a [pflag.Flag].
func Float32VarP(f *pflag.FlagSet, p *float32, name, shorthand string, value float32, usage string) *pflag.Flag {
	f.Float32VarP(p, name, shorthand, value, usage)
	return f.Lookup(name)
}

// Float32SliceVar creates a [pflag.Flag].
func Float32SliceVar(f *pflag.FlagSet, p *[]float32, name string, value []float32, usage string) *pflag.Flag {
	f.Float32SliceVar(p, name, value, usage)
	return f.Lookup(name)
}

// Float32SliceVarP creates a [pflag.Flag].
func Float32SliceVarP(f *pflag.FlagSet, p *[]float32, name, shorthand string, value []float32, usage string) *pflag.Flag {
	f.Float32SliceVarP(p, name, shorthand, value, usage)
	return f.Lookup(name)
}

// Float64Var creates a [pflag.Flag].
func Float64Var(f *pflag.FlagSet, p *float64, name string, value float64, usage string) *pflag.Flag {
	f.Float64Var(p, name, value, usage)
	return f.Lookup(name)
}

// Float64VarP creates a [pflag.Flag].
func Float64VarP(f *pflag.FlagSet, p *float64, name, shorthand string, value float64, usage string) *pflag.Flag {
	f.Float64VarP(p, name, shorthand, value, usage)
	return f.Lookup(name)
}

// Float64SliceVar creates a [pflag.Flag].
func Float64SliceVar(f *pflag.FlagSet, p *[]float64, name string, value []float64, usage string) *pflag.Flag {
	f.Float64SliceVar(p, name, value, usage)
	return f.Lookup(name)
}

// Float64SliceVarP creates a [pflag.Flag].
func Float64SliceVarP(f *pflag.FlagSet, p *[]float64, name, shorthand string, value []float64, usage string) *pflag.Flag {
	f.Float64SliceVarP(p, name, shorthand, value, usage)
	return f.Lookup(name)
}

/* IP flag types */
// IP
// IPMask
// IPNet
// IPSlice

/* Int flag types */
// Int8
// Int8Slice
// Int16
// Int16Slice
// Int32
// Int32Slice

// IntVar creates a [pflag.Flag].
func IntVar(f *pflag.FlagSet, p *int, name string, value int, usage string) *pflag.Flag {
	f.IntVar(p, name, value, usage)
	return f.Lookup(name)
}

// IntVarP creates a [pflag.Flag].
func IntVarP(f *pflag.FlagSet, p *int, name, shorthand string, value int, usage string) *pflag.Flag {
	f.IntVarP(p, name, shorthand, value, usage)
	return f.Lookup(name)
}

// IntSliceVar creates a [pflag.Flag].
func IntSliceVar(f *pflag.FlagSet, p *[]int, name string, value []int, usage string) *pflag.Flag {
	f.IntSliceVar(p, name, value, usage)
	return f.Lookup(name)
}

// IntSliceVarP creates a [pflag.Flag].
func IntSliceVarP(f *pflag.FlagSet, p *[]int, name, shorthand string, value []int, usage string) *pflag.Flag {
	f.IntSliceVarP(p, name, shorthand, value, usage)
	return f.Lookup(name)
}

// Int8Var creates a [pflag.Flag].
func Int8Var(f *pflag.FlagSet, p *int8, name string, value int8, usage string) *pflag.Flag {
	f.Int8Var(p, name, value, usage)
	return f.Lookup(name)
}

// Int8VarP creates a [pflag.Flag].
func Int8VarP(f *pflag.FlagSet, p *int8, name, shorthand string, value int8, usage string) *pflag.Flag {
	f.Int8VarP(p, name, shorthand, value, usage)
	return f.Lookup(name)
}

// Int16Var creates a [pflag.Flag].
func Int16Var(f *pflag.FlagSet, p *int16, name string, value int16, usage string) *pflag.Flag {
	f.Int16Var(p, name, value, usage)
	return f.Lookup(name)
}

// Int16VarP creates a [pflag.Flag].
func Int16VarP(f *pflag.FlagSet, p *int16, name, shorthand string, value int16, usage string) *pflag.Flag {
	f.Int16VarP(p, name, shorthand, value, usage)
	return f.Lookup(name)
}

// Int32Var creates a [pflag.Flag].
func Int32Var(f *pflag.FlagSet, p *int32, name string, value int32, usage string) *pflag.Flag {
	f.Int32Var(p, name, value, usage)
	return f.Lookup(name)
}

// Int32VarP creates a [pflag.Flag].
func Int32VarP(f *pflag.FlagSet, p *int32, name, shorthand string, value int32, usage string) *pflag.Flag {
	f.Int32VarP(p, name, shorthand, value, usage)
	return f.Lookup(name)
}

// Int32SliceVar creates a [pflag.Flag].
func Int32SliceVar(f *pflag.FlagSet, p *[]int32, name string, value []int32, usage string) *pflag.Flag {
	f.Int32SliceVar(p, name, value, usage)
	return f.Lookup(name)
}

// Int32SliceVarP creates a [pflag.Flag].
func Int32SliceVarP(f *pflag.FlagSet, p *[]int32, name, shorthand string, value []int32, usage string) *pflag.Flag {
	f.Int32SliceVarP(p, name, shorthand, value, usage)
	return f.Lookup(name)
}

// Int64Var creates a [pflag.Flag].
func Int64Var(f *pflag.FlagSet, p *int64, name string, value int64, usage string) *pflag.Flag {
	f.Int64Var(p, name, value, usage)
	return f.Lookup(name)
}

// Int64VarP creates a [pflag.Flag].
func Int64VarP(f *pflag.FlagSet, p *int64, name, shorthand string, value int64, usage string) *pflag.Flag {
	f.Int64VarP(p, name, shorthand, value, usage)
	return f.Lookup(name)
}

// Int64SliceVar creates a [pflag.Flag].
func Int64SliceVar(f *pflag.FlagSet, p *[]int64, name string, value []int64, usage string) *pflag.Flag {
	f.Int64SliceVar(p, name, value, usage)
	return f.Lookup(name)
}

// Int64SliceVarP creates a [pflag.Flag].
func Int64SliceVarP(f *pflag.FlagSet, p *[]int64, name, shorthand string, value []int64, usage string) *pflag.Flag {
	f.Int64SliceVarP(p, name, shorthand, value, usage)
	return f.Lookup(name)
}

/* String flag types */
// StringArray

// StringVar creates a [pflag.Flag].
func StringVar(f *pflag.FlagSet, p *string, name string, value string, usage string) *pflag.Flag {
	f.StringVar(p, name, value, usage)
	return f.Lookup(name)
}

// StringVarP creates a [pflag.Flag].
func StringVarP(f *pflag.FlagSet, p *string, name, shorthand string, value string, usage string) *pflag.Flag {
	f.StringVarP(p, name, shorthand, value, usage)
	return f.Lookup(name)
}

// StringSliceVar creates a [pflag.Flag].
func StringSliceVar(f *pflag.FlagSet, p *[]string, name string, value []string, usage string) *pflag.Flag {
	f.StringSliceVar(p, name, value, usage)
	return f.Lookup(name)
}

// StringSliceVarP creates a [pflag.Flag].
func StringSliceVarP(f *pflag.FlagSet, p *[]string, name, shorthand string, value []string, usage string) *pflag.Flag {
	f.StringSliceVarP(p, name, shorthand, value, usage)
	return f.Lookup(name)
}

/* Map flag types */

// StringToIntVar creates a [pflag.Flag].
func StringToIntVar(f *pflag.FlagSet, p *map[string]int, name string, value map[string]int, usage string) *pflag.Flag {
	f.StringToIntVar(p, name, value, usage)
	return f.Lookup(name)
}

// StringToIntVarP creates a [pflag.Flag].
func StringToIntVarP(f *pflag.FlagSet, p *map[string]int, name, shorthand string, value map[string]int, usage string) *pflag.Flag {
	f.StringToIntVarP(p, name, shorthand, value, usage)
	return f.Lookup(name)
}

// StringToInt64Var creates a [pflag.Flag].
func StringToInt64Var(f *pflag.FlagSet, p *map[string]int64, name string, value map[string]int64, usage string) *pflag.Flag {
	f.StringToInt64Var(p, name, value, usage)
	return f.Lookup(name)
}

// StringToInt64VarP creates a [pflag.Flag].
func StringToInt64VarP(f *pflag.FlagSet, p *map[string]int64, name, shorthand string, value map[string]int64, usage string) *pflag.Flag {
	f.StringToInt64VarP(p, name, shorthand, value, usage)
	return f.Lookup(name)
}

// StringToStringVar creates a [pflag.Flag].
func StringToStringVar(f *pflag.FlagSet, p *map[string]string, name string, value map[string]string, usage string) *pflag.Flag {
	f.StringToStringVar(p, name, value, usage)
	return f.Lookup(name)
}

// StringToStringVarP creates a [pflag.Flag].
func StringToStringVarP(f *pflag.FlagSet, p *map[string]string, name, shorthand string, value map[string]string, usage string) *pflag.Flag {
	f.StringToStringVarP(p, name, shorthand, value, usage)
	return f.Lookup(name)
}

/* Uint flag types */
// Uint
// UintSlice
// Uint8
// Uint16
// Uint32
// Uint64
