package schemagen

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/token"
	"log/slog"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/google/jsonschema-go/jsonschema"

	"github.com/act3-ai/go-common/pkg/astutil"
	"github.com/act3-ai/go-common/pkg/schemautil"
)

func NewGenerator() *Generator {
	return &Generator{
		standardSchemas: standardSchemas,
		results:         map[reflect.Type]result{},
		directiveTool:   "jsonschema",
	}
}

type Generator struct {
	info            *astutil.PackageInfo
	standardSchemas map[reflect.Type]func() *jsonschema.Schema
	results         map[reflect.Type]result

	// directive tool name.
	directiveTool string

	// extensions to provide a schema for types,
	// first provider to return a non-nil schema
	// will replace the regular schema generation.
	schemaProviders []func(t reflect.Type) *jsonschema.Schema

	// extensions to modify the generated schema for types.
	// will be called after the regular schema generation
	// and after the SchemaExtender interface implementation.
	schemaExtenders map[reflect.Type][]func(schema *jsonschema.Schema)
}

func (gen *Generator) WithPackageInfo(info *astutil.PackageInfo) *Generator {
	gen.info = info
	return gen
}

func (gen *Generator) Definitions() map[string]*jsonschema.Schema {
	defs := make(map[string]*jsonschema.Schema, len(gen.results))
	for t, r := range gen.results {
		schemaName := t.PkgPath() + "." + t.Name()
		defs[schemaName] = r.Schema
		if r.Schema == nil && r.Err != nil {
			defs[schemaName] = &jsonschema.Schema{
				Comment: "Error: " + r.Err.Error(),
			}
		}
	}
	return defs
}

type result struct {
	Schema *jsonschema.Schema
	Err    error
}

func (gen *Generator) getTypeComment(t reflect.Type) *ast.CommentGroup {
	if gen.info == nil || t.PkgPath() == "" {
		return nil
	}
	comment, err := gen.info.TypeComment(t.PkgPath(), t.Name())
	if err != nil {
		slog.Error("type comment lookup",
			slog.String("pkgPath", t.PkgPath()),
			slog.String("typeName", t.Name()),
		)
		// return nil, err
	}
	return comment
}

func (gen *Generator) getFieldComment(t reflect.Type, field reflect.StructField) *ast.CommentGroup {
	if gen.info == nil || t.PkgPath() == "" {
		return nil
	}
	comment, err := gen.info.FieldComment(t.PkgPath(), t.Name(), field.Name)
	if err != nil {
		slog.Error("field comment lookup",
			slog.String("pkgPath", t.PkgPath()),
			slog.String("typeName", t.Name()),
			slog.String("fieldName", field.Name),
		)
		// return err
	}
	return comment
}

func GenerateSchemaFor[T any](gen *Generator) (*jsonschema.Schema, error) {
	return gen.GenerateSchemaForType(reflect.TypeFor[T]())
}

func (gen *Generator) GenerateSchemaForType(t reflect.Type) (*jsonschema.Schema, error) {
	// Return cached result
	if r, ok := gen.results[t]; ok {
		return r.Schema, r.Err
	}

	// Generate new schema
	schema, err := generateSchemaForType(gen, t)

	// Cache result if it is a named schema
	if t.PkgPath() != "" && t.Name() != "" {
		gen.results[t] = result{Schema: schema, Err: err}
	}

	return schema, err
}

func generateSchemaForType(gen *Generator, t reflect.Type) (*jsonschema.Schema, error) {
	if t == nil {
		return nil, nil
	}

	var (
		schema *jsonschema.Schema
		err    error
	)

	// Get the comment for this type
	comment := gen.getTypeComment(t)

	// If the type provides a schema, use that schema.
	if t.Implements(typeSchemaProvider) {
		v, _ := reflect.TypeAssert[SchemaProvider](reflect.New(t).Elem())
		schema = v.JSONSchema()
	} else {
		// Generate the schema from the type
		schema, err = generateSchema(gen, t)
		if err != nil {
			return nil, err
		}

		// Add description from comment
		schema.Description = formatCommentAsDescription(comment)
	}

	// If the type defines a schema extension method, call the method.
	if t.Implements(typeSchemaExtender) {
		v, _ := reflect.TypeAssert[SchemaExtender](reflect.New(t).Elem())
		v.ExtendJSONSchema(schema)
	}

	return schema, nil
}

