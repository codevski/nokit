package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

type Handler func(w http.ResponseWriter, r *http.Request) error

func Wrap(logger *slog.Logger, h Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := h(w, r)
		if err == nil {
			return
		}

		var he *HTTPError
		if errors.As(err, &he) {
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

func JSON(w http.ResponseWriter, status int, body any) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		return nil
	}
	return nil
}

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
