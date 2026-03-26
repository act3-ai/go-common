package openapi

import (
	"iter"
	"strconv"

	"github.com/google/jsonschema-go/jsonschema"

	"github.com/act3-ai/go-common/pkg/jsonpointer"
)

// AllSchemas returns an iterator over all contained JSON Schemas.
// The first argument is an RFC6901 JSON Pointer to the schema within the object.
func (doc *Document) AllSchemas() iter.Seq2[string, *jsonschema.Schema] {
	return func(yield func(loc string, schema *jsonschema.Schema) bool) {
		if doc == nil {
			return
		}
		for loc, schema := range concatSeq2(
			withPrefix("/paths", doc.Paths.AllSchemas()),
			withPrefix("/webhooks", schemasFromMap(doc.Webhooks)),
			withPrefix("/components", doc.Components.AllSchemas()),
		) {
			if !yield(loc, schema) {
				return
			}
		}
	}
}

// AllSchemas returns an iterator over all contained JSON Schemas.
// The first argument is an RFC6901 JSON Pointer to the schema within the object.
func (paths Paths) AllSchemas() iter.Seq2[string, *jsonschema.Schema] {
	return schemasFromMap(paths)
}

// AllSchemas returns an iterator over all contained JSON Schemas.
// The first argument is an RFC6901 JSON Pointer to the schema within the object.
func (c *Components) AllSchemas() iter.Seq2[string, *jsonschema.Schema] {
	return func(yield func(loc string, schema *jsonschema.Schema) bool) {
		if c == nil {
			return
		}
		for name, schema := range c.Schemas {
			if !yield("/schemas/"+jsonpointer.Escape(name), schema) {
				return
			}
		}
		for loc, schema := range concatSeq2(
			withPrefix("/responses", schemasFromMap(c.Responses)),
			withPrefix("/parameters", schemasFromMap(c.Parameters)),
			withPrefix("/requestBodies", schemasFromMap(c.RequestBodies)),
			withPrefix("/headers", schemasFromMap(c.Headers)),
			withPrefix("/callbacks", schemasFromMap(c.Callbacks)),
			withPrefix("/pathItems", schemasFromMap(c.PathItems)),
			withPrefix("/mediaTypes", schemasFromMap(c.MediaTypes)),
		) {
			if !yield(loc, schema) {
				return
			}
		}
	}
}

// AllSchemas returns an iterator over all contained JSON Schemas.
// The first argument is an RFC6901 JSON Pointer to the schema within the object.
func (item *PathItem) AllSchemas() iter.Seq2[string, *jsonschema.Schema] {
	return func(yield func(loc string, schema *jsonschema.Schema) bool) {
		if item == nil {
			return
		}
		for loc, schema := range withPrefix("/parameters", schemasFromSlice(item.Parameters)) {
			if !yield(loc, schema) {
				return
			}
		}
		for method, op := range item.AllOperations() {
			for loc, schema := range op.AllSchemas() {
				if !yield("/"+jsonpointer.Escape(method)+loc, schema) {
					return
				}
			}
		}
	}
}

// AllSchemas returns an iterator over all contained JSON Schemas.
// The first argument is an RFC6901 JSON Pointer to the schema within the object.
func (op *Operation) AllSchemas() iter.Seq2[string, *jsonschema.Schema] {
	return func(yield func(loc string, schema *jsonschema.Schema) bool) {
		if op == nil {
			return
		}
		for loc, schema := range concatSeq2(
			withPrefix("/parameters", schemasFromSlice(op.Parameters)),
			withPrefix("/requestBody", op.RequestBody.AllSchemas()),
			withPrefix("/responses", schemasFromMap(op.Responses)),
			withPrefix("/callbacks", schemasFromMap(op.Callbacks)),
		) {
			if !yield(loc, schema) {
				return
			}
		}
	}
}

// AllSchemas returns an iterator over all contained JSON Schemas.
// The first argument is an RFC6901 JSON Pointer to the schema within the object.
func (cb *Callback) AllSchemas() iter.Seq2[string, *jsonschema.Schema] {
	return func(yield func(loc string, schema *jsonschema.Schema) bool) {
		if cb == nil {
			return
		}
		for loc, schema := range withPrefix("/paths", schemasFromMap(cb.Paths)) {
			if !yield(loc, schema) {
				return
			}
		}
	}
}

// AllSchemas returns an iterator over all contained JSON Schemas.
// The first argument is an RFC6901 JSON Pointer to the schema within the object.
func (parameter *Parameter) AllSchemas() iter.Seq2[string, *jsonschema.Schema] {
	return func(yield func(loc string, schema *jsonschema.Schema) bool) {
		if parameter == nil {
			return
		}
		if parameter.Schema != nil &&
			!yield("/schema", parameter.Schema) {
			return
		}
	}
}

