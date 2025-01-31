package optionshelp

import (
	"fmt"
	"reflect"
	"strings"

	"gitlab.com/act3-ai/asce/go-common/pkg/options"
)

/*
| Option | Description |
| ------ | ----------- |
{{- range .Options }}
| {{ .MarkdownLink }} | {{ .ShortDescription }} |
{{- end }}
*/
func groupTable(g *options.Group) string {
	w := &strings.Builder{}

	rows := [][2]string{}

	for _, o := range g.Options {
		rows = append(rows, [2]string{
			o.MarkdownLink(), o.ShortDescription(),
		})
	}

	descMax := len("Description")
	for _, row := range rows {
		descLen := len(row[1])
		if descLen > descMax {
			descMax = descLen
		}
	}

	fmtRow := fmtRowFunc(descMax)

	_, _ = w.WriteString(fmtRow("Option", "Description"))
	_, _ = w.WriteString(fmt.Sprintf("| %s | %s |\n", strings.Repeat("-", nameMax), strings.Repeat("-", descMax)))
	for _, row := range rows {
		_, _ = w.WriteString(fmtRow(row[0], row[1]))
	}

	return w.String()
}

/*
| Name      | Value |
| --------- | ----- |
| type      | {{ .FormattedType }} |
{{- if eq .Type "map" }}
| keys      | string |
| values    | {{ .TargetLink | default "any" }} |
{{- else if eq .Type "list" }}
| values    | {{ .TargetLink | default "any" }} |
{{- end }}
{{- with .FormattedDefault }}
| default   | `{{ . }}` |
{{- end }}
{{- if .JSON }}
| json/yaml | `{{ .JSON }}` |
{{- end }}
{{- if or .Flag .FlagShorthand }}
| cli       | {{ with .Flag }}`--{{ . }}`{{ end }}{{ with .FlagShorthand }}, `-{{ . }}`{{ end }} |
{{- end }}
{{- if .Env }}
| env       | `{{ .Env }}` |
{{- end }}
*/
func optionTable(o *options.Option) string {
	w := &strings.Builder{}

	rows := [][2]string{
		{"type", o.FormattedType()},
	}

	if o.Type == options.StringMap {
		rows = append(rows, [2]string{
			"keys", "string",
		})
	}
	if o.Type == options.StringMap || o.Type == options.List {
		fvalues := o.TargetLink()
		if o.TargetLink() == "" {
			fvalues = "any"
		}
		rows = append(rows, [2]string{
			"values", fvalues,
		})
	}
	if fdefault := o.FormattedDefault(); fdefault != "" {
		rows = append(rows, [2]string{
			"default", "`" + fdefault + "`",
		})
	}
	if o.JSON != "" {
		rows = append(rows, [2]string{
			"json/yaml", "`" + o.JSON + "`",
		})
	}
	if o.Flag != "" || o.FlagShorthand != "" {
		var fflag string
		switch {
		case o.Flag != "" && o.FlagShorthand != "":
			fflag = fmt.Sprintf("`--%s`, `-%s`", o.Flag, o.FlagShorthand)
		case o.Flag != "":
			fflag = fmt.Sprintf("`--%s`", o.Flag)
		case o.FlagShorthand != "":
			fflag = fmt.Sprintf("`-%s`", o.FlagShorthand)
		}
		rows = append(rows, [2]string{
			"cli", fflag,
		})
	}
	if o.Env != "" {
		rows = append(rows, [2]string{
			"env", "`" + o.Env + "`",
		})
	}

	valueMax := 0
	for _, row := range rows {
		valueLen := len(row[1])
		if valueLen > valueMax {
			valueMax = valueLen
		}
	}

	fmtRow := fmtRowFunc(valueMax)

	_, _ = w.WriteString(fmtRow("Name", "Value"))
	_, _ = w.WriteString(fmt.Sprintf("| %s | %s |\n", strings.Repeat("-", nameMax), strings.Repeat("-", valueMax)))
	for _, row := range rows {
		_, _ = w.WriteString(fmtRow(row[0], row[1]))
	}

	return w.String()
}

const nameMax = 9

func fmtRowFunc(valueMax int) func(name, value string) string {
	fmtName := fmt.Sprintf("%%-%ds", nameMax)
	fmtValue := fmt.Sprintf("%%-%ds", valueMax)
	fmtRow := fmt.Sprintf("| %s | %s |\n", fmtName, fmtValue)
	return func(name, value string) string {
		return fmt.Sprintf(fmtRow, name, value)
	}
}

/* Vendored functions from sprig to avoid bringing in a dependency */

// dfault checks whether `given` is set, and returns default if not set.
//
// This returns `d` if `given` appears not to be set, and `given` otherwise.
//
// For numeric types 0 is unset.
// For strings, maps, arrays, and slices, len() = 0 is considered unset.
// For bool, false is unset.
// Structs are never considered unset.
//
// For everything else, including pointers, a nil value is unset.
func dfault(d any, given ...any) any {
	if empty(given) || empty(given[0]) {
		return d
	}
	return given[0]
}

// empty returns true if the given value has the zero value for its type.
func empty(given any) bool {
	g := reflect.ValueOf(given)
	if !g.IsValid() {
		return true
	}

	// Basically adapted from text/template.isTrue
	switch g.Kind() {
	default:
		return g.IsNil()
	case reflect.Array, reflect.Slice, reflect.Map, reflect.String:
		return g.Len() == 0
	case reflect.Bool:
		return !g.Bool()
	case reflect.Complex64, reflect.Complex128:
		return g.Complex() == 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return g.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return g.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return g.Float() == 0
	case reflect.Struct:
		return false
	}
}
