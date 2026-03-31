// Package schemamd renders Markdown documentation from JSON Schemas.
package schemamd

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"

	"github.com/act3-ai/go-common/pkg/md"
	"github.com/act3-ai/go-common/pkg/schemautil"
)

// Renderer renders Markdown documentation from a JSON Schema.
type Renderer struct {
	refFormatter func(ref string) string
}

// NewRenderer creates a Renderer with the default configuration.
func NewRenderer() *Renderer {
	return &Renderer{
		refFormatter: defaultRefFormatter,
	}
}

func defaultRefFormatter(ref string) string {
	base := path.Base(ref)
	return md.Link(md.Code(base), md.HeaderLinkTarget(base))
}

// RenderMarkdown renders Markdown documentation from a JSON Schema.
func (r *Renderer) RenderMarkdown(schema *jsonschema.Schema) string {
	return r.schemaDocumentation(0, schema)
}

//nolint:gocognit
func (r *Renderer) schemaDocumentation(n int, schema *jsonschema.Schema) string {
	pad := strings.Repeat(" ", n)
	out := &strings.Builder{}

	if schema.Description != "" {
		if strings.Count(schema.Description, "\n") > 0 {
			fmt.Fprint(out,
				pad+"- Description:\n\n"+
					mdIndent(n+2, schema.Description)+"\n\n")
		} else {
			fmt.Fprint(out,
				pad+"- Description: "+schema.Description+"\n")
		}
	}

	r.writeSchemaType(n, out, schema)

	if numProps := len(schema.Properties); numProps > 0 {
		fmt.Fprintln(out,
			detailsBullet(n,
				"Properties",
				r.propertiesDocumentation(0, schema),
				numProps < 10, // Collapse schemas with more than 10 properties
			))
	}

	if schema.AdditionalProperties != nil &&
		!schemautil.IsTrueSchema(schema.AdditionalProperties) &&
		!schemautil.IsFalseSchema(schema.AdditionalProperties) {
		fmt.Fprint(out, pad+"- Additional properties:\n")
		// Write key name if set in extension
		if keyName := schemautil.GetXAdditionalPropertiesName(schema.AdditionalProperties); keyName != "" {
			fmt.Fprint(out, pad+"  - Key: `"+keyName+"`\n")
		}
		fmt.Fprint(out, r.schemaDocumentation(n+2, schema.AdditionalProperties))
	}

	if len(schema.AllOf) > 1 {
		fmt.Fprint(out, pad+"- All of:\n")
		for _, subschema := range schema.AllOf {
			fmt.Fprint(out, r.schemaDocumentation(n+2, subschema))
		}
	}
	if len(schema.AnyOf) > 1 {
		fmt.Fprint(out, pad+"- Any of:\n")
		for _, subschema := range schema.AnyOf {
			fmt.Fprint(out, r.schemaDocumentation(n+2, subschema))
		}
	}
	if len(schema.OneOf) > 1 {
		fmt.Fprint(out, pad+"- One of:\n")
		for _, subschema := range schema.OneOf {
			fmt.Fprint(out, r.schemaDocumentation(n+2, subschema))
		}
	}

	if len(schema.Enum) > 0 {
		fmt.Fprint(out, pad+"- Enum:\n")
		for _, enumValue := range schema.Enum {
			enumValueJSON, _ := mustToPrettyJSON(enumValue)
			fmt.Fprint(out, pad+"  - `"+enumValueJSON+"`\n")
		}
	}

	if len(schema.Examples) > 0 {
		fmt.Fprint(out, pad+"- Examples:\n")
		for _, exampleValue := range schema.Examples {
			exampleValueJSON, _ := mustToPrettyJSON(exampleValue)
			lines := strings.Count(exampleValueJSON, "\n") + 1
			switch {
			case lines == 1:
				fmt.Fprintf(out, pad+"  - %s\n", md.Code(exampleValueJSON))
			case lines < 10:
				fmt.Fprintln(out, pad+"  - \n"+
					mdIndent(n+4, md.CodeBlock("json", exampleValueJSON)))
			default:
				fmt.Fprintln(out, pad+"  - "+
					mdIndentNext(n+4, md.Details("Example JSON Document", md.CodeBlock("json", exampleValueJSON)+"\n", false)))
			}
		}
	}

	return out.String()
}

func (r *Renderer) writeSchemaType(n int, out *strings.Builder, schema *jsonschema.Schema) {
	pad := strings.Repeat(" ", n)

	typeMD := r.typeToMarkdown(schema)
	if typeMD != "" {
		fmt.Fprintf(out, pad+"- Type: %s\n", typeMD)
	}

	if schema.Pattern != "" {
		fmt.Fprintf(out, pad+"- Pattern: %s\n", md.Code(schema.Pattern))
	}

	if schema.Const != nil {
		constValue, _ := json.Marshal(schema.Const) //nolint:errchkjson // ignore errors
		fmt.Fprintf(out, pad+"- Value: %s\n", md.Code(string(constValue)))
	}

	// Write length of static size array
	if schema.Type == schemautil.TypeArray &&
		schema.MinItems != nil && schema.MaxItems != nil &&
		*schema.MinItems == *schema.MaxItems {
		fmt.Fprintf(out, pad+"- Items: `%d`\n", *schema.MinItems)
	}
}

