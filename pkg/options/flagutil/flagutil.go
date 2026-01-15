// Package flagutil defines utilities for registering and parsing command line flags.
package flagutil

import (
	"time"

	"github.com/spf13/pflag"
)

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

// BoolSliceVar creates a [pflag.Flag].
func BoolSliceVar(f *pflag.FlagSet, p *[]bool, name string, value []bool, usage string) *pflag.Flag {
	f.BoolSliceVar(p, name, value, usage)
	return f.Lookup(name)
}

// BoolSliceVarP creates a [pflag.Flag].
func BoolSliceVarP(f *pflag.FlagSet, p *[]bool, name, shorthand string, value []bool, usage string) *pflag.Flag {
	f.BoolSliceVarP(p, name, shorthand, value, usage)
	return f.Lookup(name)
}

// BoolFunc creates a [pflag.Flag].
func BoolFunc(f *pflag.FlagSet, name string, usage string, fn func(string) error) *pflag.Flag {
	f.BoolFunc(name, usage, fn)
	return f.Lookup(name)
}

// BoolFuncP creates a [pflag.Flag].
func BoolFuncP(f *pflag.FlagSet, name, shorthand string, usage string, fn func(string) error) *pflag.Flag {
	f.BoolFuncP(name, shorthand, usage, fn)
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

// DurationVar creates a [pflag.Flag].
func DurationVar(f *pflag.FlagSet, p *time.Duration, name string, value time.Duration, usage string) *pflag.Flag {
	f.DurationVar(p, name, value, usage)
	return f.Lookup(name)
}

// DurationVarP creates a [pflag.Flag].
func DurationVarP(f *pflag.FlagSet, p *time.Duration, name, shorthand string, value time.Duration, usage string) *pflag.Flag {
	f.DurationVarP(p, name, shorthand, value, usage)
	return f.Lookup(name)
}

// DurationSliceVar creates a [pflag.Flag].
func DurationSliceVar(f *pflag.FlagSet, p *[]time.Duration, name string, value []time.Duration, usage string) *pflag.Flag {
	f.DurationSliceVar(p, name, value, usage)
	return f.Lookup(name)
}

// DurationSliceVarP creates a [pflag.Flag].
func DurationSliceVarP(f *pflag.FlagSet, p *[]time.Duration, name, shorthand string, value []time.Duration, usage string) *pflag.Flag {
	f.DurationSliceVarP(p, name, shorthand, value, usage)
	return f.Lookup(name)
}

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

/* Func flag types */

// Func creates a [pflag.Flag].
func Func(f *pflag.FlagSet, name string, usage string, fn func(string) error) *pflag.Flag {
	f.Func(name, usage, fn)
	return f.Lookup(name)
}

// FuncP creates a [pflag.Flag].
func FuncP(f *pflag.FlagSet, name, shorthand string, usage string, fn func(string) error) *pflag.Flag {
	f.FuncP(name, shorthand, usage, fn)
	return f.Lookup(name)
}

/* IP flag types */
// IP
// IPMask
// IPNet
// IPSlice

/* Int flag types */

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

// UintVar creates a [pflag.Flag].
func UintVar(f *pflag.FlagSet, p *uint, name string, value uint, usage string) *pflag.Flag {
	f.UintVar(p, name, value, usage)
	return f.Lookup(name)
}

// UintVarP creates a [pflag.Flag].
func UintVarP(f *pflag.FlagSet, p *uint, name, shorthand string, value uint, usage string) *pflag.Flag {
	f.UintVarP(p, name, shorthand, value, usage)
	return f.Lookup(name)
}

// UintSliceVar creates a [pflag.Flag].
func UintSliceVar(f *pflag.FlagSet, p *[]uint, name string, value []uint, usage string) *pflag.Flag {
	f.UintSliceVar(p, name, value, usage)
	return f.Lookup(name)
}

// UintSliceVarP creates a [pflag.Flag].
func UintSliceVarP(f *pflag.FlagSet, p *[]uint, name, shorthand string, value []uint, usage string) *pflag.Flag {
	f.UintSliceVarP(p, name, shorthand, value, usage)
	return f.Lookup(name)
}

// Uint8Var creates a [pflag.Flag].
func Uint8Var(f *pflag.FlagSet, p *uint8, name string, value uint8, usage string) *pflag.Flag {
	f.Uint8Var(p, name, value, usage)
	return f.Lookup(name)
}

// Uint8VarP creates a [pflag.Flag].
func Uint8VarP(f *pflag.FlagSet, p *uint8, name, shorthand string, value uint8, usage string) *pflag.Flag {
	f.Uint8VarP(p, name, shorthand, value, usage)
	return f.Lookup(name)
}

// Uint16Var creates a [pflag.Flag].
func Uint16Var(f *pflag.FlagSet, p *uint16, name string, value uint16, usage string) *pflag.Flag {
	f.Uint16Var(p, name, value, usage)
	return f.Lookup(name)
}

// Uint16VarP creates a [pflag.Flag].
func Uint16VarP(f *pflag.FlagSet, p *uint16, name, shorthand string, value uint16, usage string) *pflag.Flag {
	f.Uint16VarP(p, name, shorthand, value, usage)
	return f.Lookup(name)
}

// Uint32Var creates a [pflag.Flag].
func Uint32Var(f *pflag.FlagSet, p *uint32, name string, value uint32, usage string) *pflag.Flag {
	f.Uint32Var(p, name, value, usage)
	return f.Lookup(name)
}

// Uint32VarP creates a [pflag.Flag].
func Uint32VarP(f *pflag.FlagSet, p *uint32, name, shorthand string, value uint32, usage string) *pflag.Flag {
	f.Uint32VarP(p, name, shorthand, value, usage)
	return f.Lookup(name)
}

// Uint64Var creates a [pflag.Flag].
func Uint64Var(f *pflag.FlagSet, p *uint64, name string, value uint64, usage string) *pflag.Flag {
	f.Uint64Var(p, name, value, usage)
	return f.Lookup(name)
}

// Uint64VarP creates a [pflag.Flag].
func Uint64VarP(f *pflag.FlagSet, p *uint64, name, shorthand string, value uint64, usage string) *pflag.Flag {
	f.Uint64VarP(p, name, shorthand, value, usage)
	return f.Lookup(name)
}
