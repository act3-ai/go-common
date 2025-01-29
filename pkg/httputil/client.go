package httputil

import "net/http"

// HTTPRequestDoer is an interface for HTTP clients.
type HTTPRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// HTTPRequestDoerFunc promotes a function to an HTTPRequestDoer.
type HTTPRequestDoerFunc func(req *http.Request) (*http.Response, error)

// Do implements [HTTPRequestDoer].
func (client HTTPRequestDoerFunc) Do(req *http.Request) (*http.Response, error) {
	return client(req)
}
