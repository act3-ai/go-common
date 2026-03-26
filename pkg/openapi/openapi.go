// Package openapi provides Go types for representing an OpenAPI v3.2.0 document.
//
// The OpenAPI Specification (OAS) defines a standard, language-agnostic interface
// to HTTP APIs which allows both humans and computers to discover and understand
// the capabilities of the service without access to source code, documentation,
// or through network traffic inspection.
//
// Schema Objects in this package use *jsonschema.Schema from
// github.com/google/jsonschema-go/jsonschema directly, as OpenAPI 3.2 fully
// adopts JSON Schema draft 2020-12 for its Schema Object.
//
// Reference: https://spec.openapis.org/oas/v3.2.0
package openapi

import (
	"encoding/json"
	"iter"
	"net/http"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"
)

// ─────────────────────────────────────────────────────────────────────────────
// OpenAPI Object
// ─────────────────────────────────────────────────────────────────────────────

// Document is the root object of the OpenAPI Description.
// At least one of the Components, Paths, or Webhooks fields MUST be present.
//
// https://spec.openapis.org/oas/v3.2.0#openapi-object
type Document struct {
	// REQUIRED. This string MUST be the version number of the OpenAPI
	// Specification that the OpenAPI document uses. The openapi field SHOULD
	// be used by tooling to interpret the OpenAPI document. This is not related
	// to the Info.Version string, which describes the OpenAPI document's version.
	OpenAPI string `json:"openapi"`

	// This string MUST be in the form of a URI reference as defined by RFC 3986
	// Section 4.1. The $self field provides the self-assigned URI of this document,
	// which also serves as its base URI in accordance with RFC 3986 Section 5.1.1.
	// Implementations MUST support identifying the targets of API description URIs
	// using the URI defined by this field when it is present.
	Self string `json:"$self,omitempty"`

	// REQUIRED. Provides metadata about the API. The metadata MAY be used by
	// tooling as required.
	Info Info `json:"info"`

	// The default value for the $schema keyword within Schema Objects contained
	// within this OAS document. This MUST be in the form of a URI.
	JSONSchemaDialect string `json:"jsonSchemaDialect,omitempty"`

	// An array of Server Objects, which provide connectivity information to a
	// target server. If the servers field is not provided, or is an empty array,
	// the default value would be an array consisting of a single Server Object
	// with a url value of "/".
	Servers []Server `json:"servers,omitempty"`

	// The available paths and operations for the API.
	Paths Paths `json:"paths,omitempty"`

	// The incoming webhooks that MAY be received as part of this API and that
	// the API consumer MAY choose to implement. Closely related to the callbacks
	// feature, this section describes requests initiated other than by an API
	// call, for example by an out of band registration. The key name is a unique
	// string to refer to each webhook, while the (optionally referenced) Path
	// Item Object describes a request that may be initiated by the API provider
	// and the expected responses.
	Webhooks map[string]*PathItem `json:"webhooks,omitempty"`

	// An element to hold various Objects for the OpenAPI Description.
	Components Components `json:"components,omitzero"`

	// A declaration of which security mechanisms can be used across the API.
	// The list of values includes alternative Security Requirement Objects that
	// can be used. Only one of the Security Requirement Objects need to be
	// satisfied to authorize a request. Individual operations can override this
	// definition. The list can be incomplete, up to being empty or absent. To
	// make security explicitly optional, an empty security requirement ({}) can
	// be included in the array.
	Security []SecurityRequirement `json:"security,omitempty"`

	// A list of tags used by the OpenAPI Description with additional metadata.
	// The order of the tags can be used to reflect on their order by the parsing
	// tools. Not all tags that are used by the Operation Object must be declared.
	// The tags that are not declared MAY be organized randomly or based on the
	// tools' logic. Each tag name in the list MUST be unique.
	Tags []Tag `json:"tags,omitempty"`

	// Additional external documentation.
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitempty"`

	// Specification extensions (keys MUST begin with "x-").
	Extensions map[string]json.RawMessage `json:"-"`
}

// GetSchema produces the schema at the given reference.
func (doc *Document) GetSchema(ref string) (*jsonschema.Schema, bool) {
	if doc == nil || doc.Components.Schemas == nil {
		return nil, false
	}
	name, ok := strings.CutPrefix(ref, "#/components/schemas/")
	if !ok {
		return nil, false
	}
	schema, ok := doc.Components.Schemas[name]
	return schema, ok
}

// ─────────────────────────────────────────────────────────────────────────────
// Info Object
// ─────────────────────────────────────────────────────────────────────────────

