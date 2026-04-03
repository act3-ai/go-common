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
		DirectiveTool:   "jsonschema",
	}
}

// Generator generates JSON Schemas for Go types.
type Generator struct {
	info            *astutil.PackageInfo
	standardSchemas map[reflect.Type]func() *jsonschema.Schema
	results         map[reflect.Type]result

	// // Namer provides the name for a type.
	// Namer func(t reflect.Type) string

	// If enabled, change the default behavior to allow additional properties for struct types.
	StructAllowAdditionalProperties bool

	// Set "x-order" extension to object properties.
	SetXOrder bool

	// Set "x-go-type" extension to the Go type name.
	SetXGoType bool

	// Tool name for comment directives (default "jsonschema").
	DirectiveTool string

	// Extensions to provide a schema for types,
	// First provider to return a non-nil schema
	// will replace the regular schema generation.
	SchemaProviders []func(t reflect.Type) *jsonschema.Schema

	// Extensions to modify the generated schema for types.
	// Will be called after the regular schema generation
	// and after the SchemaExtender interface implementation.
	SchemaExtenders []func(t reflect.Type, schema *jsonschema.Schema)
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
	typeSchema, err := gen.generateSchemaForType(t)
	if err != nil {
		return nil, err
	}

	// Visit all reachable references
	refs, err := schemautil.ReachableRefs(gen, typeSchema)
	if err != nil {
		return nil, err
	}

	// Return schema as-is if it is the only definition
	if len(refs) == 0 {
		return typeSchema, nil
	}

	// Add all referenced schemas to a map
	defs := make(map[string]*jsonschema.Schema, len(refs)+1)
	defs[schemaName(t)] = typeSchema
	for _, ref := range refs {
		name := strings.TrimPrefix(ref, "#/$defs/")
		defs[name], _ = gen.GetSchema(ref)
	}

	return &jsonschema.Schema{
		Ref:  schemaRef(t),
		Defs: defs,
	}, err
}

func schemaName(t reflect.Type) string {
	// if t.PkgPath() == "" {
	// 	return t.Name()
	// }
	// return t.PkgPath() + "." + t.Name()
	return t.String()
}

func schemaRef(t reflect.Type) string {
	return "#/$defs/" + schemaName(t)
}

// GetSchema implements Registry.
func (gen *Generator) GetSchema(ref string) (*jsonschema.Schema, bool) {
	for t, result := range gen.results {
		if ref == schemaRef(t) {
			return result.Schema, true
		}
	}
	return nil, false
}

func (gen *Generator) generateSchemaForType(t reflect.Type) (schema *jsonschema.Schema, err error) {
	// Check if type is a standard type
	if schema, ok := standardTypeSchema(gen, t); ok {
		if gen.SetXGoType {
			setXGoType(schema, t)
		}
		return schema, nil
	}

	// Return cached schema if already generated
	if r, ok := gen.results[t]; ok {
		return r.Schema, r.Err
	}

	defer wrapf(&err, "generating schema for type %s", t)

	// Cache result if it is a named schema
	defer func() {
		if t.PkgPath() != "" && t.Name() != "" {
			gen.results[t] = result{Schema: schema, Err: err}
		}
	}()

	// Get the comment for this type
	comment := gen.getTypeComment(t)

	// Create schema from provider
	for _, prov := range gen.SchemaProviders {
		schema = prov(t)
		if schema != nil {
			break
		}
	}

	// If no SchemaProvider set the schema, continue normal generation
	if schema == nil {
		// If the type provides a schema, use that schema.
		if t.Implements(typeSchemaProvider) {
			v, _ := reflect.TypeAssert[SchemaProvider](reflect.New(t).Elem())
			schema = v.JSONSchema()
		} else {
			// Derive the schema from the kind
			schema, err = schemaFromKind(gen, t)
			if err != nil {
				return nil, err
			}

			// Add description from comment
			schema.Description = formatCommentAsDescription(comment)
		}
	}

	// Apply comment directives
	if err = gen.applySchemaDirectives(schema, comment); err != nil {
		return nil, err
	}

	// Add x-go-type extension
	if gen.SetXGoType {
		setXGoType(schema, t)
	}

	// Nest the reference in a subschema
	schemautil.NestReference(schema)

	// If the type defines a schema extension method, call the method.
	if t.Implements(typeSchemaExtender) {
		v, _ := reflect.TypeAssert[SchemaExtender](reflect.New(t).Elem())
		v.ExtendJSONSchema(schema)
	}

	// Call all SchemaExtenders
	for _, ext := range gen.SchemaExtenders {
		ext(t, schema)
	}

	return schema, nil
}

