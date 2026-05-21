// Package rcon manages connections to CS2 servers over the Source RCON
// protocol. The Manager owns a pool of clients keyed by server ID and
// handles reconnection on failure.
package rcon

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/gorcon/rcon"
)

type Client interface {
	Execute(command string) (string, error)
	Close() error
}

var ErrNotConnected = errors.New("rcon: not connected")

type connection struct {
	id       string
	host     string
	password string

	mu     sync.Mutex
	client Client
}

type Manager struct {
	logger *slog.Logger

	mu          sync.RWMutex
	connections map[string]*connection
}

func New(logger *slog.Logger) *Manager {
	return &Manager{
		logger:      logger,
		connections: make(map[string]*connection),
	}
}

func (m *Manager) Connect(serverID, host, password string) error {
	conn := &connection{
		id:       serverID,
		host:     host,
		password: password,
	}
	m.mu.Lock()
	m.connections[serverID] = conn
	m.mu.Unlock()
	if err := m.dial(conn); err != nil {
		m.logger.Warn("rcon initial connect failed",
			"server", serverID,
			"host", host,
			"error", err,
		)
		return err
	}
	return nil
}

func (m *Manager) dial(conn *connection) error {
	c, err := rcon.Dial(conn.host, conn.password,
		rcon.SetDialTimeout(5*time.Second),
		rcon.SetDeadline(10*time.Second),
	)
	if err != nil {
		return fmt.Errorf("rcon dial %s: %w", conn.host, err)
	}

	conn.mu.Lock()
	if conn.client != nil {
		_ = conn.client.Close()
	}
	conn.client = c
	conn.mu.Unlock()

	m.logger.Info("rcon connected", "server", conn.id, "host", conn.host)
	return nil
}

func (m *Manager) Execute(serverID, command string) (string, error) {
	m.mu.RLock()
	conn, ok := m.connections[serverID]
	m.mu.RUnlock()
	if !ok {
		return "", fmt.Errorf("rcon: unknown server %q", serverID)
	}

	conn.mu.Lock()
	client := conn.client
	conn.mu.Unlock()

	if client != nil {
		out, err := client.Execute(command)
		if err == nil {
			return out, nil
		}
		m.logger.Warn("rcon execute failed, retrying",
			"server", serverID,
			"error", err,
		)
		conn.mu.Lock()
		_ = conn.client.Close()
		conn.client = nil
		conn.mu.Unlock()
	}

	if err := m.dial(conn); err != nil {
		return "", fmt.Errorf("%w: %v", ErrNotConnected, err)
	}

	conn.mu.Lock()
	client = conn.client
	conn.mu.Unlock()

	return client.Execute(command)
}

func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, conn := range m.connections {
		conn.mu.Lock()
		if conn.client != nil {
			if err := conn.client.Close(); err != nil {
				m.logger.Warn("rcon close", "server", id, "error", err)
			}
			conn.client = nil
		}
		conn.mu.Unlock()
	}
	return nil
}

func (m *Manager) Disconnect(serverID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	conn, ok := m.connections[serverID]
	if !ok {
		return
	}
	conn.mu.Lock()
	if conn.client != nil {
		_ = conn.client.Close()
		conn.client = nil
	}
	conn.mu.Unlock()
	delete(m.connections, serverID)
}
