// Package store provides the persistence layer for counters. The only
// implementation is sqliteRepo (modernc.org/sqlite, pure Go, no CGO).
// Business code depends on the Repository interface, never on the
// concrete type (AGENTS.md: store.Repository is the interface, sqliteRepo
// is the sole implementation).
package store

import "context"

// Counter is one row of tb_count: a name and its current value.
type Counter struct {
	Name string
	Num  int64
}

// Repository is the persistence interface for counters. All methods are
// safe for concurrent use within a single process; SetMulti is the
// batched upsert path used by counter.Buffer (AGENTS.md Iron Rule 5).
type Repository interface {
	// Get returns the counter for name, or false if absent.
	Get(ctx context.Context, name string) (Counter, bool, error)
	// GetAll returns every counter row.
	GetAll(ctx context.Context) ([]Counter, error)
	// Set upserts a single counter to the absolute value num.
	Set(ctx context.Context, name string, num int64) error
	// SetMulti upserts multiple counters in one transaction. Each entry
	// is an absolute value overwrite (NOT an increment): the buffer
	// owns the current count and pushes its snapshot here.
	SetMulti(ctx context.Context, items []Counter) error
}
