package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/codevski/defuse/assets"
	"github.com/codevski/defuse/internal/auth"
	"github.com/codevski/defuse/internal/rcon"
	"github.com/codevski/defuse/internal/server"
	"github.com/codevski/defuse/internal/store"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	creds, err := auth.LoadCredentials()
	if err != nil {
		logger.Error("credentials", "error", err)
		os.Exit(1)
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "nokit.db"
	}
	st, err := store.Open(dbPath)
	if err != nil {
		logger.Error("store", "error", err)
		os.Exit(1)
	}
	defer st.Close()

	mgr := rcon.New(logger)
	servers, err := st.ListServers()
	if err != nil {
		logger.Error("list servers", "error", err)
		os.Exit(1)
	}
	for _, sv := range servers {
		if err := mgr.Connect(sv.ID, sv.Host, sv.RCONPass); err != nil {
			logger.Warn("rcon connect", "server", sv.ID, "error", err)
		} else {
			logger.Info("rcon connected", "server", sv.ID, "host", sv.Host)
		}
	}

	dist, err := assets.FS()
	if err != nil {
		logger.Error("assets", "error", err)
		os.Exit(1)
	}

	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	srv := server.New(logger, dist, st, mgr, creds)
	httpSrv := &http.Server{
		Addr:    addr,
		Handler: srv.Handler(),
	}

	go func() {
		logger.Info("listening", "addr", addr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("listen", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpSrv.Shutdown(ctx); err != nil {
		logger.Error("shutdown", "error", err)
	}
	logger.Info("shutdown complete")
}