func generateSchemaForMap(gen *Generator, t reflect.Type) (*jsonschema.Schema, error) {
	var err error
	schema := &jsonschema.Schema{
		Type: schemautil.TypeObject,
	}

	// Generate schema for map keys
	schema.PropertyNames, err = gen.generateSchemaForType(t.Key())
	if err != nil {
		return nil, err
	}

	// Generate schema for map values
	schema.AdditionalProperties, err = gen.generateSchemaForType(t.Elem())
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
		Type:                 schemautil.TypeObject,
		Required:             props.Required,
		Properties:           props.Properties,
		PropertyOrder:        props.PropertyOrder,
		AdditionalProperties: schemautil.FalseSchema(),
	}

	if gen.StructAllowAdditionalProperties {
		schema.AdditionalProperties = schemautil.TrueSchema()
	}

	// Add x-order extension if enabled
	if gen.SetXOrder {
		order := 1
		for _, propSchema := range schemautil.OrderedProperties(schema) {
			schemautil.SetExtension(propSchema, schemautil.XOrder, order)
			order++
		}
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

	defer func() {
		// Nest the reference in a subschema
		schemautil.NestReference(schema)
	}()

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

func fullTypeName(t reflect.Type) string {
	if t == typeAny {
		return "any"
	}
	if t.PkgPath() != "" && t.Name() != "" {
		return t.PkgPath() + "." + t.Name()
	}
	return t.String()
}

func setXGoType(schema *jsonschema.Schema, t reflect.Type) {
	schemautil.SetExtension(schema, "x-go-type", fullTypeName(t))
}

// generateSchemaOrReference generates a schema for basic types or a reference to named types.
func generateSchemaOrReference(gen *Generator, t reflect.Type) (*jsonschema.Schema, error) {
	// Check if type is a standard type
	if schema, ok := standardTypeSchema(gen, t); ok {
		if gen.SetXGoType {
			setXGoType(schema, t)
		}
		return schema, nil
	}

	// Generate schema for the type
	schema, err := gen.generateSchemaForType(t)
	if err != nil {
		return nil, err
	}

	// If type has PkgPath and Name, return reference to schema
	if t.PkgPath() != "" && t.Name() != "" {
		return &jsonschema.Schema{
			Ref: schemaRef(t),
		}, nil
	}

	// Return the schema inline
	return schema, nil
}

// Standard types.
var (
	typeAny        = reflect.TypeFor[any]()
	typeString     = reflect.TypeFor[string]()
	typeBool       = reflect.TypeFor[bool]()
	typeUint       = reflect.TypeFor[uint]()
	typeUint8      = reflect.TypeFor[uint8]()
	typeUint16     = reflect.TypeFor[uint16]()
	typeUint32     = reflect.TypeFor[uint32]()
	typeUint64     = reflect.TypeFor[uint64]()
	typeInt        = reflect.TypeFor[int]()
	typeInt8       = reflect.TypeFor[int8]()
	typeInt16      = reflect.TypeFor[int16]()
	typeInt32      = reflect.TypeFor[int32]()
	typeInt64      = reflect.TypeFor[int64]()
	typeFloat32    = reflect.TypeFor[float32]()
	typeFloat64    = reflect.TypeFor[float64]()
	typeTime       = reflect.TypeFor[time.Time]()
	typeRawMessage = reflect.TypeFor[json.RawMessage]()
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
	typeAny:        schemautil.TrueSchema,
	typeString:     schemaString,
	typeBool:       schemaBool,
	typeUint:       schemaUint,
	typeUint8:      schemaUint8,
	typeUint16:     schemaUint16,
	typeUint32:     schemaUint32,
	typeUint64:     schemaUint64,
	typeInt:        schemaInt,
	typeInt8:       schemaInt8,
	typeInt16:      schemaInt16,
	typeInt32:      schemaInt32,
	typeInt64:      schemaInt64,
	typeFloat32:    schemaFloat32,
	typeFloat64:    schemaFloat64,
	typeTime:       schemaTime,
	typeRawMessage: schemautil.TrueSchema,
}

func standardTypeSchema(gen *Generator, t reflect.Type) (*jsonschema.Schema, bool) {
	fn, ok := gen.standardSchemas[t]
	if !ok {
		return nil, false
	}
	return fn(), true
}

// Derive the schema from the kind.
func schemaFromKind(gen *Generator, t reflect.Type) (schema *jsonschema.Schema, err error) {
	switch t.Kind() {
	case reflect.Interface:
		return schemautil.TrueSchema(), nil
	case reflect.String:
		return schemaString(), nil
	case reflect.Bool:
		return schemaBool(), nil
	case reflect.Uint:
		return schemaUint(), nil
	case reflect.Uint8:
		return schemaUint8(), nil
	case reflect.Uint16:
		return schemaUint16(), nil
	case reflect.Uint32:
		return schemaUint32(), nil
	case reflect.Uint64:
		return schemaUint64(), nil
	case reflect.Int:
		return schemaInt(), nil
	case reflect.Int8:
		return schemaInt8(), nil
	case reflect.Int16:
		return schemaInt16(), nil
	case reflect.Int32:
		return schemaInt32(), nil
	case reflect.Int64:
		return schemaInt64(), nil
	case reflect.Float32:
		return schemaFloat32(), nil
	case reflect.Float64:
		return schemaFloat64(), nil
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
		return setJSONRawMessage(gen.info.Fset, arg, &schema.Default)
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
