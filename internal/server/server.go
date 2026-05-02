package server

import (
	"io/fs"
	"log/slog"
	"net/http"
	"time"

	"github.com/codevski/defuse/internal/api"
)

type Server struct {
	logger *slog.Logger
	dist   fs.FS
}

func New(logger *slog.Logger, dist fs.FS) *Server {
	return &Server{logger: logger, dist: dist}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /api/ping", api.Wrap(s.logger, s.ping))
	mux.Handle("/", http.FileServer(http.FS(s.dist)))

	return s.recoverPanics(s.logRequests(mux))
}

func (s *Server) ping(w http.ResponseWriter, r *http.Request) error {
	return api.JSON(w, http.StatusOK, map[string]string{"message": "pong from defuse"})
}

func (s *Server) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		s.logger.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", time.Since(start),
		)
	})
}

func (s *Server) recoverPanics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				s.logger.Error("panic recovered",
					"method", r.Method,
					"path", r.URL.Path,
					"panic", rec,
				)
				http.Error(w, "internal error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
