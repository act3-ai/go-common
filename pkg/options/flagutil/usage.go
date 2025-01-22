package flagutil

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
)

// UsageFormatOptions is used to format flag usage output.
type UsageFormatOptions struct {
	// Columns sets the column wrapping.
	Columns int
	// Indentation sets the leading indent for each line.
	Indentation *string
	// FormatFlagName is used to format the name of each flag.
	FormatFlagName func(flag *pflag.Flag, name string) string
	// FormatType is called to format the type of each flag.
	FormatType func(flag *pflag.Flag, typeName string) string
	// FormatValue is called to format flag values for defaults and no-op defaults for each flag.
	FormatValue func(flag *pflag.Flag, value string) string
	// LineFunc overrides all other functions.
	LineFunc func(flag *pflag.Flag) (line string, skip bool)
}

// FlagUsages returns a string containing the usage information for all flags in
// the FlagSet
func FlagUsages(f *pflag.FlagSet, opts UsageFormatOptions) string {
	if f == nil {
		return ""
	}

	indent := "  "
	if opts.Indentation != nil {
		indent = *opts.Indentation
	}

	buf := new(strings.Builder)

	lines := []string{}

	maxlen := 0
	f.VisitAll(func(flag *pflag.Flag) {
		if opts.LineFunc != nil {
			line, skip := opts.LineFunc(flag)
			if !skip {
				lines = append(lines, line)
			}
		}

		if flag.Hidden {
			return
		}

		line := indent
		line += fmtName(flag, opts)

		varname, usage := pflag.UnquoteUsage(flag)
		if varname != "" {
			if opts.FormatType != nil {
				varname = opts.FormatType(flag, varname)
			}
			line += " " + varname
		}
		line += fmtNoOptDefVal(flag, opts)

		// This special character will be replaced with spacing once the
		// correct alignment is calculated
		line += rhsStartChar
		if len(line) > maxlen {
			maxlen = len(line)
		}

		line += usage
		line += fmtDefault(flag, DefaultIsZeroValue(flag), opts)
		if len(flag.Deprecated) != 0 {
			line += fmt.Sprintf(" (DEPRECATED: %s)", flag.Deprecated)
		}

		lines = append(lines, line)
	})

	for _, line := range lines {
		sidx := strings.Index(line, rhsStartChar)
		spacing := strings.Repeat(" ", maxlen-sidx)
		// maxlen + 2 comes from + 1 for the \x00 and + 1 for the (deliberate) off-by-one in maxlen-sidx
		_, _ = fmt.Fprintln(buf, line[:sidx], spacing, wrap(maxlen+2, opts.Columns, line[sidx+1:]))
	}

	return buf.String()
}

const (
	// rhsStartChar marks the start of the RHS of a flag description
	rhsStartChar = "\x00"
)

func fmtName(flag *pflag.Flag, opts UsageFormatOptions) string {
	namer := opts.FormatFlagName
	if namer == nil {
		namer = func(flag *pflag.Flag, name string) string { return name }
	}
	if flag.Shorthand != "" && flag.ShorthandDeprecated == "" {
		return fmt.Sprintf("%s, %s", namer(flag, "-"+flag.Shorthand), namer(flag, "--"+flag.Name))
	}
	return fmt.Sprintf("    %s", namer(flag, "--"+flag.Name))
}

func fmtNoOptDefVal(flag *pflag.Flag, opts UsageFormatOptions) string {
	if flag.NoOptDefVal != "" {
		noOptDefVal := flag.NoOptDefVal
		if opts.FormatValue != nil {
			noOptDefVal = opts.FormatValue(flag, flag.NoOptDefVal)
		}
		switch flag.Value.Type() {
		case "string":
			return fmt.Sprintf("[=\"%s\"]", noOptDefVal)
		case "bool":
			if flag.NoOptDefVal != "true" {
				return fmt.Sprintf("[=%s]", noOptDefVal)
			}
		case "count":
			if flag.NoOptDefVal != "+1" {
				return fmt.Sprintf("[=%s]", noOptDefVal)
			}
		default:
			return fmt.Sprintf("[=%s]", noOptDefVal)
		}
	}
	return ""
}

func fmtDefault(flag *pflag.Flag, defaultIsZeroValue bool, opts UsageFormatOptions) string {
	if defaultIsZeroValue {
		return ""
	}
	defValue := flag.DefValue
	if opts.FormatValue != nil {
		defValue = opts.FormatValue(flag, defValue)
	}
	if flag.Value.Type() == "string" {
		return fmt.Sprintf("(default %q)", defValue)
	}
	return fmt.Sprintf("(default %s)", defValue)
}

// DefaultIsZeroValue returns true if the default value for this flag represents
// a zero value.
//
// This is a best effort guess.
func DefaultIsZeroValue(f *pflag.Flag) bool {
	switch f.Value.Type() {
	case "bool":
		return f.DefValue == "false"
	case "duration":
		// Beginning in Go 1.7, duration zero values are "0s"
		return f.DefValue == "0" || f.DefValue == "0s"
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64", "count":
		return f.DefValue == "0"
	case "string":
		return f.DefValue == ""
	case "ip", "ipMask", "ipNet":
		return f.DefValue == "<nil>"
	case "intSlice", "stringSlice", "stringArray":
		return f.DefValue == "[]"
	default:
		switch f.Value.String() {
		case "false":
			return true
		case "<nil>":
			return true
		case "":
			return true
		case "0":
			return true
		}
		return false
	}
}

// Splits the string `s` on whitespace into an initial substring up to
// `i` runes in length and the remainder. Will go `slop` over `i` if
// that encompasses the entire string (which allows the caller to
// avoid short orphan words on the final line).
func wrapN(i, slop int, s string) (string, string) {
	if i+slop > len(s) {
		return s, ""
	}

	w := strings.LastIndexAny(s[:i], " \t\n")
	if w <= 0 {
		return s, ""
	}
	nlPos := strings.LastIndex(s[:i], "\n")
	if nlPos > 0 && nlPos < w {
		return s[:nlPos], s[nlPos+1:]
	}
	return s[:w], s[w+1:]
}

// Wraps the string `s` to a maximum width `w` with leading indent
// `i`. The first line is not indented (this is assumed to be done by
// caller). Pass `w` == 0 to do no wrapping
func wrap(i, w int, s string) string {
	if w == 0 {
		return strings.ReplaceAll(s, "\n", "\n"+strings.Repeat(" ", i))
	}

	// space between indent i and end of line width w into which
	// we should wrap the text.
	wrap := w - i

	var r, l string

	// Not enough space for sensible wrapping. Wrap as a block on
	// the next line instead.
	if wrap < 24 {
		i = 16
		wrap = w - i
		r += "\n" + strings.Repeat(" ", i)
	}
	// If still not enough space then don't even try to wrap.
	if wrap < 24 {
		return strings.ReplaceAll(s, "\n", r)
	}

	// Try to avoid short orphan words on the final line, by
	// allowing wrapN to go a bit over if that would fit in the
	// remainder of the line.
	slop := 5
	wrap -= slop

	// Handle first line, which is indented by the caller (or the
	// special case above)
	l, s = wrapN(wrap, slop, s)
	r += strings.ReplaceAll(l, "\n", "\n"+strings.Repeat(" ", i))

	// Now wrap the rest
	for s != "" {
		var t string

		t, s = wrapN(wrap, slop, s)
		r = r + "\n" + strings.Repeat(" ", i) + strings.ReplaceAll(t, "\n", "\n"+strings.Repeat(" ", i))
	}

	return r
}
