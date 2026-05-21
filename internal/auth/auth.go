package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	sessionCookie = "defuse_session"
	sessionTTL    = 30 * 24 * time.Hour // 30 days
)

type Credentials struct {
	username     string
	passwordHash []byte
}

type Store interface {
	CreateSession(token string, ttl time.Duration) error
	ValidSession(token string) (bool, error)
	DeleteSession(token string) error
}

func LoadCredentials() (Credentials, error) {
	u := os.Getenv("PANEL_USERNAME")
	p := os.Getenv("PANEL_PASSWORD")
	if u == "" {
		return Credentials{}, errors.New("PANEL_USERNAME env var is required")
	}
	if p == "" {
		return Credentials{}, errors.New("PANEL_PASSWORD env var is required")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	if err != nil {
		return Credentials{}, err
	}
	return Credentials{username: u, passwordHash: hash}, nil
}

func (c Credentials) Verify(username, password string) bool {
	if username != c.username {
		return false
	}
	return bcrypt.CompareHashAndPassword(c.passwordHash, []byte(password)) == nil
}

func Login(w http.ResponseWriter, st Store) error {
	token, err := randomToken()
	if err != nil {
		return err
	}
	if err := st.CreateSession(token, sessionTTL); err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    token,
		Path:     "/",
		MaxAge:   int(sessionTTL.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

func Logout(w http.ResponseWriter, r *http.Request, st Store) {
	c, err := r.Cookie(sessionCookie)
	if err == nil {
		_ = st.DeleteSession(c.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:   sessionCookie,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
}

func Authenticated(r *http.Request, st Store) (bool, error) {
	c, err := r.Cookie(sessionCookie)
	if err != nil {
		return false, nil
	}
	return st.ValidSession(c.Value)
}

func Middleware(st Store, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok, err := Authenticated(r, st)
		if err != nil || !ok {
			if len(r.URL.Path) >= 5 && r.URL.Path[:5] == "/api/" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
				return
			}
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
