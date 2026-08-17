package counter

import (
	"context"
	"testing"

	"github.com/rs/zerolog"

	"github.com/miaoledor/lolicount/internal/store"
)

// Reproduce: preload a value, Incr once, then Get must return the new value.
func TestDebugPreloadIncrGet(t *testing.T) {
	repo, err := store.NewSQLite(context.Background(), ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	// Seed a value.
	repo.Set(context.Background(), "persist", 1)

	b := New(repo, zerolog.Nop(), 3600)
	if err := b.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer b.Stop()

	v, _ := b.Incr(context.Background(), "persist")
	t.Logf("Incr returned %d (want 2)", v)
	g, _ := b.Get(context.Background(), "persist")
	t.Logf("Get returned %d (want 2)", g)
	if g != 2 {
		t.Errorf("Get = %d want 2", g)
	}
}
