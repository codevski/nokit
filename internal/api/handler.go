package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

// Handler is an HTTP handler that returns an error.
//
// Errors of type *HTTPError are converted to responses using their status
// code and message. All other errors are logged with full context and
// returned to the client as a generic 500. Handlers should never write
// to the response and return an error in the same call.
type Handler func(w http.ResponseWriter, r *http.Request) error

// Wrap converts a Handler into a standard http.Handler, using the given
// logger for error and panic reporting.
func Wrap(logger *slog.Logger, h Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := h(w, r)
		if err == nil {
			return
		}

		var he *HTTPError
		if errors.As(err, &he) {
			// Log 5xx with the cause; 4xx is usually client noise and
			// only logged at debug level.
			if he.Status >= 500 {
				logger.Error("handler error",
					"method", r.Method,
					"path", r.URL.Path,
					"status", he.Status,
					"error", err,
				)
			} else {
				logger.Debug("handler client error",
					"method", r.Method,
					"path", r.URL.Path,
					"status", he.Status,
					"message", he.Message,
				)
			}
			writeError(w, he.Status, he.Message)
			return
		}

		logger.Error("handler error",
			"method", r.Method,
			"path", r.URL.Path,
			"status", http.StatusInternalServerError,
			"error", err,
		)
		writeError(w, http.StatusInternalServerError, "internal error")
	})
}

// JSON writes body as a JSON response with the given status. It always
// returns nil so handlers can `return api.JSON(...)` in a single line.
func JSON(w http.ResponseWriter, status int, body any) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		// Response is already partially written; nothing to do but log
		// upstream. Returning the error here would let the wrapper try
		// to write a second response, which would fail.
		return nil
	}
	return nil
}

// Decode reads and decodes the request body as JSON into dst. It returns
// a 400 HTTPError on malformed input so handlers can `if err := ...; err
// != nil { return err }` without further wrapping.
func Decode(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return WrapHTTP(err, http.StatusBadRequest, "invalid request body")
	}
	return nil
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