func generateSchemaForMap(gen *Generator, t reflect.Type) (*jsonschema.Schema, error) {
	var err error
	schema := &jsonschema.Schema{
		Type: schemautil.TypeObject,
	}

	// Generate schema for map keys
	schema.PropertyNames, err = gen.GenerateSchemaForType(t.Key())
	if err != nil {
		return nil, err
	}

	// Generate schema for map values
	schema.AdditionalProperties, err = gen.GenerateSchemaForType(t.Elem())
	if err != nil {
		return nil, err
	}

	return schema, nil
}

func generateSchemaForStruct(gen *Generator, t reflect.Type) (*jsonschema.Schema, error) {
	// Generate schema for each field
	props, err := generateObjectPropertiesForStruct(gen, t)
	if err != nil {
		return nil, err
	}

	schema := &jsonschema.Schema{
		Type:          schemautil.TypeObject,
		Required:      props.Required,
		Properties:    props.Properties,
		PropertyOrder: props.PropertyOrder,
	}

	switch {
	case len(props.Embedded) > 0 && len(schema.Properties) > 0:
		// Embedded schemas and object schema
		return &jsonschema.Schema{
			AllOf: append(props.Embedded, schema),
		}, nil
	case len(props.Embedded) > 0:
		// Only embedded schemas
		return &jsonschema.Schema{
			AllOf: props.Embedded,
		}, nil
	default:
		// Regular object schema
		return schema, nil
	}
}

type objectProperties struct {
	Required      []string
	Properties    map[string]*jsonschema.Schema
	PropertyOrder []string
	Embedded      []*jsonschema.Schema
}

func generateObjectPropertiesForStruct(gen *Generator, t reflect.Type) (*objectProperties, error) {
	props := &objectProperties{
		Properties:    make(map[string]*jsonschema.Schema, t.NumField()),
		PropertyOrder: make([]string, 0, t.NumField()),
	}
	// Generate schema for each field
	for field := range t.Fields() {
		if err := addSchemaForStructField(gen, t, props, field); err != nil {
			return nil, fmt.Errorf("adding schema for struct field %s: %w", field.Name, err)
		}
	}
	return props, nil
}

// addSchemaForStructField generates a schema for a struct field and adds it to the object schema.
func addSchemaForStructField(gen *Generator, t reflect.Type, props *objectProperties, field reflect.StructField) error {
	// Property name
	propName := field.Name

	// Parse JSON struct tag information
	tagInfo, ok := parseJSONStructTag(field.Tag)
	if ok {
		propName = tagInfo.Name
	}

	// Skip ignored fields
	if tagInfo.Ignored {
		return nil
	}

	// Get the comment for this field
	comment := gen.getFieldComment(t, field)

	// Create property schema
	schema, err := generateSchemaOrReference(gen, field.Type)
	if err != nil {
		return err
	}

	// Add description
	schema.Description = formatCommentAsDescription(comment)

	if field.Anonymous {
		// Apply comment directives to the schema
		if err := gen.applySchemaDirectives(schema, comment); err != nil {
			return err
		}

		// Add embedded fields to list
		props.Embedded = append(props.Embedded, schema)
	} else {
		// Add to required properties if neither omitempty/omitzero are set
		if !(tagInfo.Omitempty || tagInfo.Omitzero) {
			props.Required = append(props.Required, propName)
		}

		// Add to property order
		props.PropertyOrder = append(props.PropertyOrder, propName)

		// Add to properties
		props.Properties[propName] = schema

		// Apply comment directives to the schema
		if err := gen.applyStructFieldDirectives(props, propName, schema, comment); err != nil {
			return err
		}
	}

	return nil
}

