package httputil

import "net/http"

// WithBasicAuth produces a [RequestEditorFunc] that sets
// basic auth with [http.Request.SetBasicAuth] for all requests.
func WithBasicAuth(username, password string) RequestEditorFunc {
	return func(req *http.Request) error {
		req.SetBasicAuth(username, password)
		return nil
	}
}
