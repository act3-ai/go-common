package openapi

import (
	"bytes"
	"encoding/json"
	"strings"
)

// ─────────────────────────────────────────────────────────────────────────────
// Shared helpers
// ─────────────────────────────────────────────────────────────────────────────

// marshalWithExtensions serializes v (any JSON-serializable value) and merges
// the extension keys from ext into the resulting JSON object. The caller passes
// v as a type definition of the real struct so that the default encoder is used
// without recursing into this method.
func marshalWithExtensions(v any, ext map[string]json.RawMessage) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	if len(ext) == 0 {
		return b, nil
	}

	// b is guaranteed to be a JSON object (all our structs are objects).
	// Strip the trailing `}`, append the extension key/value pairs separated
	// correctly, then close the object again.
	//
	// We must handle the empty-object case: `{}` stripped to `{` must receive
	// the first extension key without a leading comma.
	trimmed := bytes.TrimRight(b, "}")
	isEmpty := bytes.Equal(bytes.TrimSpace(trimmed), []byte("{"))

	var buf bytes.Buffer
	buf.Write(trimmed)

	first := isEmpty
	for k, v := range ext {
		if !first {
			buf.WriteByte(',')
		}
		first = false
		key, err := json.Marshal(k)
		if err != nil {
			return nil, err
		}
		buf.Write(key)
		buf.WriteByte(':')
		buf.Write(v)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// unmarshalWithExtensions decodes data into v (a pointer to a type definition
// of the real struct) and collects every key that starts with "x-" into ext.
func unmarshalWithExtensions(data []byte, v any, ext *map[string]json.RawMessage) error {
	if err := json.Unmarshal(data, v); err != nil {
		return err
	}

	// Decode the raw object a second time to pick up extension keys.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	for k, val := range raw {
		if strings.HasPrefix(k, "x-") {
			if *ext == nil {
				*ext = make(map[string]json.RawMessage)
			}
			(*ext)[k] = val
		}
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Document
// ─────────────────────────────────────────────────────────────────────────────

type documentAlias Document

func (d Document) MarshalJSON() ([]byte, error) {
	return marshalWithExtensions(documentAlias(d), d.Extensions)
}

func (d *Document) UnmarshalJSON(data []byte) error {
	return unmarshalWithExtensions(data, (*documentAlias)(d), &d.Extensions)
}

// ─────────────────────────────────────────────────────────────────────────────
// Info
// ─────────────────────────────────────────────────────────────────────────────

type infoAlias Info

func (i Info) MarshalJSON() ([]byte, error) {
	return marshalWithExtensions(infoAlias(i), i.Extensions)
}

func (i *Info) UnmarshalJSON(data []byte) error {
	return unmarshalWithExtensions(data, (*infoAlias)(i), &i.Extensions)
}

// ─────────────────────────────────────────────────────────────────────────────
// Contact
// ─────────────────────────────────────────────────────────────────────────────

type contactAlias Contact

func (c Contact) MarshalJSON() ([]byte, error) {
	return marshalWithExtensions(contactAlias(c), c.Extensions)
}

func (c *Contact) UnmarshalJSON(data []byte) error {
	return unmarshalWithExtensions(data, (*contactAlias)(c), &c.Extensions)
}

// ─────────────────────────────────────────────────────────────────────────────
// License
// ─────────────────────────────────────────────────────────────────────────────

type licenseAlias License

func (l License) MarshalJSON() ([]byte, error) {
	return marshalWithExtensions(licenseAlias(l), l.Extensions)
}

func (l *License) UnmarshalJSON(data []byte) error {
	return unmarshalWithExtensions(data, (*licenseAlias)(l), &l.Extensions)
}

// ─────────────────────────────────────────────────────────────────────────────
// Server
// ─────────────────────────────────────────────────────────────────────────────

type serverAlias Server

func (s Server) MarshalJSON() ([]byte, error) {
	return marshalWithExtensions(serverAlias(s), s.Extensions)
}

func (s *Server) UnmarshalJSON(data []byte) error {
	return unmarshalWithExtensions(data, (*serverAlias)(s), &s.Extensions)
}

// ─────────────────────────────────────────────────────────────────────────────
// ServerVariable
// ─────────────────────────────────────────────────────────────────────────────

type serverVariableAlias ServerVariable

func (sv ServerVariable) MarshalJSON() ([]byte, error) {
	return marshalWithExtensions(serverVariableAlias(sv), sv.Extensions)
}

func (sv *ServerVariable) UnmarshalJSON(data []byte) error {
	return unmarshalWithExtensions(data, (*serverVariableAlias)(sv), &sv.Extensions)
}

// ─────────────────────────────────────────────────────────────────────────────
// Components
// ─────────────────────────────────────────────────────────────────────────────

type componentsAlias Components

func (c Components) MarshalJSON() ([]byte, error) {
	return marshalWithExtensions(componentsAlias(c), c.Extensions)
}

func (c *Components) UnmarshalJSON(data []byte) error {
	return unmarshalWithExtensions(data, (*componentsAlias)(c), &c.Extensions)
}

// ─────────────────────────────────────────────────────────────────────────────
// PathItem
// ─────────────────────────────────────────────────────────────────────────────

type pathItemAlias PathItem

func (p PathItem) MarshalJSON() ([]byte, error) {
	if ref := p.Reference(); ref != nil {
		return json.Marshal(ref)
	}
	return marshalWithExtensions(pathItemAlias(p), p.Extensions)
}

func (p *PathItem) UnmarshalJSON(data []byte) error {
	return unmarshalWithExtensions(data, (*pathItemAlias)(p), &p.Extensions)
}

// ─────────────────────────────────────────────────────────────────────────────
// Operation
// ─────────────────────────────────────────────────────────────────────────────

type operationAlias Operation

func (o Operation) MarshalJSON() ([]byte, error) {
	return marshalWithExtensions(operationAlias(o), o.Extensions)
}

func (o *Operation) UnmarshalJSON(data []byte) error {
	return unmarshalWithExtensions(data, (*operationAlias)(o), &o.Extensions)
}

// ─────────────────────────────────────────────────────────────────────────────
// ExternalDocumentation
// ─────────────────────────────────────────────────────────────────────────────

type externalDocumentationAlias ExternalDocumentation

func (e ExternalDocumentation) MarshalJSON() ([]byte, error) {
	return marshalWithExtensions(externalDocumentationAlias(e), e.Extensions)
}

func (e *ExternalDocumentation) UnmarshalJSON(data []byte) error {
	return unmarshalWithExtensions(data, (*externalDocumentationAlias)(e), &e.Extensions)
}

// ─────────────────────────────────────────────────────────────────────────────
// Parameter
// ─────────────────────────────────────────────────────────────────────────────

type parameterAlias Parameter

func (p Parameter) MarshalJSON() ([]byte, error) {
	if ref := p.Reference(); ref != nil {
		return json.Marshal(ref)
	}
	return marshalWithExtensions(parameterAlias(p), p.Extensions)
}

func (p *Parameter) UnmarshalJSON(data []byte) error {
	return unmarshalWithExtensions(data, (*parameterAlias)(p), &p.Extensions)
}

// ─────────────────────────────────────────────────────────────────────────────
// RequestBody
// ─────────────────────────────────────────────────────────────────────────────

type requestBodyAlias RequestBody

func (rb RequestBody) MarshalJSON() ([]byte, error) {
	if ref := rb.Reference(); ref != nil {
		return json.Marshal(ref)
	}
	return marshalWithExtensions(requestBodyAlias(rb), rb.Extensions)
}

func (rb *RequestBody) UnmarshalJSON(data []byte) error {
	return unmarshalWithExtensions(data, (*requestBodyAlias)(rb), &rb.Extensions)
}

// ─────────────────────────────────────────────────────────────────────────────
// MediaType
// ─────────────────────────────────────────────────────────────────────────────

type mediaTypeAlias MediaType

func (m MediaType) MarshalJSON() ([]byte, error) {
	return marshalWithExtensions(mediaTypeAlias(m), m.Extensions)
}

func (m *MediaType) UnmarshalJSON(data []byte) error {
	return unmarshalWithExtensions(data, (*mediaTypeAlias)(m), &m.Extensions)
}

// ─────────────────────────────────────────────────────────────────────────────
// Encoding
// ─────────────────────────────────────────────────────────────────────────────

type encodingAlias Encoding

func (e Encoding) MarshalJSON() ([]byte, error) {
	return marshalWithExtensions(encodingAlias(e), e.Extensions)
}

func (e *Encoding) UnmarshalJSON(data []byte) error {
	return unmarshalWithExtensions(data, (*encodingAlias)(e), &e.Extensions)
}

// ─────────────────────────────────────────────────────────────────────────────
// Response
// ─────────────────────────────────────────────────────────────────────────────

type responseAlias Response

func (r Response) MarshalJSON() ([]byte, error) {
	if ref := r.Reference(); ref != nil {
		return json.Marshal(ref)
	}
	return marshalWithExtensions(responseAlias(r), r.Extensions)
}

func (r *Response) UnmarshalJSON(data []byte) error {
	return unmarshalWithExtensions(data, (*responseAlias)(r), &r.Extensions)
}

// ─────────────────────────────────────────────────────────────────────────────
// Callback
//
// Callback is unusual: it carries an optional $ref plus a dynamic map of
// runtime-expression keys to PathItem values at the top level of the JSON
// object. Neither fits naturally into a plain struct with fixed tags, so we
// handle marshalling/unmarshalling manually.
// ─────────────────────────────────────────────────────────────────────────────

func (cb Callback) MarshalJSON() ([]byte, error) {
	if ref := cb.Reference(); ref != nil {
		return json.Marshal(ref)
	}
	// Encode the path map directly as the top-level object.
	return json.Marshal(cb.Paths)
}

func (cb *Callback) UnmarshalJSON(data []byte) error {
	// Peek for a $ref key.
	var peek struct {
		Ref string `json:"$ref"`
	}
	if err := json.Unmarshal(data, &peek); err != nil {
		return err
	}
	if peek.Ref != "" {
		cb.Ref = peek.Ref
		return nil
	}

	// Otherwise every key is a runtime expression mapped to a Path Item.
	return json.Unmarshal(data, &cb.Paths)
}

// ─────────────────────────────────────────────────────────────────────────────
// Example
// ─────────────────────────────────────────────────────────────────────────────

type exampleAlias Example

func (e Example) MarshalJSON() ([]byte, error) {
	if ref := e.Reference(); ref != nil {
		return json.Marshal(ref)
	}
	return marshalWithExtensions(exampleAlias(e), e.Extensions)
}

func (e *Example) UnmarshalJSON(data []byte) error {
	return unmarshalWithExtensions(data, (*exampleAlias)(e), &e.Extensions)
}

// ─────────────────────────────────────────────────────────────────────────────
// Link
// ─────────────────────────────────────────────────────────────────────────────

type linkAlias Link

func (l Link) MarshalJSON() ([]byte, error) {
	if ref := l.Reference(); ref != nil {
		return json.Marshal(ref)
	}
	return marshalWithExtensions(linkAlias(l), l.Extensions)
}

func (l *Link) UnmarshalJSON(data []byte) error {
	return unmarshalWithExtensions(data, (*linkAlias)(l), &l.Extensions)
}

// ─────────────────────────────────────────────────────────────────────────────
// Header
// ─────────────────────────────────────────────────────────────────────────────

type headerAlias Header

func (h Header) MarshalJSON() ([]byte, error) {
	if ref := h.Reference(); ref != nil {
		return json.Marshal(ref)
	}
	return marshalWithExtensions(headerAlias(h), h.Extensions)
}

func (h *Header) UnmarshalJSON(data []byte) error {
	return unmarshalWithExtensions(data, (*headerAlias)(h), &h.Extensions)
}

// ─────────────────────────────────────────────────────────────────────────────
// Tag
// ─────────────────────────────────────────────────────────────────────────────

type tagAlias Tag

func (t Tag) MarshalJSON() ([]byte, error) {
	return marshalWithExtensions(tagAlias(t), t.Extensions)
}

func (t *Tag) UnmarshalJSON(data []byte) error {
	return unmarshalWithExtensions(data, (*tagAlias)(t), &t.Extensions)
}

// ─────────────────────────────────────────────────────────────────────────────
// SecurityScheme
// ─────────────────────────────────────────────────────────────────────────────

type securitySchemeAlias SecurityScheme

func (s SecurityScheme) MarshalJSON() ([]byte, error) {
	if ref := s.Reference(); ref != nil {
		return json.Marshal(ref)
	}
	return marshalWithExtensions(securitySchemeAlias(s), s.Extensions)
}

func (s *SecurityScheme) UnmarshalJSON(data []byte) error {
	return unmarshalWithExtensions(data, (*securitySchemeAlias)(s), &s.Extensions)
}

// ─────────────────────────────────────────────────────────────────────────────
// OAuthFlows
// ─────────────────────────────────────────────────────────────────────────────

type oauthFlowsAlias OAuthFlows

func (o OAuthFlows) MarshalJSON() ([]byte, error) {
	return marshalWithExtensions(oauthFlowsAlias(o), o.Extensions)
}

func (o *OAuthFlows) UnmarshalJSON(data []byte) error {
	return unmarshalWithExtensions(data, (*oauthFlowsAlias)(o), &o.Extensions)
}

// ─────────────────────────────────────────────────────────────────────────────
// OAuthFlow
// ─────────────────────────────────────────────────────────────────────────────

type oauthFlowAlias OAuthFlow

func (o OAuthFlow) MarshalJSON() ([]byte, error) {
	return marshalWithExtensions(oauthFlowAlias(o), o.Extensions)
}

func (o *OAuthFlow) UnmarshalJSON(data []byte) error {
	return unmarshalWithExtensions(data, (*oauthFlowAlias)(o), &o.Extensions)
}

// ─────────────────────────────────────────────────────────────────────────────
// Compile-time interface assertions
// ─────────────────────────────────────────────────────────────────────────────

var (
	_ json.Marshaler   = Document{}
	_ json.Unmarshaler = (*Document)(nil)
	_ json.Marshaler   = Info{}
	_ json.Unmarshaler = (*Info)(nil)
	_ json.Marshaler   = Contact{}
	_ json.Unmarshaler = (*Contact)(nil)
	_ json.Marshaler   = License{}
	_ json.Unmarshaler = (*License)(nil)
	_ json.Marshaler   = Server{}
	_ json.Unmarshaler = (*Server)(nil)
	_ json.Marshaler   = ServerVariable{}
	_ json.Unmarshaler = (*ServerVariable)(nil)
	_ json.Marshaler   = Components{}
	_ json.Unmarshaler = (*Components)(nil)
	_ json.Marshaler   = PathItem{}
	_ json.Unmarshaler = (*PathItem)(nil)
	_ json.Marshaler   = Operation{}
	_ json.Unmarshaler = (*Operation)(nil)
	_ json.Marshaler   = ExternalDocumentation{}
	_ json.Unmarshaler = (*ExternalDocumentation)(nil)
	_ json.Marshaler   = Parameter{}
	_ json.Unmarshaler = (*Parameter)(nil)
	_ json.Marshaler   = RequestBody{}
	_ json.Unmarshaler = (*RequestBody)(nil)
	_ json.Marshaler   = MediaType{}
	_ json.Unmarshaler = (*MediaType)(nil)
	_ json.Marshaler   = Encoding{}
	_ json.Unmarshaler = (*Encoding)(nil)
	_ json.Marshaler   = Response{}
	_ json.Unmarshaler = (*Response)(nil)
	_ json.Marshaler   = Callback{}
	_ json.Unmarshaler = (*Callback)(nil)
	_ json.Marshaler   = Example{}
	_ json.Unmarshaler = (*Example)(nil)
	_ json.Marshaler   = Link{}
	_ json.Unmarshaler = (*Link)(nil)
	_ json.Marshaler   = Header{}
	_ json.Unmarshaler = (*Header)(nil)
	_ json.Marshaler   = Tag{}
	_ json.Unmarshaler = (*Tag)(nil)
	_ json.Marshaler   = SecurityScheme{}
	_ json.Unmarshaler = (*SecurityScheme)(nil)
	_ json.Marshaler   = OAuthFlows{}
	_ json.Unmarshaler = (*OAuthFlows)(nil)
	_ json.Marshaler   = OAuthFlow{}
	_ json.Unmarshaler = (*OAuthFlow)(nil)
)
