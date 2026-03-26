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

func (doc Document) MarshalJSON() ([]byte, error) {
	return marshalWithExtensions(documentAlias(doc), doc.Extensions)
}

func (doc *Document) UnmarshalJSON(data []byte) error {
	return unmarshalWithExtensions(data, (*documentAlias)(doc), &doc.Extensions)
}

// ─────────────────────────────────────────────────────────────────────────────
// Info
// ─────────────────────────────────────────────────────────────────────────────

type infoAlias Info

func (info Info) MarshalJSON() ([]byte, error) {
	return marshalWithExtensions(infoAlias(info), info.Extensions)
}

func (info *Info) UnmarshalJSON(data []byte) error {
	return unmarshalWithExtensions(data, (*infoAlias)(info), &info.Extensions)
}

// ─────────────────────────────────────────────────────────────────────────────
// Contact
// ─────────────────────────────────────────────────────────────────────────────

type contactAlias Contact

func (contact Contact) MarshalJSON() ([]byte, error) {
	return marshalWithExtensions(contactAlias(contact), contact.Extensions)
}

func (contact *Contact) UnmarshalJSON(data []byte) error {
	return unmarshalWithExtensions(data, (*contactAlias)(contact), &contact.Extensions)
}

// ─────────────────────────────────────────────────────────────────────────────
// License
// ─────────────────────────────────────────────────────────────────────────────

type licenseAlias License

func (license License) MarshalJSON() ([]byte, error) {
	return marshalWithExtensions(licenseAlias(license), license.Extensions)
}

func (license *License) UnmarshalJSON(data []byte) error {
	return unmarshalWithExtensions(data, (*licenseAlias)(license), &license.Extensions)
}

// ─────────────────────────────────────────────────────────────────────────────
// Server
// ─────────────────────────────────────────────────────────────────────────────

type serverAlias Server

func (server Server) MarshalJSON() ([]byte, error) {
	return marshalWithExtensions(serverAlias(server), server.Extensions)
}

