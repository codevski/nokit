package api

import (
	"net/http"
	"time"

	"github.com/codevski/defuse/internal/auth"
	"github.com/codevski/defuse/internal/store"
)

type AuthStore interface {
	CreateSession(token string, ttl time.Duration) error
	ValidSession(token string) (bool, error)
	DeleteSession(token string) error
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func LoginHandler(creds auth.Credentials, st AuthStore) Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		var req LoginRequest
		if err := Decode(r, &req); err != nil {
			return err
		}
		if !creds.Verify(req.Username, req.Password) {
			return Unauthorized("invalid credentials")
		}
		if err := auth.Login(w, st); err != nil {
			return WrapHTTP(err, http.StatusInternalServerError, "session error")
		}
		return JSON(w, http.StatusOK, map[string]string{"ok": "true"})
	}
}

func LogoutHandler(st AuthStore) Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		auth.Logout(w, r, st)
		return JSON(w, http.StatusOK, map[string]string{"ok": "true"})
	}
}

func MeHandler() Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		return JSON(w, http.StatusOK, map[string]bool{"authenticated": true})
	}
}

var _ AuthStore = (*store.Store)(nil)
