// Package oapiutil implements helper functions for utilizing OpenAPI specifications.
package oapiutil

import (
	"bytes"
	"io"
	"net/http"

	"github.com/act3-ai/go-common/pkg/httputil"
)

// SpecHandler creates an [http.Handler] to serve an OpenAPI specification.
func SpecHandler(loadSpec func() ([]byte, error)) http.Handler {
	return httputil.RootHandler(func(w http.ResponseWriter, r *http.Request) error {
		spec, err := loadSpec()
		if err != nil {
			return httputil.NewHTTPError(err, http.StatusNotFound, "Fetching OpenAPI spec")
		}
		// Call SpecReaderHandler
		SpecReaderHandler(bytes.NewReader(spec)).ServeHTTP(w, r)
		return nil
	})
}

// SpecHandler creates an [http.Handler] to serve an OpenAPI specification.
func SpecReaderHandler(r io.Reader) http.Handler {
	return httputil.RootHandler(func(w http.ResponseWriter, _ *http.Request) error {
		w.Header().Set("Content-Type", "application/json")
		_, err := io.Copy(w, r)
		if err != nil {
			return httputil.NewHTTPError(err, http.StatusInternalServerError, "Writing OpenAPI spec")
		}
		return nil
	})
}