func (server *Server) UnmarshalJSON(data []byte) error {
	return unmarshalWithExtensions(data, (*serverAlias)(server), &server.Extensions)
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

func (item PathItem) MarshalJSON() ([]byte, error) {
	if ref := item.Reference(); ref != nil {
		return json.Marshal(ref)
	}
	return marshalWithExtensions(pathItemAlias(item), item.Extensions)
}

func (item *PathItem) UnmarshalJSON(data []byte) error {
	return unmarshalWithExtensions(data, (*pathItemAlias)(item), &item.Extensions)
}

// ─────────────────────────────────────────────────────────────────────────────
// Operation
// ─────────────────────────────────────────────────────────────────────────────

type operationAlias Operation

func (op Operation) MarshalJSON() ([]byte, error) {
	return marshalWithExtensions(operationAlias(op), op.Extensions)
}

func (op *Operation) UnmarshalJSON(data []byte) error {
	return unmarshalWithExtensions(data, (*operationAlias)(op), &op.Extensions)
}

// ─────────────────────────────────────────────────────────────────────────────
// ExternalDocumentation
// ─────────────────────────────────────────────────────────────────────────────

type externalDocumentationAlias ExternalDocumentation

func (docs ExternalDocumentation) MarshalJSON() ([]byte, error) {
	return marshalWithExtensions(externalDocumentationAlias(docs), docs.Extensions)
}

func (docs *ExternalDocumentation) UnmarshalJSON(data []byte) error {
	return unmarshalWithExtensions(data, (*externalDocumentationAlias)(docs), &docs.Extensions)
}

// ─────────────────────────────────────────────────────────────────────────────
// Parameter
// ─────────────────────────────────────────────────────────────────────────────

type parameterAlias Parameter

func (parameter Parameter) MarshalJSON() ([]byte, error) {
	if ref := parameter.Reference(); ref != nil {
		return json.Marshal(ref)
	}
	return marshalWithExtensions(parameterAlias(parameter), parameter.Extensions)
}

func (parameter *Parameter) UnmarshalJSON(data []byte) error {
	return unmarshalWithExtensions(data, (*parameterAlias)(parameter), &parameter.Extensions)
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

func (mt MediaType) MarshalJSON() ([]byte, error) {
	return marshalWithExtensions(mediaTypeAlias(mt), mt.Extensions)
}

func (mt *MediaType) UnmarshalJSON(data []byte) error {
	return unmarshalWithExtensions(data, (*mediaTypeAlias)(mt), &mt.Extensions)
}

// ─────────────────────────────────────────────────────────────────────────────
// Encoding
// ─────────────────────────────────────────────────────────────────────────────

type encodingAlias Encoding

func (enc Encoding) MarshalJSON() ([]byte, error) {
	return marshalWithExtensions(encodingAlias(enc), enc.Extensions)
}

func (enc *Encoding) UnmarshalJSON(data []byte) error {
	return unmarshalWithExtensions(data, (*encodingAlias)(enc), &enc.Extensions)
}

// ─────────────────────────────────────────────────────────────────────────────
// Response
// ─────────────────────────────────────────────────────────────────────────────

type responseAlias Response

func (response Response) MarshalJSON() ([]byte, error) {
	if ref := response.Reference(); ref != nil {
		return json.Marshal(ref)
	}
	return marshalWithExtensions(responseAlias(response), response.Extensions)
}

func (response *Response) UnmarshalJSON(data []byte) error {
	return unmarshalWithExtensions(data, (*responseAlias)(response), &response.Extensions)
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

func (example Example) MarshalJSON() ([]byte, error) {
	if ref := example.Reference(); ref != nil {
		return json.Marshal(ref)
	}
	return marshalWithExtensions(exampleAlias(example), example.Extensions)
}

func (example *Example) UnmarshalJSON(data []byte) error {
	return unmarshalWithExtensions(data, (*exampleAlias)(example), &example.Extensions)
}

// ─────────────────────────────────────────────────────────────────────────────
// Link
// ─────────────────────────────────────────────────────────────────────────────

type linkAlias Link

func (link Link) MarshalJSON() ([]byte, error) {
	if ref := link.Reference(); ref != nil {
		return json.Marshal(ref)
	}
	return marshalWithExtensions(linkAlias(link), link.Extensions)
}

func (link *Link) UnmarshalJSON(data []byte) error {
	return unmarshalWithExtensions(data, (*linkAlias)(link), &link.Extensions)
}

// ─────────────────────────────────────────────────────────────────────────────
// Header
// ─────────────────────────────────────────────────────────────────────────────

type headerAlias Header

func (header Header) MarshalJSON() ([]byte, error) {
	if ref := header.Reference(); ref != nil {
		return json.Marshal(ref)
	}
	return marshalWithExtensions(headerAlias(header), header.Extensions)
}

func (header *Header) UnmarshalJSON(data []byte) error {
	return unmarshalWithExtensions(data, (*headerAlias)(header), &header.Extensions)
}

// ─────────────────────────────────────────────────────────────────────────────
// Tag
// ─────────────────────────────────────────────────────────────────────────────

type tagAlias Tag

func (tag Tag) MarshalJSON() ([]byte, error) {
	return marshalWithExtensions(tagAlias(tag), tag.Extensions)
}

func (tag *Tag) UnmarshalJSON(data []byte) error {
	return unmarshalWithExtensions(data, (*tagAlias)(tag), &tag.Extensions)
}

// ─────────────────────────────────────────────────────────────────────────────
// SecurityScheme
// ─────────────────────────────────────────────────────────────────────────────

type securitySchemeAlias SecurityScheme

func (scheme SecurityScheme) MarshalJSON() ([]byte, error) {
	if ref := scheme.Reference(); ref != nil {
		return json.Marshal(ref)
	}
	return marshalWithExtensions(securitySchemeAlias(scheme), scheme.Extensions)
}

func (scheme *SecurityScheme) UnmarshalJSON(data []byte) error {
	return unmarshalWithExtensions(data, (*securitySchemeAlias)(scheme), &scheme.Extensions)
}

// ─────────────────────────────────────────────────────────────────────────────
// OAuthFlows
// ─────────────────────────────────────────────────────────────────────────────

type oauthFlowsAlias OAuthFlows

func (flows OAuthFlows) MarshalJSON() ([]byte, error) {
	return marshalWithExtensions(oauthFlowsAlias(flows), flows.Extensions)
}

func (flows *OAuthFlows) UnmarshalJSON(data []byte) error {
	return unmarshalWithExtensions(data, (*oauthFlowsAlias)(flows), &flows.Extensions)
}

// ─────────────────────────────────────────────────────────────────────────────
// OAuthFlow
// ─────────────────────────────────────────────────────────────────────────────

type oauthFlowAlias OAuthFlow

func (flow OAuthFlow) MarshalJSON() ([]byte, error) {
	return marshalWithExtensions(oauthFlowAlias(flow), flow.Extensions)
}

func (flow *OAuthFlow) UnmarshalJSON(data []byte) error {
	return unmarshalWithExtensions(data, (*oauthFlowAlias)(flow), &flow.Extensions)
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
