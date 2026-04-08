package schemagen

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/token"
	"slices"
	"strconv"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"

	"github.com/act3-ai/go-common/pkg/astutil"
	"github.com/act3-ai/go-common/pkg/schemautil"
)

func (gen *Generator) applyStructFieldDirectives(props *objectProperties, propName string, schema *jsonschema.Schema, comments *ast.CommentGroup) error {
	for _, dir := range astutil.AllDirectivesForTool(gen.DirectiveTool, comments) {
		if err := gen.applyStructFieldDirective(props, propName, schema, dir); err != nil {
			return err
		}
	}
	return nil
}

func (gen *Generator) applyStructFieldDirective(props *objectProperties, propName string, schema *jsonschema.Schema, dir ast.Directive) error {
	switch dir.Name {
	case "required":
		required := true
		if dir.Args != "" {
			args, err := dir.ParseArgs()
			if err != nil {
				return fmt.Errorf("%s: %w", gen.PackageInfo.Fset.Position(dir.ArgsPos), err)
			}
			if err := validateArgLength(args, 1); err != nil {
				return fmt.Errorf("%s: %w", gen.PackageInfo.Fset.Position(dir.ArgsPos), err)
			}
			required, err = strconv.ParseBool(args[0].Arg)
			if err != nil {
				return fmt.Errorf("%s: %w", gen.PackageInfo.Fset.Position(args[0].Pos), err)
			}
		}
		if required {
			if !slices.Contains(props.Required, propName) {
				props.Required = append(props.Required, propName)
			}
		} else {
			props.Required = slices.DeleteFunc(props.Required, func(v string) bool {
				return v == propName
			})
		}
		return nil
	default:
		return gen.applySchemaDirective(schema, dir)
	}
}

func (gen *Generator) applySchemaDirectives(schema *jsonschema.Schema, comments *ast.CommentGroup) error {
	for _, dir := range astutil.AllDirectivesForTool(gen.DirectiveTool, comments) {
		if err := gen.applySchemaDirective(schema, dir); err != nil {
			return err
		}
	}
	return nil
}

func (gen *Generator) applySchemaDirective(schema *jsonschema.Schema, dir ast.Directive) error {
	switch dir.Name {
	case "set":
		return gen.applySetDirective(schema, dir)
	default:
		return fmt.Errorf("%s: unsupported name %q in directive %s", gen.PackageInfo.Fset.Position(dir.Slash), dir.Name, directiveString(dir))
	}
}

func (gen *Generator) applySetDirective(schema *jsonschema.Schema, dir ast.Directive) error {
	args, err := dir.ParseArgs()
	if err != nil {
		return fmt.Errorf("%s: %w", gen.PackageInfo.Fset.Position(dir.ArgsPos), err)
	}
	if err := validateArgLength(args, 2); err != nil {
		return fmt.Errorf("%s: %w", gen.PackageInfo.Fset.Position(dir.ArgsPos), err)
	}

	propertyName := args[0].Arg
	arg := args[1]

	// Allow user to set any non-subschema field on the schema
	switch propertyName {
	case "$id":
		schema.ID = arg.Arg
		return nil
	case "$schema":
		schema.Schema = arg.Arg
		return nil
	case "$ref":
		schema.Ref = arg.Arg
		return nil
	case "$comment":
		schema.Comment = arg.Arg
		return nil
	case "$anchor":
		schema.Anchor = arg.Arg
		return nil
	case "$dynamicAnchor":
		schema.DynamicAnchor = arg.Arg
		return nil
	case "$dynamicRef":
		schema.DynamicRef = arg.Arg
		return nil
	case "$vocabulary":
		return setJSON(gen.PackageInfo.Fset, arg, &schema.Vocabulary)
	case "title":
		schema.Title = arg.Arg
		return nil
	case "description":
		schema.Description = arg.Arg
		return nil
	case "default":
		return setJSONRawMessage(gen.PackageInfo.Fset, arg, &schema.Default)
	case "deprecated":
		return setBoolean(gen.PackageInfo.Fset, arg, &schema.Deprecated)
	case "readOnly":
		return setBoolean(gen.PackageInfo.Fset, arg, &schema.ReadOnly)
	case "writeOnly":
		return setBoolean(gen.PackageInfo.Fset, arg, &schema.WriteOnly)
	case "examples":
		return setJSON(gen.PackageInfo.Fset, arg, &schema.Examples)
	case "type":
		schema.Type = arg.Arg
		return nil
	case "enum":
		return setJSON(gen.PackageInfo.Fset, arg, &schema.Enum)
	case "const":
		return setJSON(gen.PackageInfo.Fset, arg, &schema.Const)
	case "multipleOf":
		return setFloat64Pointer(gen.PackageInfo.Fset, arg, &schema.MultipleOf)
	case "minimum":
		return setFloat64Pointer(gen.PackageInfo.Fset, arg, &schema.Minimum)
	case "maximum":
		return setFloat64Pointer(gen.PackageInfo.Fset, arg, &schema.Maximum)
	case "exclusiveMinimum":
		return setFloat64Pointer(gen.PackageInfo.Fset, arg, &schema.ExclusiveMinimum)
	case "exclusiveMaximum":
		return setFloat64Pointer(gen.PackageInfo.Fset, arg, &schema.ExclusiveMaximum)
	case "minLength":
		return setIntPointer(gen.PackageInfo.Fset, arg, &schema.MinLength)
	case "maxLength":
		return setIntPointer(gen.PackageInfo.Fset, arg, &schema.MaxLength)
	case "pattern":
		schema.Pattern = arg.Arg
		return nil
	case "minItems":
		return setIntPointer(gen.PackageInfo.Fset, arg, &schema.MinItems)
	case "maxItems":
		return setIntPointer(gen.PackageInfo.Fset, arg, &schema.MaxItems)
	case "uniqueItems":
		return setBoolean(gen.PackageInfo.Fset, arg, &schema.UniqueItems)
	case "minContains":
		return setIntPointer(gen.PackageInfo.Fset, arg, &schema.MinContains)
	case "maxContains":
		return setIntPointer(gen.PackageInfo.Fset, arg, &schema.MaxContains)
	case "minProperties":
		return setIntPointer(gen.PackageInfo.Fset, arg, &schema.MinProperties)
	case "maxProperties":
		return setIntPointer(gen.PackageInfo.Fset, arg, &schema.MaxProperties)
	case "required":
		return setJSON(gen.PackageInfo.Fset, arg, &schema.Required)
	case "dependentRequired":
		return setJSON(gen.PackageInfo.Fset, arg, &schema.DependentRequired)
	case "additionalProperties":
		var value *bool
		if err := setBooleanPointer(gen.PackageInfo.Fset, args[1], &value); err != nil {
			return err
		}
		switch {
		case value == nil:
			schema.AdditionalProperties = nil
		case *value:
			schema.AdditionalProperties = schemautil.TrueSchema()
		default:
			schema.AdditionalProperties = schemautil.FalseSchema()
		}
		return nil
	case "contentEncoding":
		schema.ContentEncoding = arg.Arg
		return nil
	case "contentMediaType":
		schema.ContentMediaType = arg.Arg
		return nil
	case "format":
		schema.Format = arg.Arg
		return nil
	default:
		// Check if property name is an extra
		if strings.HasPrefix(propertyName, "x-") {
			var value any
			if err := setJSON(gen.PackageInfo.Fset, arg, &value); err != nil {
				return err
			}
			if schema.Extra == nil {
				schema.Extra = make(map[string]any, 1)
			}
			schema.Extra[propertyName] = value
			return nil
		}

		// Return an error for all other property names
		return fmt.Errorf("%s: unsupported property %q in directive %s", gen.PackageInfo.Fset.Position(args[0].Pos), propertyName, directiveString(dir))
	}
}

