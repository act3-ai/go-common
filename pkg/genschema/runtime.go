package genschema

import (
	"fmt"

	"github.com/iancoleman/orderedmap"
	"github.com/invopop/jsonschema"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

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
		versionSchema, err := ForAPIVersion(r, scheme, gv)
		if err != nil {
			return groupSchema, err
		}

		// "#/$defs/" + version
		groupSchema.AllOf = append(groupSchema.AllOf, apiVersionLinker(versionSchema.ID.String(), gv))

		groupSchema.Definitions[gv.Version] = versionSchema
	}

	return groupSchema, nil
}

// ForAPIVersion creates a JSONSchema validator for an API Version recognized by a runtime.Scheme.
//
// The resulting schema validates all Kinds recognized by the Scheme as part of the GroupVersion
// by mapping each "kind" to a subschema.
func ForAPIVersion(r *jsonschema.Reflector, scheme *runtime.Scheme, gv schema.GroupVersion) (*jsonschema.Schema, error) {
	versionSchema := &jsonschema.Schema{
		Version:     jsonschema.Version,
		ID:          jsonschema.ID("https://" + gv.Group).Add(gv.Version),
		Description: fmt.Sprintf("Version %s of the API %s", gv.Version, gv.Version),
		Definitions: make(jsonschema.Definitions),
	}

	// Iterate over each defined kind for this version of this group
	for name := range scheme.KnownTypes(gv) {
		kindSchema := ForAPIKind(r, scheme, gv.WithKind(name))

		// Add rule to associate the GroupVersionKind with
		// the newly-added schema definition
		versionSchema.AllOf = append(
			versionSchema.AllOf,
			kindLinker(
				kindSchema.ID.String(),
				gv.WithKind(name),
			),
		)

		versionSchema.Definitions[name] = kindSchema
	}

	return versionSchema, nil
}

// ForAPIKind creates a JSONSchema validator for an API GroupVersionKind recognized by a runtime.Scheme.
func ForAPIKind(r *jsonschema.Reflector, scheme *runtime.Scheme, gvk schema.GroupVersionKind) *jsonschema.Schema {
	r.SetBaseSchemaID(jsonschema.ID("https://" + gvk.Version).Add(gvk.Version).String())

	kindType := scheme.KnownTypes(gvk.GroupVersion())[gvk.Kind]
	kindSchema := r.ReflectFromType(kindType)

	return kindSchema
}

// func groupLinker(refToGroupDefinition, groupName string) *jsonschema.Schema {
// 	// Regex to match any apiVersion specifying the group "groupName"
// 	pattern := fmt.Sprintf("^%s(/.*)*$", groupName)

// 	propMap := orderedmap.New()
// 	propMap.Set("apiVersion", &jsonschema.Schema{
// 		Pattern: pattern,
// 	})

// 	gvkRule := &jsonschema.Schema{
// 		If: &jsonschema.Schema{
// 			Properties: propMap,
// 		},
// 		Then: &jsonschema.Schema{
// 			Ref: refToGroupDefinition,
// 		},
// 	}

// 	return gvkRule
// }

// kindLinker creates a JSONSchema "if/then" condition to associate objects matching
// the "apiVersion" field to the subschema for that GroupVersion
func apiVersionLinker(refToVersionDefinition string, gv schema.GroupVersion) *jsonschema.Schema {
	propMap := orderedmap.New()
	propMap.Set("apiVersion", &jsonschema.Schema{
		Const: gv.String(),
	})

	gvkRule := &jsonschema.Schema{
		If: &jsonschema.Schema{
			Properties: propMap,
		},
		Then: &jsonschema.Schema{
			Ref: refToVersionDefinition,
		},
	}

	return gvkRule
}

// kindLinker creates a JSONSchema "if/then" condition to associate objects matching
// the "apiVersion" and "kind" fields to the subschema for that GroupVersionKind
func kindLinker(refToKindDefinition string, gvk schema.GroupVersionKind) *jsonschema.Schema {
	apiVersionConst := &jsonschema.Schema{
		Const: gvk.GroupVersion().String(),
	}
	kindConst := &jsonschema.Schema{
		Const: gvk.Kind,
	}

	propMap := orderedmap.New()
	propMap.Set("apiVersion", apiVersionConst)
	propMap.Set("kind", kindConst)

	gvkRule := &jsonschema.Schema{
		If: &jsonschema.Schema{
			Properties: propMap,
		},
		Then: &jsonschema.Schema{
			Ref: refToKindDefinition,
		},
	}

	return gvkRule
}