// generateSchemaOrReference generates a schema for basic types or a reference to named types.
func generateSchemaOrReference(gen *Generator, t reflect.Type) (*jsonschema.Schema, error) {
	schema, err := gen.GenerateSchemaForType(t)
	if err != nil {
		return nil, err
	}

	// If type has PkgPath and Name, return reference to schema
	if t.PkgPath() != "" && t.Name() != "" {
		schemaName := t.PkgPath() + "." + t.Name()
		return &jsonschema.Schema{
			Ref: "#/$defs/" + schemaName,
		}, nil
	}

	// Return the schema inline
	return schema, nil
}

// Standard types.
var (
	typeString  = reflect.TypeFor[string]()
	typeBool    = reflect.TypeFor[bool]()
	typeUint    = reflect.TypeFor[uint]()
	typeUint8   = reflect.TypeFor[uint8]()
	typeUint16  = reflect.TypeFor[uint16]()
	typeUint32  = reflect.TypeFor[uint32]()
	typeUint64  = reflect.TypeFor[uint64]()
	typeInt     = reflect.TypeFor[int]()
	typeInt8    = reflect.TypeFor[int8]()
	typeInt16   = reflect.TypeFor[int16]()
	typeInt32   = reflect.TypeFor[int32]()
	typeInt64   = reflect.TypeFor[int64]()
	typeFloat32 = reflect.TypeFor[float32]()
	typeFloat64 = reflect.TypeFor[float64]()
	typeTime    = reflect.TypeFor[time.Time]()
)

func schemaString() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type: schemautil.TypeString,
	}
}

func schemaBool() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type: schemautil.TypeBoolean,
	}
}

func schemaUint() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:   schemautil.TypeInteger,
		Format: "uint32",
	}
}

func schemaUint8() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:   schemautil.TypeInteger,
		Format: "uint8",
	}
}

func schemaUint16() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:   schemautil.TypeInteger,
		Format: "uint16",
	}
}

func schemaUint32() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:   schemautil.TypeInteger,
		Format: "uint32",
	}
}

func schemaUint64() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:   schemautil.TypeInteger,
		Format: "uint64",
	}
}

func schemaInt() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:   schemautil.TypeInteger,
		Format: "int32",
	}
}

func schemaInt8() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:   schemautil.TypeInteger,
		Format: "int8",
	}
}

func schemaInt16() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:   schemautil.TypeInteger,
		Format: "int16",
	}
}
func schemaInt32() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:   schemautil.TypeInteger,
		Format: "int32",
	}
}

func schemaInt64() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:   schemautil.TypeInteger,
		Format: "int64",
	}
}

func schemaFloat32() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:   schemautil.TypeNumber,
		Format: "float",
	}
}

func schemaFloat64() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:   schemautil.TypeNumber,
		Format: "double",
	}
}

func schemaTime() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:   schemautil.TypeString,
		Format: "date-time",
	}
}

var standardSchemas = map[reflect.Type]func() *jsonschema.Schema{
	typeString:  schemaString,
	typeBool:    schemaBool,
	typeUint:    schemaUint,
	typeUint8:   schemaUint8,
	typeUint16:  schemaUint16,
	typeUint32:  schemaUint32,
	typeUint64:  schemaUint64,
	typeInt:     schemaInt,
	typeInt8:    schemaInt8,
	typeInt16:   schemaInt16,
	typeInt32:   schemaInt32,
	typeInt64:   schemaInt64,
	typeFloat32: schemaFloat32,
	typeFloat64: schemaFloat64,
	typeTime:    schemaTime,
}

func standardTypeSchema(gen *Generator, t reflect.Type) (*jsonschema.Schema, bool) {
	fn, ok := gen.standardSchemas[t]
	if !ok {
		return nil, false
	}
	return fn(), true
}

