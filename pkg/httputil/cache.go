package httputil

import (
	"fmt"
	"net/http"
)

const defaultAge = 86400 // one day in seconds

// AllowCaching adds caching headers
func AllowCaching(headers http.Header) {
	// allow the client to cache this for a day
	headers.Set("Cache-Control", fmt.Sprintf("max-age=%d", defaultAge))
}
