package optionshelp

import (
	"fmt"
	"reflect"
	"strings"
	"text/template"

	"github.com/charmbracelet/x/ansi"

	"github.com/act3-ai/go-common/pkg/md"
	"github.com/act3-ai/go-common/pkg/options"
)

type templateScope struct {
	groupsByKey map[string]*options.Group
}

func newTemplateScope(groups ...*options.Group) *templateScope {
	scope := &templateScope{
		groupsByKey: map[string]*options.Group{},
	}
	for _, group := range groups {
		scope.groupsByKey[group.Key] = group
	}
	return scope
}

func (scope *templateScope) templateFuncs() template.FuncMap {
	return template.FuncMap{
		"default":     dfault,
		"groupTable":  scope.GroupTable,
		"optionTable": scope.OptionTable,
	}
}

func (scope *templateScope) mustGetGroup(key string) *options.Group {
	group, ok := scope.groupsByKey[key]
	if !ok {
		panic(fmt.Errorf("no group with key %q", key))
	}
	return group
}

/*
| Option | Description |
| ------ | ----------- |
{{- range .Options }}
| {{ .MarkdownLink }} | {{ .ShortDescription }} |
{{- end }}
*/
func (scope *templateScope) GroupTable(g *options.Group) string {
	header := []string{"Option", "Description"}
	rows := [][]string{}

	for _, o := range g.Options {
		header := o.Header()
		switch {
		case o.TargetGroupName != "":
			group := scope.mustGetGroup(o.TargetGroupName)
			rows = append(rows, []string{
				md.Link(header, md.HeaderLinkTarget(group.Title)), o.ShortDescription(),
			})
		default:
			rows = append(rows, []string{
				md.Link(header, md.HeaderLinkTarget(header)), o.ShortDescription(),
			})
		}
	}

	return writeTable(header, rows)
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
func (scope *templateScope) OptionTable(o *options.Option) string {
	header := []string{"Name", "Value"}
	rows := [][]string{}

	switch o.Type {
	case options.Object:
		rows = append(rows, []string{
			"type", scope.formattedValueType(o),
		})
	case options.List:
		rows = append(rows, [][]string{
			{"type", "list"},
			{"values", scope.formattedValueType(o)},
		}...)
	case options.StringMap:
		rows = append(rows, [][]string{
			{"type", "object"},
			{"keys", "string"},
			{"values", scope.formattedValueType(o)},
		}...)
	default:
		rows = append(rows, []string{
			"type", string(o.Type),
		})
	}
	if o.Default != "" {
		rows = append(rows, []string{
			"default", md.Code(o.Default),
		})
	}
	if o.JSON != "" {
		rows = append(rows, []string{
			"json/yaml", md.Code(o.JSON),
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
		rows = append(rows, []string{
			"cli", fflag,
		})
	}
	if o.Env != "" {
		rows = append(rows, []string{
			"env", md.Code(o.Env),
		})
	}

	return writeTable(header, rows)
}

// func (scope *templateScope) formattedType(o *options.Option) string {
// 	switch o.Type {
// 	case options.Object:
// 		if o.TargetGroupName != "" {
// 			return scope.targetLink(o.TargetGroupName)
// 		}
// 		return "object"
// 	case options.List:
// 		return fmt.Sprintf("list(values: %s)", scope.formattedValueType(o))
// 	case options.StringMap:
// 		return fmt.Sprintf("object(keys: string, values: %s)", scope.formattedValueType(o))
// 	default:
// 		return string(o.Type)
// 	}
// }

func (scope *templateScope) formattedValueType(o *options.Option) string {
	switch {
	case o.ValueType != "":
		return string(o.ValueType)
	case o.TargetGroupName != "":
		return scope.targetLink(o.TargetGroupName)
	default:
		return "any"
	}
}

func (scope *templateScope) targetLink(groupKey string) string {
	group := scope.mustGetGroup(groupKey)
	return md.Link(group.Title, md.HeaderLinkTarget(group.Title))
}

// writeTable writes a markdown table with equal length columns
func writeTable(header []string, rows [][]string) string {
	// Get maximum width of each column
	colMaxLens := make([]int, len(header))
	for _, row := range rows {
		for col, cell := range row {
			cellLen := ansi.StringWidth(cell) // ansi-aware string width
			if cellLen > colMaxLens[col] {
				colMaxLens[col] = cellLen
			}
		}
	}

	fmtStrings := make([]string, len(header))
	for col, maxLen := range colMaxLens {
		fmtStrings[col] = fmt.Sprintf("%%-%ds", maxLen)
	}

	w := &strings.Builder{}

	writeRow := func(row []string) {
		for col, cell := range row {
			_, _ = w.WriteString("| " + fmt.Sprintf(fmtStrings[col], cell) + " ")
		}
		_, _ = w.WriteString("|\n")
	}

	// Write header row
	writeRow(header)

	// Write separator row
	for col := range header {
		_, _ = fmt.Fprintf(w, "| %s ", strings.Repeat("-", colMaxLens[col]))
	}
	_, _ = w.WriteString("|\n")

	// Write separator row
	for _, row := range rows {
		writeRow(row)
	}

	return w.String()
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