func generateSchema(gen *Generator, t reflect.Type) (schema *jsonschema.Schema, err error) {
	defer wrapf(&err, "generating schema for type %s", t)

	// Check if type is a standard type
	if schema, ok := standardTypeSchema(gen, t); ok {
		return schema, nil
	}

	// Derive the schema from the kind
	switch t.Kind() {
	case reflect.String:
		return &jsonschema.Schema{
			Type: schemautil.TypeString,
		}, nil
	case reflect.Bool:
		return &jsonschema.Schema{
			Type: schemautil.TypeBoolean,
		}, nil
	case reflect.Uint:
		return &jsonschema.Schema{
			Type: schemautil.TypeBoolean,
		}, nil
	case reflect.Pointer:
		elemSchema, err := generateSchemaOrReference(gen, t.Elem())
		if err != nil {
			return nil, err
		}
		return &jsonschema.Schema{
			OneOf: []*jsonschema.Schema{
				elemSchema,
				{Type: schemautil.TypeNull},
			},
		}, nil
	case reflect.Array:
		elemSchema, err := generateSchemaOrReference(gen, t.Elem())
		if err != nil {
			return nil, err
		}
		return &jsonschema.Schema{
			Type: schemautil.TypeArray,
			// Set constant length
			MinItems: new(t.Len()),
			MaxItems: new(t.Len()),
			Items:    elemSchema,
		}, nil
	case reflect.Slice:
		elemSchema, err := generateSchemaOrReference(gen, t.Elem())
		if err != nil {
			return nil, err
		}
		return &jsonschema.Schema{
			Type:  schemautil.TypeArray,
			Items: elemSchema,
		}, nil
	case reflect.Struct:
		return generateSchemaForStruct(gen, t)
	case reflect.Map:
		return generateSchemaForMap(gen, t)
	default:
		return nil, fmt.Errorf("unsupported type")
	}
}

type jsonTagInfo struct {
	Ignored   bool // true IFF the field was ignored using the tag `json:"-"`
	Name      string
	Omitempty bool
	Omitzero  bool

	// The "string" option signals that a field is stored as JSON inside a JSON-encoded string. It applies only to fields of string, floating point, integer, or boolean types. This extra level of encoding is sometimes used when communicating with JavaScript programs:
	String bool // Int64String int64 `json:",string"`
}

func parseJSONStructTag(tag reflect.StructTag) (jsonTagInfo, bool) {
	value, ok := tag.Lookup("json")
	if !ok {
		return jsonTagInfo{}, false
	}

	if value == "-" {
		return jsonTagInfo{Ignored: true}, true
	}

	args := strings.Split(value, ",")

	info := jsonTagInfo{
		Name: args[0], // First value is always the name
	}

	// Parse remaining options
	if len(args) > 1 {
		for _, arg := range args[1:] {
			switch arg {
			case "omitempty":
				info.Omitempty = true
			case "omitzero":
				info.Omitzero = true
			case "string":
				info.String = true
			}
		}
	}

	return info, true
}

type SchemaExtender interface {
	ExtendJSONSchema(schema *jsonschema.Schema)
}

type SchemaProvider interface {
	// JSONSchema produces the JSON Schema representation of an object.
	JSONSchema() *jsonschema.Schema
}

var (
	typeSchemaExtender = reflect.TypeFor[SchemaExtender]()
	typeSchemaProvider = reflect.TypeFor[SchemaProvider]()
)

func formatCommentAsDescription(comment *ast.CommentGroup) string {
	return strings.TrimSuffix(comment.Text(), "\n")
}

// wrapf wraps *errp with the given formatted message if *errp is not nil.
func wrapf(errp *error, format string, args ...any) {
	if *errp != nil {
		*errp = fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), *errp)
	}
}

