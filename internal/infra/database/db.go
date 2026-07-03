package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"

	"wood-hot-monitor/ent"
)

type DB struct {
	Client *ent.Client
	conn   *sql.DB
}

func New(dataDir string) (*DB, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}

	dsn := filepath.Join(dataDir, "wood-hot-monitor.db") + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)"
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	conn.SetMaxOpenConns(4)

	drv := entsql.OpenDB(dialect.SQLite, conn)
	client := ent.NewClient(ent.Driver(drv))

	if err := client.Schema.Create(context.Background()); err != nil {
		conn.Close()
		return nil, fmt.Errorf("create schema: %w", err)
	}

	return &DB{Client: client, conn: conn}, nil
}

func (db *DB) Close() error {
	if db.Client != nil {
		return db.Client.Close()
	}
	return nil
}
