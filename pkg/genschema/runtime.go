package genschema

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/invopop/jsonschema"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// GenerateGroupSchemas is a helper to generate all the schemas you want into dir
func GenerateGroupSchemas(dir string, scheme *runtime.Scheme, apiGroups []string, moduleName string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create schema directory: %w", err)
	}

	/*
		JSON Schema Generator Setup

		AddGoComments: enables setting descriptions from Go comments
		SetBaseSchemaID: changes the $id field of the schema to start with this string
								rather than the module path
	*/

	r := new(jsonschema.Reflector)
	r.DoNotReference = true

	if moduleName != "" {
		// WARNING: because of the "./" argument, this only works when running on the source files
		// 	This can cause errors when running in an executable since it will try to parse any .go files
		// 	on the system. This limitation is why we generate the schema at build time and embed the
		// 	schema files into the executable.
		err := r.AddGoComments(moduleName, "./")
		if err != nil {
			return fmt.Errorf("could not add comments to schema generator: %w", err)
		}
	}

	// Iterate over each schema that needs generated
	for _, group := range apiGroups {
		// Create the JSON Schema
		schema, err := ForAPIGroup(r, scheme, group)
		if err != nil {
			return err
		}

		schemaFile := group + ".schema.json"
		if err = WriteSchema(schema, filepath.Join(dir, schemaFile)); err != nil {
			return err
		}
	}

	return nil
}

// ForAPIGroup creates a JSONSchema validator for an API Group recognized by a runtime.Scheme.
//
// The resulting schema validates all Kinds recognized by the Scheme as part of the Group
// by mapping each "apiVersion" and "kind" to a subschema.
func ForAPIGroup(r *jsonschema.Reflector, scheme *runtime.Scheme, group string) (*jsonschema.Schema, error) {
	// This defines a wrapper schema that validates for the entire group
	groupSchema := &jsonschema.Schema{
		Version:     jsonschema.Version,
		ID:          jsonschema.ID("https://" + group),
		Description: fmt.Sprintf("Definition of the API " + group),
		Definitions: make(jsonschema.Definitions),
	}

	// Iterate over each defined version for this group
	for _, gv := range scheme.PrioritizedVersionsForGroup(group) {
		versionSchema, typeNames, err := forAPIVersion(r, scheme, gv)
		if err != nil {
			return groupSchema, err
		}

		groupSchema.Definitions[gv.Version] = versionSchema

		// Add a rule for each definition
		for _, name := range typeNames {
			groupSchema.AllOf = append(groupSchema.AllOf, linkGVK(gv.WithKind(name)))
		}
	}

	return groupSchema, nil
}

// forAPIVersion creates a JSONSchema validator for an API Version recognized by a runtime.Scheme.
//
// The resulting schema validates all Kinds recognized by the Scheme as part of the GroupVersion
// by mapping each "kind" to a subschema.
func forAPIVersion(r *jsonschema.Reflector, scheme *runtime.Scheme, gv schema.GroupVersion) (*jsonschema.Schema, []string, error) {
	versionSchema := &jsonschema.Schema{
		Version: jsonschema.Version,
		ID:      jsonschema.ID("https://" + gv.Group).Add(gv.Version),
		// ID:          jsonschema.ID("/" + gv.Version),
		Description: fmt.Sprintf("Version %s of the API %s", gv.Version, gv.Version),
		Definitions: make(jsonschema.Definitions),
	}

	// Sort names so resulting schema is stable
	knownTypes := scheme.KnownTypes(gv)
	typeNames := make([]string, 0, len(knownTypes))
	for name := range knownTypes {
		typeNames = append(typeNames, name)
	}
	slices.Sort(typeNames)

	// Iterate over each defined kind for this version of this group
	for _, name := range typeNames {
		versionSchema.Definitions[name] = forAPIKind(r, scheme, gv.WithKind(name))
	}

	return versionSchema, typeNames, nil
}

// forAPIKind creates a JSONSchema validator for an API GroupVersionKind recognized by a runtime.Scheme.
func forAPIKind(r *jsonschema.Reflector, scheme *runtime.Scheme, gvk schema.GroupVersionKind) *jsonschema.Schema {
	r.SetBaseSchemaID(jsonschema.ID("https://" + gvk.Group).Add(gvk.Version).String())
	// r.SetBaseSchemaID("/" + gvk.Version)

	kindType := scheme.KnownTypes(gvk.GroupVersion())[gvk.Kind]
	kindSchema := r.ReflectFromType(kindType)

	kindSchema.Properties.Set("apiVersion", &jsonschema.Schema{
		Type:        "string",
		Const:       gvk.GroupVersion().String(),
		Description: "Identifies the API group name and version for this data",
	})

	kindSchema.Properties.Set("kind", &jsonschema.Schema{
		Type:        "string",
		Const:       gvk.Kind,
		Description: "Identifies the API kind for this data",
	})

	return kindSchema
}

// linkGVK creates a JSONSchema "if/then" condition to associate objects matching
// the "apiVersion" and "kind" fields to the subschema for that GroupVersionKind
func linkGVK(gvk schema.GroupVersionKind) *jsonschema.Schema {
	gvkRule := &jsonschema.Schema{
		If: &jsonschema.Schema{Properties: jsonschema.NewProperties()},
		Then: &jsonschema.Schema{
			Ref: "#/$defs/" + gvk.Version + "/$defs/" + gvk.Kind,
		},
	}

	//
	gvkRule.If.Properties.Set("apiVersion", &jsonschema.Schema{
		Const: gvk.GroupVersion().String(),
	})

	gvkRule.If.Properties.Set("kind", &jsonschema.Schema{
		Const: gvk.Kind,
	})

	return gvkRule
}
