package api

import (
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/codevski/defuse/internal/rcon"
	"github.com/codevski/defuse/internal/store"
)

type ServerStore interface {
	ListServers() ([]store.Server, error)
	GetServer(id string) (store.Server, error)
	AddServer(sv store.Server) error
	DeleteServer(id string) error
}

type ServerInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Host string `json:"host"`
}

func ListServersHandler(st ServerStore) Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		servers, err := st.ListServers()
		if err != nil {
			return WrapHTTP(err, http.StatusInternalServerError, "could not list servers")
		}
		out := make([]ServerInfo, len(servers))
		for i, s := range servers {
			out[i] = ServerInfo{ID: s.ID, Name: s.Name, Host: s.Host}
		}
		return JSON(w, http.StatusOK, out)
	}
}

type AddServerRequest struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	RCONPass string `json:"rcon_pass"`
}

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	s = strings.ToLower(s)
	s = slugRe.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

func AddServerHandler(st ServerStore, mgr *rcon.Manager) Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		var req AddServerRequest
		if err := Decode(r, &req); err != nil {
			return err
		}
		if req.Name == "" {
			return BadRequest("name is required")
		}
		if req.Host == "" {
			return BadRequest("host is required")
		}
		if req.RCONPass == "" {
			return BadRequest("rcon_pass is required")
		}

		id := slugify(req.Name)
		if id == "" {
			return BadRequest("name produces an empty id")
		}

		sv := store.Server{
			ID:        id,
			Name:      req.Name,
			Host:      req.Host,
			RCONPass:  req.RCONPass,
			CreatedAt: time.Now(),
		}
		if err := st.AddServer(sv); err != nil {
			if err == store.ErrConflict {
				return Conflict("a server with that name already exists")
			}
			return WrapHTTP(err, http.StatusInternalServerError, "could not save server")
		}

		go func() {
			_ = mgr.Connect(id, req.Host, req.RCONPass)
		}()

		return JSON(w, http.StatusCreated, ServerInfo{ID: id, Name: req.Name, Host: req.Host})
	}
}

func DeleteServerHandler(st ServerStore, mgr *rcon.Manager) Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		id := r.PathValue("id")
		if id == "" {
			return BadRequest("missing server id")
		}
		if _, err := st.GetServer(id); err != nil {
			if err == store.ErrNotFound {
				return NotFound("server not found")
			}
			return WrapHTTP(err, http.StatusInternalServerError, "could not fetch server")
		}
		if err := st.DeleteServer(id); err != nil {
			return WrapHTTP(err, http.StatusInternalServerError, "could not delete server")
		}
		mgr.Disconnect(id)
		return JSON(w, http.StatusOK, map[string]string{"ok": "true"})
	}
}

var _ ServerStore = (*store.Store)(nil)
