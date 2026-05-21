package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type Server struct {
	ID        string
	Name      string
	Host      string
	RCONPass  string
	CreatedAt time.Time
}

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path+"?_journal=WAL&_timeout=5000&_fk=true")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS servers (
			id         TEXT PRIMARY KEY,
			name       TEXT NOT NULL,
			host       TEXT NOT NULL,
			rcon_pass  TEXT NOT NULL,
			created_at INTEGER NOT NULL
		);

		CREATE TABLE IF NOT EXISTS sessions (
			token      TEXT PRIMARY KEY,
			created_at INTEGER NOT NULL,
			expires_at INTEGER NOT NULL
		);
	`)
	return err
}

func (s *Store) ListServers() ([]Server, error) {
	rows, err := s.db.Query(
		`SELECT id, name, host, rcon_pass, created_at FROM servers ORDER BY created_at ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Server
	for rows.Next() {
		var sv Server
		var ts int64
		if err := rows.Scan(&sv.ID, &sv.Name, &sv.Host, &sv.RCONPass, &ts); err != nil {
			return nil, err
		}
		sv.CreatedAt = time.Unix(ts, 0)
		out = append(out, sv)
	}
	return out, rows.Err()
}

func (s *Store) GetServer(id string) (Server, error) {
	var sv Server
	var ts int64
	err := s.db.QueryRow(
		`SELECT id, name, host, rcon_pass, created_at FROM servers WHERE id = ?`, id,
	).Scan(&sv.ID, &sv.Name, &sv.Host, &sv.RCONPass, &ts)
	if errors.Is(err, sql.ErrNoRows) {
		return Server{}, ErrNotFound
	}
	if err != nil {
		return Server{}, err
	}
	sv.CreatedAt = time.Unix(ts, 0)
	return sv, nil
}

func (s *Store) AddServer(sv Server) error {
	_, err := s.db.Exec(
		`INSERT INTO servers (id, name, host, rcon_pass, created_at) VALUES (?, ?, ?, ?, ?)`,
		sv.ID, sv.Name, sv.Host, sv.RCONPass, sv.CreatedAt.Unix(),
	)
	if err != nil && strings.Contains(err.Error(), "UNIQUE constraint") {
		return ErrConflict
	}
	return err
}

func (s *Store) DeleteServer(id string) error {
	_, err := s.db.Exec(`DELETE FROM servers WHERE id = ?`, id)
	return err
}

func (s *Store) CreateSession(token string, ttl time.Duration) error {
	now := time.Now()
	_, err := s.db.Exec(
		`INSERT INTO sessions (token, created_at, expires_at) VALUES (?, ?, ?)`,
		token, now.Unix(), now.Add(ttl).Unix(),
	)
	return err
}

func (s *Store) ValidSession(token string) (bool, error) {
	_, _ = s.db.Exec(`DELETE FROM sessions WHERE expires_at < ?`, time.Now().Unix())

	var count int
	err := s.db.QueryRow(
		`SELECT COUNT(*) FROM sessions WHERE token = ? AND expires_at >= ?`,
		token, time.Now().Unix(),
	).Scan(&count)
	return count > 0, err
}

func (s *Store) DeleteSession(token string) error {
	_, err := s.db.Exec(`DELETE FROM sessions WHERE token = ?`, token)
	return err
}

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("already exists")
)
