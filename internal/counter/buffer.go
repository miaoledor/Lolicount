// Package counter holds the in-memory counter buffer that fronts SQLite.
// Storage path (AGENTS.md Iron Rule 5, the only path):
// request -> Buffer.Incr (in-memory map) -> time.Ticker -> flush ->
// store.SetMulti (batched upsert) -> SQLite.
package counter

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"github.com/miaoledor/lolicount/internal/store"
)

// maxCacheEntries guards memory under extreme cardinality: once more
// than this many distinct names are buffered, new names degrade to
// read-only (served from DB) instead of being cached (AGENTS.md).
const maxCacheEntries = 10000

// Buffer is the in-memory counter fronting SQLite. It keeps the current
// count for each name in a map so the hot path (Incr) never touches the
// DB; a ticker flushes snapshots to the store in batches.
type Buffer struct {
	repo   store.Repository
	logger zerolog.Logger

	mu    sync.Mutex
	cache map[string]int64 // name -> current count (absolute)

	tickInterval time.Duration
	stop         chan struct{}
	done         chan struct{}
}

// New creates a Buffer backed by repo. dbIntervalSec is the flush period
// in seconds (DB_INTERVAL). Call Start to begin the ticker, Stop to
// flush once and halt.
func New(repo store.Repository, logger zerolog.Logger, dbIntervalSec int) *Buffer {
	if dbIntervalSec < 1 {
		dbIntervalSec = 1
	}
	return &Buffer{
		repo:         repo,
		logger:       logger,
		cache:        make(map[string]int64),
		tickInterval: time.Duration(dbIntervalSec) * time.Second,
		stop:         make(chan struct{}),
		done:         make(chan struct{}),
	}
}

// Start launches the background flush ticker. It loads existing counts
// from the store so a restart resumes from the persisted values.
func (b *Buffer) Start(ctx context.Context) error {
	all, err := b.repo.GetAll(ctx)
	if err != nil {
		return err
	}
	b.mu.Lock()
	for _, c := range all {
		b.cache[c.Name] = c.Num
	}
	b.mu.Unlock()
	b.logger.Info().Int("loaded", len(b.cache)).Msg("counter buffer loaded")

	go b.loop()
	return nil
}

// loop runs the periodic flush until Stop is called.
func (b *Buffer) loop() {
	ticker := time.NewTicker(b.tickInterval)
	defer ticker.Stop()
	defer close(b.done)
	for {
		select {
		case <-ticker.C:
			if err := b.flush(context.Background()); err != nil {
				b.logger.Error().Err(err).Msg("counter flush failed")
			}
		case <-b.stop:
			// Final flush on shutdown to minimize the loss window.
			if err := b.flush(context.Background()); err != nil {
				b.logger.Error().Err(err).Msg("counter final flush failed")
			}
			return
		}
	}
}

// Stop signals the flush loop to do a final flush and exit, then waits.
func (b *Buffer) Stop() {
	close(b.stop)
	<-b.done
}

// Incr increments the counter for name by 1 and returns the new value.
// If the buffer is at capacity and name is not already cached, it
// degrades to read-only: the current value is served from the store
// WITHOUT incrementing (AGENTS.md Iron Rule 3-style degradation, here
// for memory protection).
func (b *Buffer) Incr(ctx context.Context, name string) (int64, error) {
	b.mu.Lock()
	if _, ok := b.cache[name]; !ok && len(b.cache) >= maxCacheEntries {
		// Over capacity for a new name: degrade read-only.
		b.mu.Unlock()
		b.logger.Warn().
			Str("name", name).
			Int("cache_size", len(b.cache)).
			Msg("counter buffer full, degrading new name to read-only")
		c, ok, err := b.repo.Get(ctx, name)
		if err != nil {
			return 0, err
		}
		if !ok {
			return 0, nil
		}
		return c.Num, nil
	}
	b.cache[name]++
	val := b.cache[name]
	b.mu.Unlock()
	return val, nil
}

// Get returns the current count for name: from the buffer if cached,
// otherwise from the store.
func (b *Buffer) Get(ctx context.Context, name string) (int64, error) {
	b.mu.Lock()
	if v, ok := b.cache[name]; ok {
		b.mu.Unlock()
		return v, nil
	}
	b.mu.Unlock()
	c, ok, err := b.repo.Get(ctx, name)
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, nil
	}
	return c.Num, nil
}

// flush snapshots the cache and upserts it to the store. The cache is
// NOT cleared: it holds absolute counts (AGENTS.md: "内存维护当前计数"),
// so each name's baseline must persist in memory for subsequent Incr to
// build on. Clearing/swapping the map (the Moe-Counter increment-delta
// model) would reset a name to 0 after its first flush and lose the
// baseline — exactly the bug M3 is meant to fix. SetMulti does absolute
// value overwrite, so re-pushing the same growing value each tick is
// correct and idempotent.
//
// Increments arriving during the in-flight SetMulti mutate the live
// cache directly; they are included in the NEXT flush, never lost.
func (b *Buffer) flush(ctx context.Context) error {
	b.mu.Lock()
	if len(b.cache) == 0 {
		b.mu.Unlock()
		return nil
	}
	items := make([]store.Counter, 0, len(b.cache))
	for name, num := range b.cache {
		items = append(items, store.Counter{Name: name, Num: num})
	}
	b.mu.Unlock()

	if err := b.repo.SetMulti(ctx, items); err != nil {
		// Nothing to merge back: the cache was never mutated by flush.
		// The snapshot values are still in cache (possibly larger now
		// due to concurrent Incr), and the next flush will retry.
		return err
	}
	return nil
}
