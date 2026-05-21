package api

import (
	"net/http"

	"github.com/codevski/defuse/internal/rcon"
)

type RCONRequest struct {
	Command string `json:"command"`
}

type RCONResponse struct {
	Output string `json:"output"`
}

func RCONHandler(mgr *rcon.Manager) Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		id := r.PathValue("id")
		if id == "" {
			return BadRequest("missing server id")
		}
		var req RCONRequest
		if err := Decode(r, &req); err != nil {
			return err
		}
		if req.Command == "" {
			return BadRequest("command is required")
		}
		out, err := mgr.Execute(id, req.Command)
		if err != nil {
			return WrapHTTP(err, http.StatusBadGateway, "rcon error")
		}
		return JSON(w, http.StatusOK, RCONResponse{Output: out})
	}
}