func (r *Renderer) typeToMarkdown(schema *jsonschema.Schema) string {
	switch {
	// Format basic type
	case schema.Type != "":
		switch schema.Type {
		case schemautil.TypeString:
			return r.typeStringToMarkdown(schema)
		case schemautil.TypeNumber:
			return r.typeNumberToMarkdown(schema)
		case schemautil.TypeInteger:
			return r.typeIntegerToMarkdown(schema)
		case schemautil.TypeBoolean:
			return md.Code(schemautil.TypeBoolean)
		case schemautil.TypeObject:
			return md.Code(schemautil.TypeObject)
		case schemautil.TypeArray:
			return "list of " + r.typeToMarkdown(schema.Items)
		default:
			return md.Code(schema.Type)
		}
	// Link to referenced type
	case schema.Ref != "":
		return r.refFormatter(schema.Ref)
	// Recurse into single AllOf
	case len(schema.AllOf) == 1 &&
		schema.AllOf[0].Ref != "":
		return r.typeToMarkdown(schema.AllOf[0])
	case len(schema.OneOf) == 2 &&
		schema.OneOf[1].Type == schemautil.TypeNull:
		return r.typeToMarkdown(schema.OneOf[0]) + " or `null`"
	case len(schema.AnyOf) == 2 &&
		schema.AnyOf[1].Type == schemautil.TypeNull:
		return r.typeToMarkdown(schema.AnyOf[0]) + " or `null`"
	default:
		// slog.Warn("cannot determine type of schema", slog.Any("description", schema.Description))
		return "any"
	}
}

func (r *Renderer) typeStringToMarkdown(schema *jsonschema.Schema) string {
	switch schema.Format {
	case schemautil.FormatDate:
		return md.Code(schemautil.FormatDate)
	case schemautil.FormatDateTime:
		return md.Code(schemautil.FormatDateTime)
	case schemautil.FormatTime:
		return md.Code(schemautil.FormatTime)
	default:
		return md.Code(schemautil.TypeString)
	}
}

func (r *Renderer) typeNumberToMarkdown(schema *jsonschema.Schema) string {
	switch schema.Format {
	case schemautil.FormatDouble:
		return md.Code(schemautil.FormatDouble)
	case schemautil.FormatFloat:
		return md.Code(schemautil.FormatFloat)
	default:
		return md.Code(schema.Type)
	}
}

func (r *Renderer) typeIntegerToMarkdown(schema *jsonschema.Schema) string {
	switch schema.Format {
	case schemautil.FormatInt64:
		return md.Code(schemautil.FormatInt64)
	case schemautil.FormatInt32:
		return md.Code(schemautil.FormatInt32)
	default:
		return md.Code(schema.Type)
	}
}

func (r *Renderer) propertiesDocumentation(n int, schema *jsonschema.Schema) string {
	out := &strings.Builder{}
	pad := strings.Repeat(" ", n)

	// Use to check if property is required
	required := toLookupMap(schema.Required)

	for propName, prop := range schemautil.OrderedProperties(schema) {
		propNameFmt := "`" + propName + "`"
		if required[propName] {
			propNameFmt += " **REQUIRED**"
		}
		// fmt.Fprint(out,
		// 	pad+"- "+propNameFmt+"\n\n"+
		// 		indent(n+2, md.Details(
		// 			"Property schema",
		// 			schemaDocumentation(n+2, prop, new(required[propName])),
		// 			true,
		// 		))+"\n",
		// )
		fmt.Fprintln(out, pad+"- "+propNameFmt)
		fmt.Fprint(out, r.schemaDocumentation(n+2, prop))
	}

	return out.String()
}

func toLookupMap[K comparable](slice []K) map[K]bool {
	out := make(map[K]bool, len(slice))
	for _, key := range slice {
		out[key] = true
	}
	return out
}

// func schemaValidationDocumentation(n int, schema *jsonschema.Schema) string {
// 	out := &strings.Builder{}
// 	pad := strings.Repeat(" ", n)
//
// 	// Min/Max Length
// 	writeRangeInclusive(out, pad, "Length", "n", schema.MinLength, schema.MaxLength)
// 	// Min/Max Items
// 	writeRangeInclusive(out, pad, "Array length", "n", schema.MinItems, schema.MaxItems)
// 	// Min/max
// 	writeRangeNumber(out, pad, "Value", "value",
// 		schema.Minimum, schema.ExclusiveMinimum,
// 		schema.Maximum, schema.ExclusiveMaximum)
//
// 	return out.String()
// }

