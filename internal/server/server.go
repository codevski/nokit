package server

import (
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/codevski/defuse/internal/api"
	"github.com/codevski/defuse/internal/auth"
	"github.com/codevski/defuse/internal/rcon"
	"github.com/codevski/defuse/internal/store"
)

type Server struct {
	logger *slog.Logger
	dist   fs.FS
	store  *store.Store
	rcon   *rcon.Manager
	creds  auth.Credentials
}

func New(logger *slog.Logger, dist fs.FS, st *store.Store, mgr *rcon.Manager, creds auth.Credentials) *Server {
	return &Server{
		logger: logger,
		dist:   dist,
		store:  st,
		rcon:   mgr,
		creds:  creds,
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("POST /api/login", api.Wrap(s.logger, api.LoginHandler(s.creds, s.store)))
	mux.Handle("GET /api/health", api.Wrap(s.logger, s.health))

	protected := http.NewServeMux()
	protected.Handle("POST /api/logout", api.Wrap(s.logger, api.LogoutHandler(s.store)))
	protected.Handle("GET /api/me", api.Wrap(s.logger, api.MeHandler()))
	protected.Handle("GET /api/servers", api.Wrap(s.logger, api.ListServersHandler(s.store)))
	protected.Handle("POST /api/servers", api.Wrap(s.logger, api.AddServerHandler(s.store, s.rcon)))
	protected.Handle("DELETE /api/servers/{id}", api.Wrap(s.logger, api.DeleteServerHandler(s.store, s.rcon)))
	protected.Handle("POST /api/servers/{id}/rcon", api.Wrap(s.logger, api.RCONHandler(s.rcon)))

	mux.Handle("/api/", auth.Middleware(s.store, protected))

	mux.Handle("/", s.spaHandler())

	return s.logRequests(s.recoverPanics(mux))
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) error {
	return api.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) spaHandler() http.Handler {
	fileServer := http.FileServer(http.FS(s.dist))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := s.dist.Open(r.URL.Path[1:])
		if err == nil {
			f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}
		r2 := r.Clone(r.Context())
		r2.URL.Path = "/"
		fileServer.ServeHTTP(w, r2)
	})
}

func (s *Server) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.logger.Info("request", "method", r.Method, "path", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func (s *Server) recoverPanics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				s.logger.Error("panic recovered", "error", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
