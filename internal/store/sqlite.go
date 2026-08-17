package store

import (
	"context"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite" // pure-Go SQLite driver, no CGO
)

const createTableSQL = `
CREATE TABLE IF NOT EXISTS tb_count (
    id    INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
    name  VARCHAR(32) NOT NULL UNIQUE,
    num   BIGINT      NOT NULL DEFAULT 0
);
`

// upsertSQL inserts a row or, on name conflict, overwrites num with the
// new absolute value. The UNIQUE constraint on name is both the conflict
// target and the guarantee against duplicate rows (AGENTS.md).
const upsertSQL = `
INSERT INTO tb_count (name, num) VALUES (?, ?)
ON CONFLICT(name) DO UPDATE SET num = excluded.num;
`

// sqliteRepo is the sole Repository implementation. It wraps a *sql.DB
// holding one open connection pool to the SQLite file.
type sqliteRepo struct {
	db *sql.DB
}

// NewSQLite opens (or creates) the SQLite database at dsn, ensures the
// tb_count table exists, and returns a Repository. dsn is a file path
// (e.g. "data/count.db"); use ":memory:" for tests.
func NewSQLite(ctx context.Context, dsn string) (Repository, error) {
	// _pragma keys tune SQLite for single-writer safety without CGO.
	dsn = dsn + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("store: open %s: %w", dsn, err)
	}
	// Single writer: SQLite serializes writes; one connection avoids
	// "database is locked" under the batched upsert path.
	db.SetMaxOpenConns(1)
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("store: ping %s: %w", dsn, err)
	}
	if _, err := db.ExecContext(ctx, createTableSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("store: create table: %w", err)
	}
	return &sqliteRepo{db: db}, nil
}

// Close releases the underlying database handle.
func (r *sqliteRepo) Close() error {
	return r.db.Close()
}

// Get returns the counter for name, or false if no row exists.
func (r *sqliteRepo) Get(ctx context.Context, name string) (Counter, bool, error) {
	var c Counter
	err := r.db.QueryRowContext(ctx, "SELECT name, num FROM tb_count WHERE name = ?", name).
		Scan(&c.Name, &c.Num)
	if err == sql.ErrNoRows {
		return Counter{}, false, nil
	}
	if err != nil {
		return Counter{}, false, fmt.Errorf("store: get %s: %w", name, err)
	}
	return c, true, nil
}

// GetAll returns every counter row, ordered by name for stable output.
func (r *sqliteRepo) GetAll(ctx context.Context) ([]Counter, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT name, num FROM tb_count ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("store: get all: %w", err)
	}
	defer rows.Close()
	var out []Counter
	for rows.Next() {
		var c Counter
		if err := rows.Scan(&c.Name, &c.Num); err != nil {
			return nil, fmt.Errorf("store: scan: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// Set upserts a single counter to the absolute value num.
func (r *sqliteRepo) Set(ctx context.Context, name string, num int64) error {
	_, err := r.db.ExecContext(ctx, upsertSQL, name, num)
	if err != nil {
		return fmt.Errorf("store: set %s: %w", name, err)
	}
	return nil
}

// SetMulti upserts multiple counters in one transaction. Each item is an
// absolute value overwrite (AGENTS.md Iron Rule 5: SetMulti does absolute
// value coverage, not increments).
func (r *sqliteRepo) SetMulti(ctx context.Context, items []Counter) error {
	if len(items) == 0 {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("store: begin tx: %w", err)
	}
	stmt, err := tx.PrepareContext(ctx, upsertSQL)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("store: prepare upsert: %w", err)
	}
	defer stmt.Close()
	for _, it := range items {
		if _, err := stmt.ExecContext(ctx, it.Name, it.Num); err != nil {
			tx.Rollback()
			return fmt.Errorf("store: upsert %s: %w", it.Name, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("store: commit: %w", err)
	}
	return nil
}
