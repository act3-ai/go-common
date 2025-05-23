package httputil

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/act3-ai/go-common/pkg/logger"
	"github.com/act3-ai/go-common/pkg/logger/logutil"
)

// adapted from https://medium.com/@ozdemir.zynl/rest-api-error-handling-in-go-behavioral-type-assertion-509d93636afd

// RootHandler a wrapper around the handler functions to allow uniform error handling
type RootHandler func(http.ResponseWriter, *http.Request) error

// ServeHTTP performs uniform error handling.
// Implements http.Handler interface.
func (fn RootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := fn(w, r) // Call handler function
	if err == nil {
		return
	}

	ctx := r.Context()
	log := logger.FromContext(ctx)

	// Handle the error
	uid := InstanceFromContext(ctx).String()
	w.Header().Set(HeaderInstance, uid)

	var clientError ClientError
	if !errors.As(err, &clientError) {
		// If the error is not ClientError, assume that it is a ServerError.
		log.ErrorContext(ctx, "Internal error", logutil.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		// dump the instance out in the body as a field in JSON so the user can use it in reporting the error (so we can correlate it with the log on the server-side)
		if err := WriteJSON(w, map[string]any{"instance": uid, "statusCode": http.StatusInternalServerError}); err != nil {
			log.ErrorContext(ctx, "Failed to write error body", logutil.Err(err))
		}
		return
	}

	// It is a ClientError
	log.DebugContext(ctx, "ClientError", logutil.Err(clientError))

	// Provide the error to the client
	body, err := clientError.ResponseBody()
	if err != nil {
		log.ErrorContext(ctx, "Failed to get the response body", logutil.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	status, headers := clientError.ResponseHeaders()
	for k, v := range headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(status)
	if _, err := w.Write(body); err != nil {
		log.ErrorContext(ctx, "Failed to write error body", logutil.Err(err))
	}
}

// WriteJSON writes obj as JSON to the response
func WriteJSON(w http.ResponseWriter, obj any) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(obj)
}