func setString(args []ast.DirectiveArg, v *string) error {
	*v = args[0].Arg
	return nil
}

func setJSON[T any](fset *token.FileSet, arg ast.DirectiveArg, v *T) error {
	var value T
	if err := json.Unmarshal([]byte(arg.Arg), &value); err != nil {
		return fmt.Errorf("%s: parsing argument as JSON: %w", fset.Position(arg.Pos), err)
	}
	*v = value
	return nil
}

func setJSONRawMessage(fset *token.FileSet, arg ast.DirectiveArg, v *json.RawMessage) error {
	var value any
	if err := json.Unmarshal([]byte(arg.Arg), &value); err != nil {
		return fmt.Errorf("%s: parsing argument as JSON: %w", fset.Position(arg.Pos), err)
	}
	*v = json.RawMessage(arg.Arg)
	return nil
}

func setBoolean(fset *token.FileSet, arg ast.DirectiveArg, v *bool) error {
	value, err := strconv.ParseBool(arg.Arg)
	if err != nil {
		return fmt.Errorf("%s: value must be a boolean: %w", fset.Position(arg.Pos), err)
	}
	*v = value
	return nil
}

func setBooleanPointer(fset *token.FileSet, arg ast.DirectiveArg, v **bool) error {
	if arg.Arg == "null" {
		*v = nil
		return nil
	}
	value, err := strconv.ParseBool(arg.Arg)
	if err != nil {
		return fmt.Errorf("%s: value must be a boolean or null: %w", fset.Position(arg.Pos), err)
	}
	**v = value
	return nil
}

func setIntPointer(fset *token.FileSet, arg ast.DirectiveArg, v **int) error {
	if arg.Arg == "null" {
		*v = nil
		return nil
	}
	value, err := strconv.Atoi(arg.Arg)
	if err != nil {
		return fmt.Errorf("%s: value must be an int or null: %w", fset.Position(arg.Pos), err)
	}
	*v = &value
	return nil
}

func setFloat64Pointer(fset *token.FileSet, arg ast.DirectiveArg, v **float64) error {
	if arg.Arg == "null" {
		*v = nil
		return nil
	}
	value, err := strconv.ParseFloat(arg.Arg, 64)
	if err != nil {
		return fmt.Errorf("%s: value must be a float64 or null: %w", fset.Position(arg.Pos), err)
	}
	*v = &value
	return nil
}

func validateArgLength[T any](args []T, want int) error {
	if len(args) != want {
		return fmt.Errorf("incorrect number of arguments: want %d, got %d", want, len(args))
	}
	return nil
}

func directiveString(dir ast.Directive) string {
	return fmt.Sprintf("//%s:%s %s", dir.Tool, dir.Name, dir.Args)
}