func (gen *Generator) applyStructFieldDirectives(props *objectProperties, propName string, schema *jsonschema.Schema, comments *ast.CommentGroup) error {
	for _, dir := range astutil.AllDirectivesForTool(gen.directiveTool, comments) {
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
				return fmt.Errorf("%s: %w", gen.info.Fset.Position(dir.ArgsPos), err)
			}
			if err := validateArgLength(args, 1); err != nil {
				return fmt.Errorf("%s: %w", gen.info.Fset.Position(dir.ArgsPos), err)
			}
			required, err = strconv.ParseBool(args[0].Arg)
			if err != nil {
				return fmt.Errorf("%s: %w", gen.info.Fset.Position(args[0].Pos), err)
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
	for _, dir := range astutil.AllDirectivesForTool(gen.directiveTool, comments) {
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
		return fmt.Errorf("%s: unsupported name %q in directive %s", gen.info.Fset.Position(dir.Slash), dir.Name, directiveString(dir))
	}
}

func (gen *Generator) applySetDirective(schema *jsonschema.Schema, dir ast.Directive) error {
	args, err := dir.ParseArgs()
	if err != nil {
		return fmt.Errorf("%s: %w", gen.info.Fset.Position(dir.ArgsPos), err)
	}
	if err := validateArgLength(args, 2); err != nil {
		return fmt.Errorf("%s: %w", gen.info.Fset.Position(dir.ArgsPos), err)
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
		return setJSON(gen.info.Fset, arg, &schema.Vocabulary)
	case "title":
		schema.Title = arg.Arg
		return nil
	case "description":
		schema.Description = arg.Arg
		return nil
	case "default":
		schema.Default = json.RawMessage(arg.Arg)
		return nil
	case "deprecated":
		return setBoolean(gen.info.Fset, arg, &schema.Deprecated)
	case "readOnly":
		return setBoolean(gen.info.Fset, arg, &schema.ReadOnly)
	case "writeOnly":
		return setBoolean(gen.info.Fset, arg, &schema.WriteOnly)
	case "examples":
		return setJSON(gen.info.Fset, arg, &schema.Examples)
	case "type":
		schema.Type = arg.Arg
		return nil
	case "enum":
		return setJSON(gen.info.Fset, arg, &schema.Enum)
	case "const":
		return setJSON(gen.info.Fset, arg, &schema.Const)
	case "multipleOf":
		return setFloat64Pointer(gen.info.Fset, arg, &schema.MultipleOf)
	case "minimum":
		return setFloat64Pointer(gen.info.Fset, arg, &schema.Minimum)
	case "maximum":
		return setFloat64Pointer(gen.info.Fset, arg, &schema.Maximum)
	case "exclusiveMinimum":
		return setFloat64Pointer(gen.info.Fset, arg, &schema.ExclusiveMinimum)
	case "exclusiveMaximum":
		return setFloat64Pointer(gen.info.Fset, arg, &schema.ExclusiveMaximum)
	case "minLength":
		return setIntPointer(gen.info.Fset, arg, &schema.MinLength)
	case "maxLength":
		return setIntPointer(gen.info.Fset, arg, &schema.MaxLength)
	case "pattern":
		schema.Pattern = arg.Arg
		return nil
	case "minItems":
		return setIntPointer(gen.info.Fset, arg, &schema.MinItems)
	case "maxItems":
		return setIntPointer(gen.info.Fset, arg, &schema.MaxItems)
	case "uniqueItems":
		return setBoolean(gen.info.Fset, arg, &schema.UniqueItems)
	case "minContains":
		return setIntPointer(gen.info.Fset, arg, &schema.MinContains)
	case "maxContains":
		return setIntPointer(gen.info.Fset, arg, &schema.MaxContains)
	case "minProperties":
		return setIntPointer(gen.info.Fset, arg, &schema.MinProperties)
	case "maxProperties":
		return setIntPointer(gen.info.Fset, arg, &schema.MaxProperties)
	case "required":
		return setJSON(gen.info.Fset, arg, &schema.Required)
	case "dependentRequired":
		return setJSON(gen.info.Fset, arg, &schema.DependentRequired)
	case "additionalProperties":
		var value *bool
		if err := setBooleanPointer(gen.info.Fset, args[1], &value); err != nil {
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
			if err := setJSON(gen.info.Fset, arg, &value); err != nil {
				return err
			}
			if schema.Extra == nil {
				schema.Extra = make(map[string]any, 1)
			}
			schema.Extra[propertyName] = value
			return nil
		}

		// Return an error for all other property names
		return fmt.Errorf("%s: unsupported property %q in directive %s", gen.info.Fset.Position(args[0].Pos), propertyName, directiveString(dir))
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
