package httputil

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	// HeaderUsername is the header set by the auth system (reverse proxy) to denote the username
	HeaderUsername = "X-Auth-Username"

	// HeaderInstance is a header used for identify this unique request/response (primitive tracing)
	HeaderInstance = "X-Instance"

	// HeaderCreationDate denotes the date at which this item was first uploaded to the telemetry server (used for replication purposes)
	HeaderCreationDate = "X-Creation-Date"
)

const (
	// MediaTypeProblem is the content type for errors producted by the server
	MediaTypeProblem = "application/problem+json; charset=utf-8"
)

// ClientError is an error whose details to be shared with client.
type ClientError interface {
	Error() string
	// Extra KeyValue pairs associated with the error
	ErrorArgs() []any
	// ResponseBody returns response body.
	ResponseBody() ([]byte, error)
	// ResponseHeaders returns http status code and headers.
	ResponseHeaders() (int, map[string]string)
}

// TODO we could implement https://datatracker.ietf.org/doc/html/rfc7807
// This would add fields like type (URI), title, instance (URI) (but we need at least the UUID)

// HTTPError implements ClientError interface.
type HTTPError struct {
	Cause      error  `json:"-"`
	CauseArgs  []any  `json:"-"`
	Detail     string `json:"detail"`
	StatusCode int    `json:"-"`
	Status     string `json:"status"`
}

func (e *HTTPError) Error() string {
	if e.Cause == nil {
		return e.Detail
	}
	return e.Detail + " : " + e.Cause.Error()
}

// ErrorArgs returns extra KV args for logging the error
func (e *HTTPError) ErrorArgs() []any {
	return e.CauseArgs
}

// ResponseBody returns JSON response body.
func (e *HTTPError) ResponseBody() ([]byte, error) {
	body, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("error while marshalling response body: %w", err)
	}
	return body, nil
}

// ResponseHeaders returns http status code and headers.
func (e *HTTPError) ResponseHeaders() (int, map[string]string) {
	return e.StatusCode, map[string]string{
		"Content-Type": MediaTypeProblem,
	}
}

// ensure HTTPError implements ClientError
var _ error = &HTTPError{}
var _ ClientError = &HTTPError{}

// NewHTTPError returns a new error
func NewHTTPError(err error, statusCode int, detail string, extraKV ...any) *HTTPError {
	return &HTTPError{
		Cause:      err,
		CauseArgs:  extraKV,
		Detail:     detail,
		StatusCode: statusCode,
		Status:     http.StatusText(statusCode),
	}
}
