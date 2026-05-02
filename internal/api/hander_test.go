package api

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestWrap_Success(t *testing.T) {
	h := Wrap(newTestLogger(), func(w http.ResponseWriter, r *http.Request) error {
		return JSON(w, http.StatusOK, map[string]string{"hello": "world"})
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", rr.Code, http.StatusOK)
	}
	var body map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["hello"] != "world" {
		t.Errorf("got body %v", body)
	}
}

func TestWrap_HTTPError(t *testing.T) {
	h := Wrap(newTestLogger(), func(w http.ResponseWriter, r *http.Request) error {
		return NotFound("server not found")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want %d", rr.Code, http.StatusNotFound)
	}
	var body map[string]string
	_ = json.NewDecoder(rr.Body).Decode(&body)
	if body["error"] != "server not found" {
		t.Errorf("got error %q, want %q", body["error"], "server not found")
	}
}

func TestWrap_GenericError(t *testing.T) {
	h := Wrap(newTestLogger(), func(w http.ResponseWriter, r *http.Request) error {
		return errors.New("database exploded")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("got status %d, want %d", rr.Code, http.StatusInternalServerError)
	}
	var body map[string]string
	_ = json.NewDecoder(rr.Body).Decode(&body)
	if body["error"] != "internal error" {
		t.Errorf("got error %q, expected generic message (cause should not leak)", body["error"])
	}
	// The actual cause must NOT appear in the response body.
	if strings.Contains(rr.Body.String(), "database exploded") {
		t.Error("internal error message leaked into response")
	}
}

func TestDecode_RejectsUnknownFields(t *testing.T) {
	type input struct {
		Name string `json:"name"`
	}

	body := strings.NewReader(`{"name":"sash","sneaky":true}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)

	var got input
	err := Decode(req, &got)
	if err == nil {
		t.Fatal("expected error for unknown field")
	}
	he := AsHTTPError(err)
	if he == nil || he.Status != http.StatusBadRequest {
		t.Errorf("expected 400 HTTPError, got %v", err)
	}
}

func TestWrapHTTP_PreservesCause(t *testing.T) {
	root := errors.New("root cause")
	wrapped := WrapHTTP(root, http.StatusConflict, "user already exists")

	if !errors.Is(wrapped, root) {
		t.Error("errors.Is should find the root cause")
	}
	if AsHTTPError(wrapped).Status != http.StatusConflict {
		t.Errorf("got status %d, want 409", AsHTTPError(wrapped).Status)
	}
}