func detailsBullet(n int, summary, body string, open bool) string {
	pad := strings.Repeat(" ", n)
	return pad + "- " + mdIndentNext(n+2, md.Details(summary, body, open))
}

func mustNumberFloat64(n json.Number) float64 {
	nInt, err := n.Int64()
	if err == nil {
		return float64(nInt)
	}
	nFloat, _ := n.Float64()
	return nFloat
}

// func writeRangeInclusive[T cmp.Ordered](out *strings.Builder, pad string, key, placeholder string, minValue, maxValue *T) {
// 	writeRange(out, pad, key, placeholder, minValue, nil, maxValue, nil)
// }

// func writeRange[T cmp.Ordered](out *strings.Builder, pad string, key, placeholder string,
// 	minValue, exclMinValue,
// 	maxValue, exclMaxValue *T,
// ) {
// 	bounds := ToBoundsStringValue(placeholder, minValue, exclMinValue, maxValue, exclMaxValue)
// 	if bounds == "" {
// 		return
// 	}
// 	fmt.Fprintf(out, pad+"- %s: `%s`\n", key, bounds)
// }

// func writeRangeNumber(out *strings.Builder, pad string, key, placeholder string,
// 	minNumber, exclMinNumber,
// 	maxNumber, exclMaxNumber json.Number,
// ) {
// 	bounds := ToBoundsStringNumber(placeholder, minNumber, exclMinNumber, maxNumber, exclMaxNumber)
// 	if bounds == "" {
// 		return
// 	}
// 	fmt.Fprintf(out, pad+"- %s: `%s`\n", key, bounds)
// }

// If no bounds are set, the empty string is returned.
func ToBoundsStringValue[T cmp.Ordered](placeholder string,
	minValue, exclMinValue,
	maxValue, exclMaxValue *T) string {
	numberIfNotNil := func(in *T) json.Number {
		if in == nil {
			return ""
		}
		return json.Number(fmt.Sprint(*in))
	}
	return ToBoundsStringNumber(placeholder,
		numberIfNotNil(minValue),
		numberIfNotNil(exclMinValue),
		numberIfNotNil(maxValue),
		numberIfNotNil(exclMaxValue),
	)
}

// If no bounds are set, the empty string is returned.
func ToBoundsStringNumber(
	placeholder string,
	minNumber, exclMinNumber,
	maxNumber, exclMaxNumber json.Number,
) string {
	// If inclusive min and max are set to same number, write equality expression and return
	if minNumber != "" && maxNumber != "" && mustNumberFloat64(minNumber) == mustNumberFloat64(maxNumber) {
		return placeholder + " == " + minNumber.String()
	}

	var minBound string
	switch {
	case minNumber != "":
		minBound = minNumber.String() + " <= "
	case exclMinNumber != "":
		minBound = exclMinNumber.String() + " < "
	}

	var maxBound string
	switch {
	case maxNumber != "":
		maxBound = " <= " + maxNumber.String()
	case exclMaxNumber != "":
		maxBound = " < " + exclMaxNumber.String()
	}

	if minBound == "" && maxBound == "" {
		return ""
	}

	return minBound + placeholder + maxBound
}

// nolint:unused,nolintlint
func mustToPrettyJSON(v any) (string, error) {
	buf := new(bytes.Buffer)
	e := json.NewEncoder(buf)
	e.SetEscapeHTML(false)
	e.SetIndent("", "  ")
	err := e.Encode(v)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(buf.String()), nil
}

// mdIndentNext indents a string with spaces, starting with the second line.
func mdIndentNext(spaces int, v string) string {
	// Create the indent
	pad := strings.Repeat(" ", spaces)
	// Indent the string
	indented := strings.ReplaceAll(v, "\n", "\n"+pad)
	// Clear whitespace lines
	indented = clearWhitespaceLines(indented)
	return indented
}

// mdIndent indents a string with spaces.
func mdIndent(spaces int, v string) string {
	// Create the indent
	pad := strings.Repeat(" ", spaces)
	// Indent the string
	indented := strings.ReplaceAll(v, "\n", "\n"+pad)
	// Add leading indent
	indented = pad + indented
	// Clear whitespace lines
	indented = clearWhitespaceLines(indented)
	return indented
}

// clearWhitespaceLines returns the string with all whitespace-only lines cleared (newlines are preserved).
func clearWhitespaceLines(s string) string {
	w := &strings.Builder{}
	for line := range strings.Lines(s) {
		// If line is only whitespace, set to the empty string
		if strings.TrimSpace(line) == "" {
			if strings.HasSuffix(line, "\n") {
				line = "\n" // preserve trailing newline
			} else {
				line = "" // do not introduce trailing newline
			}
		}
		fmt.Fprint(w, line)
	}
	return w.String()
}