// Info provides metadata about the API.
// The metadata MAY be used by the clients if needed, and MAY be presented in
// editing or documentation generation tools for convenience.
//
// https://spec.openapis.org/oas/v3.2.0#info-object
type Info struct {
	// REQUIRED. The title of the API.
	Title string `json:"title"`

	// A short summary of the API.
	Summary string `json:"summary,omitempty"`

	// A description of the API. CommonMark syntax MAY be used for rich text
	// representation.
	Description string `json:"description,omitempty"`

	// A URI for the Terms of Service for the API. This MUST be in the form of
	// a URI.
	TermsOfService string `json:"termsOfService,omitempty"`

	// The contact information for the exposed API.
	Contact *Contact `json:"contact,omitempty"`

	// The license information for the exposed API.
	License *License `json:"license,omitempty"`

	// REQUIRED. The version of the OpenAPI document (which is distinct from the
	// OpenAPI Specification version or the version of the API being described or
	// the version of the OpenAPI Description).
	Version string `json:"version"`

	// Specification extensions (keys MUST begin with "x-").
	Extensions map[string]json.RawMessage `json:"-"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Contact Object
// ─────────────────────────────────────────────────────────────────────────────

// Contact provides contact information for the exposed API.
//
// https://spec.openapis.org/oas/v3.2.0#contact-object
type Contact struct {
	// The identifying name of the contact person/organization.
	Name string `json:"name,omitempty"`

	// The URI for the contact information. This MUST be in the form of a URI.
	URL string `json:"url,omitempty"`

	// The email address of the contact person/organization. This MUST be in the
	// form of an email address.
	Email string `json:"email,omitempty"`

	// Specification extensions (keys MUST begin with "x-").
	Extensions map[string]json.RawMessage `json:"-"`
}

// ─────────────────────────────────────────────────────────────────────────────
// License Object
// ─────────────────────────────────────────────────────────────────────────────

// License provides license information for the exposed API.
//
// https://spec.openapis.org/oas/v3.2.0#license-object
type License struct {
	// REQUIRED. The license name used for the API.
	Name string `json:"name"`

	// An SPDX-Licenses expression for the API. The identifier field is mutually
	// exclusive of the url field.
	Identifier string `json:"identifier,omitempty"`

	// A URI for the license used for the API. This MUST be in the form of a URI.
	// The url field is mutually exclusive of the identifier field.
	URL string `json:"url,omitempty"`

	// Specification extensions (keys MUST begin with "x-").
	Extensions map[string]json.RawMessage `json:"-"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Server Object
// ─────────────────────────────────────────────────────────────────────────────

// Server represents a Server.
//
// https://spec.openapis.org/oas/v3.2.0#server-object
type Server struct {
	// REQUIRED. A URL to the target host. This URL supports Server Variables
	// and MAY be relative, to indicate that the host location is relative to
	// the location where the document containing the Server Object is being
	// served. Query and fragment MUST NOT be part of this URL. Variable
	// substitutions will be made when a variable is named in {braces}.
	URL string `json:"url"`

	// An optional string describing the host designated by the URL. CommonMark
	// syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty"`

	// An optional unique string to refer to the host designated by the URL.
	Name string `json:"name,omitempty"`

	// A map between a variable name and its value. The value is used for
	// substitution in the server's URL template.
	Variables map[string]*ServerVariable `json:"variables,omitempty"`

	// Specification extensions (keys MUST begin with "x-").
	Extensions map[string]json.RawMessage `json:"-"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Server Variable Object
// ─────────────────────────────────────────────────────────────────────────────

// ServerVariable represents a Server Variable for server URL template substitution.
//
// https://spec.openapis.org/oas/v3.2.0#server-variable-object
type ServerVariable struct {
	// An enumeration of string values to be used if the substitution options
	// are from a limited set. The array MUST NOT be empty.
	Enum []string `json:"enum,omitempty"`

	// REQUIRED. The default value to use for substitution, which SHALL be sent
	// if an alternate value is not supplied. If the enum is defined, the value
	// MUST exist in the enum's values. Note that this behavior is different from
	// the Schema Object's default keyword, which documents the receiver's
	// behavior rather than inserting the value into the data.
	Default string `json:"default"`

	// An optional description for the server variable. CommonMark syntax MAY be
	// used for rich text representation.
	Description string `json:"description,omitempty"`

	// Specification extensions (keys MUST begin with "x-").
	Extensions map[string]json.RawMessage `json:"-"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Components Object
// ─────────────────────────────────────────────────────────────────────────────

// Components holds a set of reusable objects for different aspects of the OAS.
// All objects defined within the Components Object will have no effect on the
// API unless they are explicitly referenced from outside the Components Object.
//
// https://spec.openapis.org/oas/v3.2.0#components-object
type Components struct {
	// An object to hold reusable Schema Objects.
	Schemas map[string]*Schema `json:"schemas,omitempty"`

	// An object to hold reusable Response Objects.
	Responses map[string]*Response `json:"responses,omitempty"`

	// An object to hold reusable Parameter Objects.
	Parameters map[string]*Parameter `json:"parameters,omitempty"`

	// An object to hold reusable Example Objects.
	Examples map[string]*Example `json:"examples,omitempty"`

	// An object to hold reusable Request Body Objects.
	RequestBodies map[string]*RequestBody `json:"requestBodies,omitempty"`

	// An object to hold reusable Header Objects.
	Headers map[string]*Header `json:"headers,omitempty"`

	// An object to hold reusable Security Scheme Objects.
	SecuritySchemes map[string]*SecurityScheme `json:"securitySchemes,omitempty"`

	// An object to hold reusable Link Objects.
	Links map[string]*Link `json:"links,omitempty"`

	// An object to hold reusable Callback Objects.
	Callbacks map[string]*Callback `json:"callbacks,omitempty"`

	// An object to hold reusable Path Item Objects.
	PathItems map[string]*PathItem `json:"pathItems,omitempty"`

	// An object to hold reusable Media Type Objects.
	MediaTypes map[string]*MediaType `json:"mediaTypes,omitempty"`

	// Specification extensions (keys MUST begin with "x-").
	Extensions map[string]json.RawMessage `json:"-"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Paths Object
// ─────────────────────────────────────────────────────────────────────────────

// Paths holds the relative paths to the individual endpoints and their
// operations. The path is appended to the URL from the Server Object in order
// to construct the full URL. The Paths Object MAY be empty, due to Access
// Control List (ACL) constraints.
//
// Each key MUST begin with a forward slash (/). The URL from the Server
// Object's url field, resolved and with template variables substituted, has the
// path appended (no relative URL resolution) to it in order to construct the
// full URL. Path templating is allowed. When matching URLs, concrete
// (non-templated) paths would be matched before their templated counterparts.
// Templated paths with the same hierarchy but different templated names MUST
// NOT exist as they are identical.
//
// https://spec.openapis.org/oas/v3.2.0#paths-object
type Paths map[string]*PathItem

// ─────────────────────────────────────────────────────────────────────────────
// Path Item Object
// ─────────────────────────────────────────────────────────────────────────────

// PathItem describes the operations available on a single path. A Path Item
// MAY be empty, due to ACL constraints. The path itself is still exposed to
// the documentation viewer but they will not know which operations and
// parameters are available.
//
// https://spec.openapis.org/oas/v3.2.0#path-item-object
type PathItem struct {
	// Allows for a referenced definition of this path item. The value MUST be
	// in the form of a URI, and the referenced structure MUST be in the form of
	// a Path Item Object. In case a Path Item Object field appears both in the
	// defined object and the referenced object, the behavior is undefined. See
	// the rules for resolving Relative References.
	Ref string `json:"$ref,omitempty"`

	// An optional string summary, intended to apply to all operations in this
	// path.
	Summary string `json:"summary,omitempty"`

	// An optional string description, intended to apply to all operations in
	// this path. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty"`

	// A definition of a GET operation on this path.
	Get *Operation `json:"get,omitempty"`

	// A definition of a PUT operation on this path.
	Put *Operation `json:"put,omitempty"`

	// A definition of a POST operation on this path.
	Post *Operation `json:"post,omitempty"`

	// A definition of a DELETE operation on this path.
	Delete *Operation `json:"delete,omitempty"`

	// A definition of an OPTIONS operation on this path.
	Options *Operation `json:"options,omitempty"`

	// A definition of a HEAD operation on this path.
	Head *Operation `json:"head,omitempty"`

	// A definition of a PATCH operation on this path.
	Patch *Operation `json:"patch,omitempty"`

	// A definition of a TRACE operation on this path.
	Trace *Operation `json:"trace,omitempty"`

	// A definition of a QUERY operation, as defined in the most recent IETF
	// draft (draft-ietf-httpbis-safe-method-w-body or its RFC successor), on
	// this path.
	Query *Operation `json:"query,omitempty"`

	// A map of additional operations on this path. The map key is the HTTP
	// method with the same capitalization that is to be sent in the request.
	// This map MUST NOT contain any entry for the methods that can be defined by
	// other fixed fields with Operation Object values (e.g. no POST entry, as
	// the post field is used for this method).
	AdditionalOperations map[string]*Operation `json:"additionalOperations,omitempty"`

	// An alternative servers array to service all operations in this path. If a
	// servers array is specified at the OpenAPI Object level, it will be
	// overridden by this value.
	Servers []Server `json:"servers,omitempty"`

	// A list of parameters that are applicable for all the operations described
	// under this path. These parameters can be overridden at the operation
	// level, but cannot be removed there. The list MUST NOT include duplicated
	// parameters. A unique parameter is defined by a combination of a name and
	// location. The list can use the Reference Object to link to parameters that
	// are defined in the OpenAPI Object's components.parameters.
	Parameters []*Parameter `json:"parameters,omitempty"`

	// Specification extensions (keys MUST begin with "x-").
	Extensions map[string]json.RawMessage `json:"-"`
}

// Reference produces a reference object if the object is a reference.
func (pathItem *PathItem) Reference() *Reference {
	if pathItem == nil || pathItem.Ref == "" {
		return nil
	}
	return &Reference{
		Ref:         pathItem.Ref,
		Summary:     pathItem.Summary,
		Description: pathItem.Description,
	}
}

// GetOperationForMethod returns the operation for the given HTTP method, or nil if no operation is defined.
func (pathItem *PathItem) GetOperationForMethod(method string) *Operation {
	switch strings.ToUpper(method) {
	case http.MethodGet, "":
		return pathItem.Get
	case http.MethodPut:
		return pathItem.Put
	case http.MethodPost:
		return pathItem.Post
	case http.MethodDelete:
		return pathItem.Delete
	case http.MethodOptions:
		return pathItem.Options
	case http.MethodHead:
		return pathItem.Head
	case http.MethodPatch:
		return pathItem.Patch
	case http.MethodTrace:
		return pathItem.Trace
	case "QUERY":
		return pathItem.Query
	default:
		if pathItem.AdditionalOperations != nil {
			return pathItem.AdditionalOperations[strings.ToLower(method)]
		}
		return nil
	}
}

// SetOperationForMethod sets the operation for the given HTTP method.
func (pathItem *PathItem) SetOperationForMethod(method string, op *Operation) {
	switch method {
	case http.MethodGet, "":
		pathItem.Get = op
	case http.MethodPut:
		pathItem.Put = op
	case http.MethodPost:
		pathItem.Post = op
	case http.MethodDelete:
		pathItem.Delete = op
	case http.MethodOptions:
		pathItem.Options = op
	case http.MethodHead:
		pathItem.Head = op
	case http.MethodPatch:
		pathItem.Patch = op
	case http.MethodTrace:
		pathItem.Trace = op
	case "QUERY":
		pathItem.Query = op
	default:
		if pathItem.AdditionalOperations == nil {
			pathItem.AdditionalOperations = make(map[string]*Operation, 1)
		}
		pathItem.AdditionalOperations[strings.ToLower(method)] = op
	}
}

// AllOperations returns an iterator over all operations defined on the path item.
func (pathItem *PathItem) AllOperations() iter.Seq2[string, *Operation] {
	return func(yield func(string, *Operation) bool) {
		if pathItem == nil {
			return
		}
		if pathItem.Get != nil && !yield(http.MethodGet, pathItem.Get) {
			return
		}
		if pathItem.Put != nil && !yield(http.MethodPut, pathItem.Put) {
			return
		}
		if pathItem.Post != nil && !yield(http.MethodPost, pathItem.Post) {
			return
		}
		if pathItem.Delete != nil && !yield(http.MethodDelete, pathItem.Delete) {
			return
		}
		if pathItem.Options != nil && !yield(http.MethodOptions, pathItem.Options) {
			return
		}
		if pathItem.Head != nil && !yield(http.MethodHead, pathItem.Head) {
			return
		}
		if pathItem.Patch != nil && !yield(http.MethodPatch, pathItem.Patch) {
			return
		}
		if pathItem.Trace != nil && !yield(http.MethodTrace, pathItem.Trace) {
			return
		}
		if pathItem.Query != nil && !yield("QUERY", pathItem.Query) {
			return
		}
		for method, op := range pathItem.AdditionalOperations {
			if !yield(method, op) {
				return
			}
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Operation Object
// ─────────────────────────────────────────────────────────────────────────────

// Operation describes a single API operation on a path.
//
// https://spec.openapis.org/oas/v3.2.0#operation-object
type Operation struct {
	// A list of tags for API documentation control. Tags can be used for
	// logical grouping of operations by resources or any other qualifier.
	Tags []string `json:"tags,omitempty"`

	// A short summary of what the operation does.
	Summary string `json:"summary,omitempty"`

	// A verbose explanation of the operation behavior. CommonMark syntax MAY be
	// used for rich text representation.
	Description string `json:"description,omitempty"`

	// Additional external documentation for this operation.
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitempty"`

	// Unique string used to identify the operation. The id MUST be unique among
	// all operations described in the API. The operationId value is
	// case-sensitive. Tools and libraries MAY use the operationId to uniquely
	// identify an operation, therefore, it is RECOMMENDED to follow common
	// programming naming conventions.
	OperationID string `json:"operationId,omitempty"`

	// A list of parameters that are applicable for this operation. If a
	// parameter is already defined at the Path Item, the new definition will
	// override it but can never remove it. The list MUST NOT include duplicated
	// parameters. A unique parameter is defined by a combination of a name and
	// location. The list can use the Reference Object to link to parameters that
	// are defined in the OpenAPI Object's components.parameters.
	Parameters []*Parameter `json:"parameters,omitempty"`

	// The request body applicable for this operation. The requestBody is fully
	// supported in HTTP methods where the HTTP specification has explicitly
	// defined semantics for request bodies. In other cases where the HTTP spec
	// discourages message content (such as GET and DELETE), requestBody is
	// permitted but does not have well-defined semantics and SHOULD be avoided
	// if possible.
	RequestBody *RequestBody `json:"requestBody,omitempty"`

	// The list of possible responses as they are returned from executing this
	// operation.
	Responses Responses `json:"responses,omitempty"`

	// A map of possible out-of band callbacks related to the parent operation.
	// The key is a unique identifier for the Callback Object. Each value in the
	// map is a Callback Object that describes a request that may be initiated by
	// the API provider and the expected responses.
	Callbacks map[string]*Callback `json:"callbacks,omitempty"`

	// Declares this operation to be deprecated. Consumers SHOULD refrain from
	// usage of the declared operation. Default value is false.
	Deprecated bool `json:"deprecated,omitempty"`

	// A declaration of which security mechanisms can be used for this operation.
	// The list of values includes alternative Security Requirement Objects that
	// can be used. Only one of the Security Requirement Objects need to be
	// satisfied to authorize a request. To make security optional, an empty
	// security requirement ({}) can be included in the array. This definition
	// overrides any declared top-level security. To remove a top-level security
	// declaration, an empty array can be used.
	Security []SecurityRequirement `json:"security,omitempty"`

	// An alternative servers array to service this operation. If a servers
	// array is specified at the Path Item Object or OpenAPI Object level, it
	// will be overridden by this value.
	Servers []Server `json:"servers,omitempty"`

	// Specification extensions (keys MUST begin with "x-").
	Extensions map[string]json.RawMessage `json:"-"`
}

// ─────────────────────────────────────────────────────────────────────────────
// External Documentation Object
// ─────────────────────────────────────────────────────────────────────────────

// ExternalDocumentation allows referencing an external resource for extended
// documentation.
//
// https://spec.openapis.org/oas/v3.2.0#external-documentation-object
type ExternalDocumentation struct {
	// A description of the target documentation. CommonMark syntax MAY be used
	// for rich text representation.
	Description string `json:"description,omitempty"`

	// REQUIRED. The URI for the target documentation. This MUST be in the form
	// of a URI.
	URL string `json:"url"`

	// Specification extensions (keys MUST begin with "x-").
	Extensions map[string]json.RawMessage `json:"-"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Parameter Object
// ─────────────────────────────────────────────────────────────────────────────

// Parameter describes a single operation parameter.
// A unique parameter is defined by a combination of a name and location.
//
// Parameter Objects MUST include either a content field or a schema field,
// but not both. When $ref is non-empty the other fields are ignored and the
// $ref is resolved to a Parameter in components.parameters.
//
// https://spec.openapis.org/oas/v3.2.0#parameter-object
type Parameter struct {
	// Allows this parameter to be defined by reference. When non-empty the
	// value MUST be a URI reference to a Parameter Object. All other fields
	// MUST be ignored when this field is set.
	Ref string `json:"$ref,omitempty"`

	// REQUIRED. The name of the parameter. Parameter names are case-sensitive.
	// If in is "path", the name field MUST correspond to a template expression
	// occurring within the path field in the Paths Object. If in is "header"
	// and the name field is "Accept", "Content-Type" or "Authorization", the
	// parameter definition SHALL be ignored.
	Name string `json:"name,omitempty"`

	// REQUIRED. The location of the parameter. Possible values are "query",
	// "querystring", "header", "path" or "cookie".
	In ParameterLocation `json:"in,omitempty"`

	// A brief description of the parameter. This could contain examples of use.
	// CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty"`

	// Determines whether this parameter is mandatory. If the parameter location
	// is "path", this field is REQUIRED and its value MUST be true. Otherwise,
	// the field MAY be included and its default value is false.
	Required bool `json:"required,omitempty"`

	// Specifies that a parameter is deprecated and SHOULD be transitioned out of
	// usage. Default value is false.
	Deprecated bool `json:"deprecated,omitempty"`

	// If true, clients MAY pass a zero-length string value in place of parameters
	// that would otherwise be omitted entirely, which the server SHOULD interpret
	// as the parameter being unused. Default value is false. If style is used,
	// and if behavior is n/a (cannot be serialized), the value of allowEmptyValue
	// SHALL be ignored. Interactions between this field and the parameter's Schema
	// Object are implementation-defined. This field is valid only for query
	// parameters.
	//
	// Deprecated: Use of this field is NOT RECOMMENDED, and it is likely to be
	// removed in a later revision.
	AllowEmptyValue bool `json:"allowEmptyValue,omitempty"`

	// Example of the parameter's potential value; see Working With Examples.
	Example json.RawMessage `json:"example,omitempty"`

	// Examples of the parameter's potential value; see Working With Examples.
	Examples map[string]*Example `json:"examples,omitempty"`

	// Describes how the parameter value will be serialized depending on the type
	// of the parameter value. Default values (based on value of in): for "query"
	// — "form"; for "path" — "simple"; for "header" — "simple"; for "cookie" —
	// "form".
	Style string `json:"style,omitempty"`

	// When this is true, parameter values of type array or object generate
	// separate parameters for each value of the array or key-value pair of the
	// map. For other types of parameters, or when style is "deepObject", this
	// field has no effect. When style is "form" or "cookie", the default value
	// is true. For all other styles, the default value is false.
	Explode *bool `json:"explode,omitempty"`

	// When this is true, parameter values are serialized using reserved
	// expansion, as defined by RFC6570 Section 3.2.3, which allows RFC3986's
	// reserved character set, as well as percent-encoded triples, to pass
	// through unchanged, while still percent-encoding all other disallowed
	// characters. The default value is false. This field only applies to in and
	// style values that automatically percent-encode.
	AllowReserved bool `json:"allowReserved,omitempty"`

	// The schema defining the type used for the parameter.
	Schema *Schema `json:"schema,omitempty"`

	// A map containing the representations for the parameter. The key is the
	// media type and the value describes it. The map MUST only contain one
	// entry. Mutually exclusive with Schema.
	Content map[string]*MediaType `json:"content,omitempty"`

	// Specification extensions (keys MUST begin with "x-").
	Extensions map[string]json.RawMessage `json:"-"`
}

// Reference produces a reference object if the object is a reference.
func (p *Parameter) Reference() *Reference {
	if p == nil || p.Ref == "" {
		return nil
	}
	return &Reference{
		Ref:         p.Ref,
		Summary:     "",
		Description: p.Description,
	}
}

// ParameterLocation defines the location of an operation parameter.
type ParameterLocation string

// Parameter location values.
const (
	ParameterLocationQuery       ParameterLocation = "query"
	ParameterLocationQueryString ParameterLocation = "querystring"
	ParameterLocationHeader      ParameterLocation = "header"
	ParameterLocationPath        ParameterLocation = "path"
	ParameterLocationCookie      ParameterLocation = "cookie"
)

// ParameterStyle defines how an operation parameter will be serialized depending
// on the type of the parameter value.
type ParameterStyle string

// Parameter style values.
const (
	ParameterStyleMatrix         ParameterStyle = "matrix"         // Path-style parameters defined by RFC6570 Section 3.2.7
	ParameterStyleLabel          ParameterStyle = "label"          // Label style parameters defined by RFC6570 Section 3.2.5
	ParameterStyleSimple         ParameterStyle = "simple"         // Simple style parameters defined by RFC6570 Section 3.2.2
	ParameterStyleForm           ParameterStyle = "form"           // Form style parameters defined by RFC6570 Section 3.2.8
	ParameterStyleSpaceDelimited ParameterStyle = "spaceDelimited" // Space separated array values or object properties and values
	ParameterStylePipeDelimited  ParameterStyle = "pipeDelimited"  // Pipe separated array values or object properties and values
	ParameterStyleDeepObject     ParameterStyle = "deepObject"     // Allows objects with scalar properties to be represented using form parameters
	ParameterStyleCookie         ParameterStyle = "cookie"         // Analogous to form, but following RFC6265 Cookie syntax rules
)

// ─────────────────────────────────────────────────────────────────────────────
// Request Body Object
// ─────────────────────────────────────────────────────────────────────────────

// RequestBody describes a single request body. When $ref is non-empty the other
// fields are ignored and the $ref is resolved to a RequestBody in
// components.requestBodies.
//
// https://spec.openapis.org/oas/v3.2.0#request-body-object
type RequestBody struct {
	// Allows this request body to be defined by reference. When non-empty the
	// value MUST be a URI reference to a Request Body Object. All other fields
	// MUST be ignored when this field is set.
	Ref string `json:"$ref,omitempty"`

	// A brief description of the request body. This could contain examples of
	// use. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty"`

	// REQUIRED. The content of the request body. The key is a media type or
	// media type range and the value describes it. The map SHOULD have at least
	// one entry; if it does not, the behavior is implementation-defined. For
	// requests that match multiple keys, only the most specific key is applicable.
	// e.g. "text/plain" overrides "text/*".
	Content map[string]*MediaType `json:"content,omitempty"`

	// Determines if the request body is required in the request. Defaults to
	// false.
	Required bool `json:"required,omitempty"`

	// Specification extensions (keys MUST begin with "x-").
	Extensions map[string]json.RawMessage `json:"-"`
}

// Reference produces a reference object if the object is a reference.
func (rb *RequestBody) Reference() *Reference {
	if rb == nil || rb.Ref == "" {
		return nil
	}
	return &Reference{
		Ref:         rb.Ref,
		Summary:     "",
		Description: rb.Description,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Media Type Object
// ─────────────────────────────────────────────────────────────────────────────

// MediaType describes content structured in accordance with the media type
// identified by its key. Multiple Media Type Objects can be used to describe
// content that can appear in any of several different media types.
//
// When example or examples are provided, the example SHOULD match the specified
// schema and be in the correct format as specified by the media type and its
// encoding. The example and examples fields are mutually exclusive.
//
// https://spec.openapis.org/oas/v3.2.0#media-type-object
type MediaType struct {
	// A schema describing the complete content of the request, response,
	// parameter, or header.
	Schema *Schema `json:"schema,omitempty"`

	// A schema describing each item within a sequential media type.
	ItemSchema *Schema `json:"itemSchema,omitempty"`

	// Example of the media type; see Working With Examples.
	Example json.RawMessage `json:"example,omitempty"`

	// Examples of the media type; see Working With Examples.
	Examples map[string]*Example `json:"examples,omitempty"`

	// A map between a property name and its encoding information, as defined
	// under Encoding By Name. The encoding field SHALL only apply when the media
	// type is multipart or application/x-www-form-urlencoded. If no Encoding
	// Object is provided for a property, the behavior is determined by the
	// default values documented for the Encoding Object. This field MUST NOT be
	// present if prefixEncoding or itemEncoding are present.
	Encoding map[string]*Encoding `json:"encoding,omitempty"`

	// An array of positional encoding information, as defined under Encoding By
	// Position. The prefixEncoding field SHALL only apply when the media type is
	// multipart. This field MUST NOT be present if encoding is present.
	PrefixEncoding []*Encoding `json:"prefixEncoding,omitempty"`

	// A single Encoding Object that provides encoding information for multiple
	// array items, as defined under Encoding By Position. The itemEncoding field
	// SHALL only apply when the media type is multipart. This field MUST NOT be
	// present if encoding is present.
	ItemEncoding *Encoding `json:"itemEncoding,omitempty"`

	// Specification extensions (keys MUST begin with "x-").
	Extensions map[string]json.RawMessage `json:"-"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Encoding Object
// ─────────────────────────────────────────────────────────────────────────────

// Encoding provides serialisation encoding for a specific schema property.
//
// https://spec.openapis.org/oas/v3.2.0#encoding-object
type Encoding struct {
	// The Content-Type for encoding a specific property. Default value depends
	// on the property type: for string with format being binary — application/
	// octet-stream; for other primitive types — text/plain; for object —
	// application/json; for array — the default is defined based on the inner
	// type.
	ContentType string `json:"contentType,omitempty"`

	// A map allowing additional information to be provided as headers. The key
	// is the name of the header and the value is its definition. This field
	// SHALL be ignored if the request body media type is not a multipart.
	Headers map[string]*Header `json:"headers,omitempty"`

	// Describes how a specific property value will be serialized depending on
	// its type.
	Style string `json:"style,omitempty"`

	// When this is true, property values of type array or object generate
	// separate parameters for each value of the array, or key-value pair of the
	// map.
	Explode *bool `json:"explode,omitempty"`

	// When this is true, parameter values are serialized using reserved
	// expansion. The default value is false.
	AllowReserved bool `json:"allowReserved,omitempty"`

	// Specification extensions (keys MUST begin with "x-").
	Extensions map[string]json.RawMessage `json:"-"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Responses Object
// ─────────────────────────────────────────────────────────────────────────────

// Responses is a map from HTTP status code string (e.g. "200", "4XX",
// "default") to a Response Object (which may itself carry a $ref).
//
// https://spec.openapis.org/oas/v3.2.0#responses-object
type Responses map[string]*Response

// ─────────────────────────────────────────────────────────────────────────────
// Response Object
// ─────────────────────────────────────────────────────────────────────────────

// Response describes a single response from an API operation. When $ref is
// non-empty the other fields are ignored and the $ref is resolved to a Response
// in components.responses.
//
// https://spec.openapis.org/oas/v3.2.0#response-object
type Response struct {
	// Allows this response to be defined by reference. When non-empty the value
	// MUST be a URI reference to a Response Object. All other fields MUST be
	// ignored when this field is set.
	Ref string `json:"$ref,omitempty"`

	// REQUIRED. A description of the response. CommonMark syntax MAY be used
	// for rich text representation.
	Description string `json:"description,omitempty"`

	// Maps a header name to its definition. RFC 9110 states header names are
	// case-insensitive. If a response header is defined with the name
	// "Content-Type", it SHALL be ignored.
	Headers map[string]*Header `json:"headers,omitempty"`

	// A map containing descriptions of potential response payloads. The key is
	// a media type or media type range and the value describes it. For responses
	// that match multiple keys, only the most specific key is applicable.
	Content map[string]*MediaType `json:"content,omitempty"`

	// A map of operations links that can be followed from the response. The key
	// of the map is a short name for the link, following the naming constraints
	// of the names for Component Objects.
	Links map[string]*Link `json:"links,omitempty"`

	// Specification extensions (keys MUST begin with "x-").
	Extensions map[string]json.RawMessage `json:"-"`
}

// Reference produces a reference object if the object is a reference.
func (r *Response) Reference() *Reference {
	if r == nil || r.Ref == "" {
		return nil
	}
	return &Reference{
		Ref:         r.Ref,
		Summary:     "",
		Description: r.Description,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Callback Object
// ─────────────────────────────────────────────────────────────────────────────

// Callback is a map of possible out-of-band callbacks related to the parent
// operation. Each value in the map is a Path Item Object that describes a set
// of requests that may be initiated by the API provider and the expected
// responses. The key value used to identify the path item object is an
// expression, evaluated at runtime, that identifies a URL to use for the
// callback operation. When $ref is non-empty the other fields are ignored and
// the $ref is resolved to a Callback in components.callbacks.
//
// https://spec.openapis.org/oas/v3.2.0#callback-object
type Callback struct {
	// Allows this callback to be defined by reference. When non-empty the value
	// MUST be a URI reference to a Callback Object. All other fields MUST be
	// ignored when this field is set.
	Ref string `json:"$ref,omitempty"`

	// The map of runtime-expression keys to Path Item Objects describing the
	// callback requests and expected responses.
	Paths map[string]*PathItem `json:"-"`
}

// Reference produces a reference object if the object is a reference.
func (cb *Callback) Reference() *Reference {
	if cb == nil || cb.Ref == "" {
		return nil
	}
	return &Reference{
		Ref:         cb.Ref,
		Summary:     "",
		Description: "",
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Example Object
// ─────────────────────────────────────────────────────────────────────────────

// Example provides an example value. When $ref is non-empty the other fields
// are ignored and the $ref is resolved to an Example in components.examples.
//
// https://spec.openapis.org/oas/v3.2.0#example-object
type Example struct {
	// Allows this example to be defined by reference. When non-empty the value
	// MUST be a URI reference to an Example Object. All other fields MUST be
	// ignored when this field is set.
	Ref string `json:"$ref,omitempty"`

	// Short description for the example.
	Summary string `json:"summary,omitempty"`

	// Long description for the example. CommonMark syntax MAY be used for rich
	// text representation.
	Description string `json:"description,omitempty"`

	// Embedded literal example. The value field and externalValue field are
	// mutually exclusive. To represent examples of media types that cannot
	// naturally be represented in JSON or YAML, use a string value to contain
	// the example with escaping where necessary.
	Value json.RawMessage `json:"value,omitempty"`

	// A URI that identifies the literal example. This provides the capability
	// to reference examples that cannot easily be included in JSON or YAML
	// documents. The value field and externalValue field are mutually exclusive.
	// See the rules for resolving Relative References.
	ExternalValue string `json:"externalValue,omitempty"`

	// Specification extensions (keys MUST begin with "x-").
	Extensions map[string]json.RawMessage `json:"-"`
}

// Reference produces a reference object if the object is a reference.
func (e *Example) Reference() *Reference {
	if e == nil || e.Ref == "" {
		return nil
	}
	return &Reference{
		Ref:         e.Ref,
		Summary:     e.Summary,
		Description: e.Description,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Link Object
// ─────────────────────────────────────────────────────────────────────────────

// Link represents a possible design-time link for a response. The presence of
// a link does not guarantee the caller's ability to successfully invoke it,
// rather it provides a known relationship and traversal mechanism between
// responses and other operations. When $ref is non-empty the other fields are
// ignored and the $ref is resolved to a Link in components.links.
//
// https://spec.openapis.org/oas/v3.2.0#link-object
type Link struct {
	// Allows this link to be defined by reference. When non-empty the value
	// MUST be a URI reference to a Link Object. All other fields MUST be
	// ignored when this field is set.
	Ref string `json:"$ref,omitempty"`

	// A relative or absolute URI reference to an OAS operation. This field is
	// mutually exclusive of the operationId field, and MUST point to an
	// Operation Object. Relative operationRef values MAY be used to locate an
	// existing Operation Object in the OpenAPI Description. See the rules for
	// resolving Relative References.
	OperationRef string `json:"operationRef,omitempty"`

	// The name of an existing, resolvable OAS operation, as defined with a
	// unique operationId. This field is mutually exclusive of the operationRef
	// field.
	OperationID string `json:"operationId,omitempty"`

	// A map representing parameters to pass to an operation as specified with
	// operationId or identified via operationRef. The key is the parameter name
	// to be used, whereas the value can be a constant or an expression to be
	// evaluated and passed to the linked operation.
	Parameters map[string]json.RawMessage `json:"parameters,omitempty"`

	// A literal value or {expression} to use as a request body when calling the
	// target operation.
	RequestBody json.RawMessage `json:"requestBody,omitempty"`

	// A description of the link. CommonMark syntax MAY be used for rich text
	// representation.
	Description string `json:"description,omitempty"`

	// A server object to be used by the target operation.
	Server *Server `json:"server,omitempty"`

	// Specification extensions (keys MUST begin with "x-").
	Extensions map[string]json.RawMessage `json:"-"`
}

// Reference produces a reference object if the object is a reference.
func (l *Link) Reference() *Reference {
	if l == nil || l.Ref == "" {
		return nil
	}
	return &Reference{
		Ref:         l.Ref,
		Summary:     "",
		Description: l.Description,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Header Object
// ─────────────────────────────────────────────────────────────────────────────

// Header follows the same structure as the Parameter Object, with the following
// changes: name MUST NOT be specified, it is given in the corresponding headers
// map; in MUST NOT be specified, it is implicitly in "header"; all traits that
// are affected by the location MUST be applicable to a location of "header".
// When $ref is non-empty the other fields are ignored and the $ref is resolved
// to a Header in components.headers.
//
// https://spec.openapis.org/oas/v3.2.0#header-object
type Header struct {
	// Allows this header to be defined by reference. When non-empty the value
	// MUST be a URI reference to a Header Object. All other fields MUST be
	// ignored when this field is set.
	Ref string `json:"$ref,omitempty"`

	// A brief description of the header. This could contain examples of use.
	// CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty"`

	// Determines whether this header is mandatory. Default value is false.
	Required bool `json:"required,omitempty"`

	// Specifies that the header is deprecated and SHOULD be transitioned out of
	// usage. Default value is false.
	Deprecated bool `json:"deprecated,omitempty"`

	// Describes how the header value will be serialized. The default (and only
	// legal value for headers) is "simple".
	Style HeaderStyle `json:"style,omitempty"`

	// When this is true, header values of type array or object generate separate
	// parameters for each value of the array or key-value pair of the map. The
	// default value is false.
	Explode *bool `json:"explode,omitempty"`

	// The schema defining the type used for the header.
	Schema *Schema `json:"schema,omitempty"`

	// Example of the header's potential value; see Working With Examples.
	Example json.RawMessage `json:"example,omitempty"`

	// Examples of the header's potential value; see Working With Examples.
	Examples map[string]*Example `json:"examples,omitempty"`

	// A map containing the representations for the header. The key is the
	// media type and the value describes it. The map MUST only contain one
	// entry. Mutually exclusive with Schema.
	Content map[string]*MediaType `json:"content,omitempty"`

	// Specification extensions (keys MUST begin with "x-").
	Extensions map[string]json.RawMessage `json:"-"`
}

// Reference produces a reference object if the object is a reference.
func (h *Header) Reference() *Reference {
	if h == nil || h.Ref == "" {
		return nil
	}
	return &Reference{
		Ref:         h.Ref,
		Summary:     "",
		Description: h.Description,
	}
}

// HeaderStyle describes how the header value will be serialized. The default
// (and only legal value for headers) is "simple".
type HeaderStyle string

// Header style values.
const (
	HeaderStyleSimple HeaderStyle = "simple"
)

// ─────────────────────────────────────────────────────────────────────────────
// Tag Object
// ─────────────────────────────────────────────────────────────────────────────

// Tag adds metadata to a single tag used by the Operation Object. It is not
// mandatory to have a Tag Object per tag defined in the Operation Object
// instances.
//
// https://spec.openapis.org/oas/v3.2.0#tag-object
type Tag struct {
	// REQUIRED. The name of the tag. Use this value in the tags array of an
	// Operation.
	Name string `json:"name"`

	// A short summary of the tag, used for display purposes.
	Summary string `json:"summary,omitempty"`

	// A description for the tag. CommonMark syntax MAY be used for rich text
	// representation.
	Description string `json:"description,omitempty"`

	// Additional external documentation for this tag.
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitempty"`

	// The name of a tag that this tag is nested under. The named tag MUST exist
	// in the API description, and circular references between parent and child
	// tags MUST NOT be used.
	Parent string `json:"parent,omitempty"`

	// A machine-readable string to categorize what sort of tag it is. Any string
	// value can be used; common uses are nav for Navigation, badge for visible
	// badges, audience for APIs used by different groups. A registry of the most
	// commonly used values is available.
	//
	// https://spec.openapis.org/registry/tag-kind
	Kind TagKind `json:"kind,omitempty"`

	// Specification extensions (keys MUST begin with "x-").
	Extensions map[string]json.RawMessage `json:"-"`
}

// TagKind is a machine-readable string to categorize what sort of tag it is.
// Any string value can be used; common uses are nav for Navigation, badge for
// visible badges, audience for APIs used by different groups. A registry of
// the most commonly used values is available.
//
// https://spec.openapis.org/registry/tag-kind
type TagKind string

// Tag kind values.
const (
	TagKindAudience TagKind = "audience"
	TagKindBadge    TagKind = "badge"
	TagKindNav      TagKind = "nav"
)

// ─────────────────────────────────────────────────────────────────────────────
// Reference Object
// ─────────────────────────────────────────────────────────────────────────────

// Reference is simple object to allow referencing other components in the
// OpenAPI Description, internally and externally.
//
// The $ref string value contains a URI RFC3986, which identifies the value
// being referenced.
//
// See the rules for resolving Relative References.
//
// https://spec.openapis.org/oas/v3.2.0#reference-object
type Reference struct {
	// REQUIRED. The reference identifier. This MUST be in the form of a URI.
	Ref string `json:"$ref"`

	// A short summary which by default SHOULD override that of the referenced
	// component. If the referenced object-type does not allow a summary field,
	// then this field has no effect.
	Summary string `json:"summary,omitempty"`

	// A short summary which by default SHOULD override that of the referenced
	// component. If the referenced object-type does not allow a summary field,
	// then this field has no effect.
	Description string `json:"description,omitempty"`
}

// Reference produces a reference object if the object is a reference.
func (r *Reference) Reference() *Reference {
	if r == nil || r.Ref == "" {
		return nil
	}
	return &Reference{
		Ref:         r.Ref,
		Summary:     r.Summary,
		Description: r.Description,
	}
}

// Referencer can reference a component in an OpenAPI Description.
type Referencer interface {
	// Reference produces a reference object if the object is a reference.
	Reference() *Reference
}

// Referencer interface checks.
var (
	_ Referencer = (*PathItem)(nil)
	_ Referencer = (*Parameter)(nil)
	_ Referencer = (*RequestBody)(nil)
	_ Referencer = (*Response)(nil)
	_ Referencer = (*Callback)(nil)
	_ Referencer = (*Example)(nil)
	_ Referencer = (*Link)(nil)
	_ Referencer = (*Header)(nil)
	_ Referencer = (*Reference)(nil)
	_ Referencer = (*SecurityScheme)(nil)
)

// SchemaReference produces a reference object if the schema object is a reference.
func SchemaReference(schema *Schema) *Reference {
	if schema == nil || schema.Ref == "" {
		return nil
	}
	return &Reference{
		Ref:         schema.Ref,
		Summary:     "",
		Description: schema.Description,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Schema Object
// ─────────────────────────────────────────────────────────────────────────────

// Schema allows the definition of input and output data types. These types can
// be objects, but also primitives and arrays. This object is a superset of the
// [JSON Schema Specification Draft 2020-12]. The empty schema (which allows
// any instance to validate) MAY be represented by the boolean value true and a
// schema which allows no instance to validate MAY be represented by the boolean
// value false.
//
// https://spec.openapis.org/oas/v3.2.0#schema-object
//
// [JSON Schema Specification Draft 2020-12]: https://json-schema.org/draft/2020-12
type Schema = jsonschema.Schema

// ─────────────────────────────────────────────────────────────────────────────
// Discriminator Object
// ─────────────────────────────────────────────────────────────────────────────

// Discriminator defines a discriminator for a polymorphic schema.
//
// When request bodies or response payloads may be one of a number of different
// schemas, these should use the JSON Schema anyOf or oneOf keywords to describe
// the possible schemas (see Composition and Inheritance).
//
// A polymorphic schema MAY include a Discriminator Object, which defines the
// name of the property that may be used as a hint for which schema of the anyOf
// or oneOf, or which schema that references the current schema in an allOf, is
// expected to validate the structure of the model. This hint can be used to aid
// in serialization, deserialization, and validation. The Discriminator Object
// does this by implicitly or explicitly associating the possible values of a
// named property with alternative schemas.
//
// Note that discriminator MUST NOT change the validation outcome of the schema.
//
// https://spec.openapis.org/oas/v3.2.0#discriminator-object
type Discriminator struct {
	// REQUIRED. The name of the discriminating property in the payload that will
	// hold the discriminating value. The discriminating property MAY be defined
	// as required or optional, but when defined as optional the Discriminator
	// Object MUST include a defaultMapping field that specifies which schema is
	// expected to validate the structure of the model when the discriminating
	// property is not present.
	PropertyName string `json:"propertyName"`

	// An object to hold mappings between payload values and schema names or URI
	// references.
	Mapping map[string]string `json:"mapping,omitempty"`

	// The schema name or URI reference to a schema that is expected to validate
	// the structure of the model when the discriminating property is not present
	// in the payload or contains a value for which there is no explicit or
	// implicit mapping.
	DefaultMapping map[string]string `json:"defaultMapping,omitempty"`
}

// ─────────────────────────────────────────────────────────────────────────────
// XML Object
// ─────────────────────────────────────────────────────────────────────────────

// XML contains metadata that allows for more fine-tuned XML model definitions.
// When using a Schema Object with XML, if no XML Object is present, the
// behavior is determined by the XML Object’s default field values. defines a
// discriminator for a polymorphic schema.
//
// https://spec.openapis.org/oas/v3.2.0#xml-object
type XML struct {
	// One of "element", "attribute", "text", "cdata", or "none", as explained
	// under XML Node Types. The default value is "none" if '$ref', '$dynamicRef',
	// or 'type: "array"' is present in the Schema Object containing the
	// XML Object, and "element" otherwise.
	NodeType string `json:"nodeType,omitempty"`

	// Sets the name of the element/attribute corresponding to the schema,
	// replacing the name that was inferred as described under XML Node Names.
	// This field SHALL be ignored if the nodeType is text, cdata, or none.
	Name string `json:"name,omitempty"`

	// The IRI (RFC3987) of the namespace definition. Value MUST be in the form
	// of a non-relative IRI.
	Namespace string `json:"namespace,omitempty"`

	// The prefix to be used for the name.
	Prefix string `json:"prefix,omitempty"`

	// Declares whether the property definition translates to an attribute
	// instead of an element. Default value is false. If nodeType is present,
	// this field MUST NOT be present.
	//
	// Deprecated: Use nodeType: "attribute" instead of attribute: true
	Attribute bool `json:"attribute,omitempty"`

	// MAY be used only for an array definition. Signifies whether the array is
	// wrapped (for example, <books><book/><book/></books>)
	// or unwrapped (<book/><book/>). Default value is false.
	// The definition takes effect only when defined alongside type being
	// "array" (outside the items). If nodeType is present, this field MUST NOT
	// be present.
	//
	// Deprecated: Use nodeType: "element" instead of wrapped: true
	Wrapped bool `json:"wrapped,omitempty"`
}

// XMLNodeType defines the type of an XML node.
//
// Each Schema Object describes a particular type of XML [DOM] node which is
// specified by the nodeType field, which has the following possible values.
// Except for the special value none, these values have numeric equivalents in
// the DOM specification which are given in parentheses after the name:
//
//   - element (1): The schema represents an element and describes its contents
//   - attribute (2): The schema represents an attribute and describes its value
//   - text (3): The schema represents a text node (parsed character data)
//   - cdata (4): The schema represents a CDATA section
//   - none: The schema does not correspond to any node in the XML document,
//     and the nodes corresponding to its subschema(s) are included directly
//     under its parent schema’s node
//
// The none type is useful for JSON Schema constructs that require more Schema Objects than XML nodes, such as a schema containing only $ref that exists to facilitate re-use rather than imply any structure.
//
// https://spec.openapis.org/oas/v3.2.0#xml-node-types
type XMLNodeType string

// XML node type values.
const (
	XMLNodeTypeElement   XMLNodeType = "element"   // The schema represents an element and describes its contents
	XMLNodeTypeAttribute XMLNodeType = "attribute" // The schema represents an attribute and describes its value
	XMLNodeTypeText      XMLNodeType = "text"      // The schema represents a text node (parsed character data)
	XMLNodeTypeCDATA     XMLNodeType = "cdata"     // The schema represents a CDATA section
	XMLNodeTypeNone      XMLNodeType = "none"      // The schema does not correspond to any node in the XML document, and the nodes corresponding to its subschema(s) are included directly under its parent schema’s node
)

// ─────────────────────────────────────────────────────────────────────────────
// Security Scheme Object
// ─────────────────────────────────────────────────────────────────────────────

// SecurityType defines the type of a security scheme.
type SecurityType string

// Security scheme type constants.
const (
	SecurityTypeAPIKey        SecurityType = "apiKey"
	SecurityTypeHTTP          SecurityType = "http"
	SecurityTypeMutualTLS     SecurityType = "mutualTLS"
	SecurityTypeOAuth2        SecurityType = "oauth2"
	SecurityTypeOpenIDConnect SecurityType = "openIdConnect"
)

// APIKeyLocation defines the location of the API key for a security scheme.
type APIKeyLocation string

// Security scheme API key location constants.
const (
	APIKeyLocationQuery  APIKeyLocation = "query"
	APIKeyLocationHeader APIKeyLocation = "header"
	APIKeyLocationCookie APIKeyLocation = "cookie"
)

// SecurityScheme defines a security scheme that can be used by the operations.
// When $ref is non-empty the other fields are ignored and the $ref is resolved
// to a SecurityScheme in components.securitySchemes.
//
// https://spec.openapis.org/oas/v3.2.0#security-scheme-object
type SecurityScheme struct {
	// Allows this security scheme to be defined by reference. When non-empty
	// the value MUST be a URI reference to a Security Scheme Object. All other
	// fields MUST be ignored when this field is set.
	Ref string `json:"$ref,omitempty"`

	// REQUIRED. The type of the security scheme. Valid values are "apiKey",
	// "http", "mutualTLS", "oauth2", "openIdConnect".
	Type SecurityType `json:"type,omitempty"`

	// A description for security scheme. CommonMark syntax MAY be used for rich
	// text representation.
	Description string `json:"description,omitempty"`

	// REQUIRED for apiKey. The name of the header, query or cookie parameter to
	// be used.
	Name string `json:"name,omitempty"`

	// REQUIRED for apiKey. The location of the API key. Valid values are
	// "query", "header" or "cookie".
	In APIKeyLocation `json:"in,omitempty"`

	// REQUIRED for http. The name of the HTTP Authorization scheme to be used
	// in the Authorization header as defined in RFC 9110. The values used
	// SHOULD be registered in the IANA Authentication Scheme registry.
	Scheme string `json:"scheme,omitempty"`

	// A hint to the client to identify how the bearer token is formatted.
	// Bearer tokens are usually generated by an authorization server, so this
	// information is primarily for documentation purposes. Applies to http
	// "bearer" scheme only.
	BearerFormat string `json:"bearerFormat,omitempty"`

	// REQUIRED for oauth2. An object containing configuration information for
	// the flow types supported.
	Flows *OAuthFlows `json:"flows,omitempty"`

	// REQUIRED for openIdConnect. Well-known URL to discover the OpenID
	// provider metadata.
	OpenIDConnectURL string `json:"openIdConnectUrl,omitempty"`

	// Specification extensions (keys MUST begin with "x-").
	Extensions map[string]json.RawMessage `json:"-"`
}

// Reference produces a reference object if the object is a reference.
func (s *SecurityScheme) Reference() *Reference {
	if s == nil || s.Ref == "" {
		return nil
	}
	return &Reference{
		Ref:         s.Ref,
		Summary:     "",
		Description: s.Description,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// OAuth Flows Object
// ─────────────────────────────────────────────────────────────────────────────

// OAuthFlows allows configuration of the supported OAuth Flows.
//
// https://spec.openapis.org/oas/v3.2.0#oauth-flows-object
type OAuthFlows struct {
	// Configuration for the OAuth Implicit flow.
	Implicit *OAuthFlow `json:"implicit,omitempty"`

	// Configuration for the OAuth Resource Owner Password flow.
	Password *OAuthFlow `json:"password,omitempty"`

	// Configuration for the OAuth Client Credentials flow. Previously called
	// application in OpenAPI 2.0.
	ClientCredentials *OAuthFlow `json:"clientCredentials,omitempty"`

	// Configuration for the OAuth Authorization Code flow. Previously called
	// accessCode in OpenAPI 2.0.
	AuthorizationCode *OAuthFlow `json:"authorizationCode,omitempty"`

	// Specification extensions (keys MUST begin with "x-").
	Extensions map[string]json.RawMessage `json:"-"`
}

// ─────────────────────────────────────────────────────────────────────────────
// OAuth Flow Object
// ─────────────────────────────────────────────────────────────────────────────

// OAuthFlow describes configuration details for a supported OAuth Flow.
//
// https://spec.openapis.org/oas/v3.2.0#oauth-flow-object
type OAuthFlow struct {
	// REQUIRED for implicit and authorizationCode. The authorization URL to be
	// used for this flow. This MUST be in the form of a URI. The OAuth2
	// standard requires the use of TLS.
	AuthorizationURL string `json:"authorizationUrl,omitempty"`

	// REQUIRED for password, clientCredentials, and authorizationCode. The
	// token URL to be used for this flow. This MUST be in the form of a URI.
	// The OAuth2 standard requires the use of TLS.
	TokenURL string `json:"tokenUrl,omitempty"`

	// The URL to be used for obtaining refresh tokens. This MUST be in the form
	// of a URI. The OAuth2 standard requires the use of TLS.
	RefreshURL string `json:"refreshUrl,omitempty"`

	// REQUIRED. The available scopes for the OAuth2 security scheme. A map
	// between the scope name and a short description for it. The map MAY be
	// empty.
	Scopes map[string]string `json:"scopes"`

	// Specification extensions (keys MUST begin with "x-").
	Extensions map[string]json.RawMessage `json:"-"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Security Requirement Object
// ─────────────────────────────────────────────────────────────────────────────

// SecurityRequirement lists the required security schemes to execute an
// operation. The name used for each property MUST correspond to a security
// scheme declared in the Security Schemes under the Components Object.
//
// Security Requirement Objects that contain multiple schemes require that all
// schemes MUST be satisfied for a request to be authorized. This enables
// support for scenarios where multiple query parameters or HTTP headers are
// required to convey security information.
//
// When a list of Security Requirement Objects is defined on the OpenAPI Object
// or Operation Object, only one of the Security Requirement Objects in the list
// needs to be satisfied to authorize the request.
//
// For OAuth2 and OpenID Connect, the value is a list of scope names required
// for the execution. For all other security scheme types, the array MUST be
// empty.
//
// https://spec.openapis.org/oas/v3.2.0#security-requirement-object
type SecurityRequirement map[string][]string