// AllSchemas returns an iterator over all contained JSON Schemas.
// The first argument is an RFC6901 JSON Pointer to the schema within the object.
func (rb *RequestBody) AllSchemas() iter.Seq2[string, *jsonschema.Schema] {
	return func(yield func(loc string, schema *jsonschema.Schema) bool) {
		if rb == nil {
			return
		}
		for loc, schema := range withPrefix("/content", schemasFromMap(rb.Content)) {
			if !yield(loc, schema) {
				return
			}
		}
	}
}

// AllSchemas returns an iterator over all contained JSON Schemas.
// The first argument is an RFC6901 JSON Pointer to the schema within the object.
func (mt *MediaType) AllSchemas() iter.Seq2[string, *jsonschema.Schema] {
	return func(yield func(loc string, schema *jsonschema.Schema) bool) {
		if mt == nil {
			return
		}
		if mt.Schema != nil {
			if !yield("/schema", mt.Schema) {
				return
			}
		}
		if mt.ItemSchema != nil {
			if !yield("/itemSchema", mt.ItemSchema) {
				return
			}
		}
		for loc, schema := range concatSeq2(
			withPrefix("/encoding", schemasFromMap(mt.Encoding)),
			withPrefix("/prefixEncoding", schemasFromSlice(mt.PrefixEncoding)),
			withPrefix("/itemEncoding", mt.ItemEncoding.AllSchemas()),
		) {
			if !yield(loc, schema) {
				return
			}
		}
	}
}

// AllSchemas returns an iterator over all contained JSON Schemas.
// The first argument is an RFC6901 JSON Pointer to the schema within the object.
func (enc *Encoding) AllSchemas() iter.Seq2[string, *jsonschema.Schema] {
	return func(yield func(loc string, schema *jsonschema.Schema) bool) {
		if enc == nil {
			return
		}
		for loc, schema := range withPrefix("/headers", schemasFromMap(enc.Headers)) {
			if !yield(loc, schema) {
				return
			}
		}
	}
}

// AllSchemas returns an iterator over all contained JSON Schemas.
// The first argument is an RFC6901 JSON Pointer to the schema within the object.
func (r Responses) AllSchemas() iter.Seq2[string, *jsonschema.Schema] {
	return schemasFromMap(r)
}

// AllSchemas returns an iterator over all contained JSON Schemas.
// The first argument is an RFC6901 JSON Pointer to the schema within the object.
func (header *Header) AllSchemas() iter.Seq2[string, *jsonschema.Schema] {
	return func(yield func(loc string, schema *jsonschema.Schema) bool) {
		if header == nil {
			return
		}
		if header.Schema != nil &&
			!yield("/schema", header.Schema) {
			return
		}
	}
}

// AllSchemas returns an iterator over all contained JSON Schemas.
// The first argument is an RFC6901 JSON Pointer to the schema within the object.
func (response *Response) AllSchemas() iter.Seq2[string, *jsonschema.Schema] {
	return func(yield func(loc string, schema *jsonschema.Schema) bool) {
		if response == nil {
			return
		}
		for loc, schema := range concatSeq2(
			withPrefix("/headers", schemasFromMap(response.Headers)),
			withPrefix("/content", schemasFromMap(response.Content)),
		) {
			if !yield(loc, schema) {
				return
			}
		}
	}
}

// withPrefix prepends a prefix to each value produced by the iterator.
func withPrefix[T1 ~string, T2 any](prefix T1, values iter.Seq2[T1, T2]) iter.Seq2[T1, T2] {
	return func(yield func(T1, T2) bool) {
		for v1, v2 := range values {
			if !yield(prefix+v1, v2) {
				return
			}
		}
	}
}

// private interface for use in the iteration helpers below.
type schemaWalker interface {
	AllSchemas() iter.Seq2[string, *jsonschema.Schema]
}

// schemasFromMap returns an iterator over all JSON Schemas contained in the map values.
func schemasFromMap[M ~map[K]T, K ~string, T schemaWalker](values M) iter.Seq2[string, *jsonschema.Schema] {
	return func(yield func(string, *jsonschema.Schema) bool) {
		for key, value := range values {
			for loc, schema := range value.AllSchemas() {
				if !yield("/"+jsonpointer.Escape(string(key))+loc, schema) {
					return
				}
			}
		}
	}
}

// schemasFromSlice returns an iterator over all JSON Schemas contained in the slice values.
func schemasFromSlice[S ~[]T, T schemaWalker](values S) iter.Seq2[string, *jsonschema.Schema] {
	return func(yield func(string, *jsonschema.Schema) bool) {
		for i, value := range values {
			for loc, schema := range value.AllSchemas() {
				if !yield("/"+strconv.Itoa(i)+loc, schema) {
					return
				}
			}
		}
	}
}

// concatSeq2 concatenates a sequence of iterators.
func concatSeq2[T1, T2 any](iterators ...iter.Seq2[T1, T2]) iter.Seq2[T1, T2] {
	return func(yield func(T1, T2) bool) {
		for _, iterator := range iterators {
			for v1, v2 := range iterator {
				if !yield(v1, v2) {
					return
				}
			}
		}
	}
}
