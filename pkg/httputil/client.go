package httputil

import "net/http"

// Client is an interface for HTTP clients.
type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

// ClientFunc promotes a function to a [Client].
type ClientFunc func(req *http.Request) (*http.Response, error)

// Do implements [Client].
func (client ClientFunc) Do(req *http.Request) (*http.Response, error) {
	return client(req)
}

// ClientMiddlewareFunc is an alias for client middleware functions.
type ClientMiddlewareFunc = func(next Client) Client

// RequestEditorFunc edits HTTP requests.
type RequestEditorFunc = func(req *http.Request) error

// WrapClient wraps a [Client] with client middlewares.
func WrapClient(client Client, middlewares ...ClientMiddlewareFunc) Client {
	for _, mware := range middlewares {
		client = mware(client)
	}
	return client
}

// WithRequestEditors wraps a [Client] with request editor functions ([RequestEditorFunc]).
//
// Each given [RequestEditorFunc] will be called to modify the request before it is handled by [Client.Do].
//
// If a [RequestEditorFunc] returns an error, the error will be returned and the underlying [Client.Do] will not be called.
func WithRequestEditors(client Client, editors ...func(req *http.Request) error) Client {
	return ClientFunc(
		func(req *http.Request) (*http.Response, error) {
			// Perform all request edits
			for _, editor := range editors {
				if err := editor(req); err != nil {
					return nil, err
				}
			}
			return client.Do(req)
		},
	)
}
