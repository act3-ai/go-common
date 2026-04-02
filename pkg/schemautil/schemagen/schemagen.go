package schemagen

import (
	"fmt"
	"go/ast"
	"log/slog"
	"reflect"
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
	}
}

type Generator struct {
	info            *astutil.PackageInfo
	standardSchemas map[reflect.Type]func() *jsonschema.Schema
	results         map[reflect.Type]result

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
	// var err error
	schema := &jsonschema.Schema{
		Type:          schemautil.TypeObject,
		Properties:    make(map[string]*jsonschema.Schema, t.NumField()),
		PropertyOrder: make([]string, 0, t.NumField()),
	}

	// Generate schema for each field
	for field := range t.Fields() {
		if err := addSchemaForStructField(gen, t, schema, field); err != nil {
			return nil, fmt.Errorf("adding schema for struct field %s: %w", field.Name, err)
		}
	}

	return schema, nil
}

// addSchemaForStructField generates a schema for a struct field and adds it to the object schema.
func addSchemaForStructField(gen *Generator, t reflect.Type, schema *jsonschema.Schema, field reflect.StructField) error {
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
	propSchema, err := generateSchemaOrReference(gen, field.Type)
	if err != nil {
		return err
	}

	// Add description
	propSchema.Description = formatCommentAsDescription(comment)

	// Add to required properties if neither omitempty/omitzero are set
	if !(tagInfo.Omitempty || tagInfo.Omitzero) {
		schema.Required = append(schema.Required, propName)
	}

	// Add to property order
	schema.PropertyOrder = append(schema.PropertyOrder, propName)

	// Add to properties
	schema.Properties[propName] = propSchema

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
