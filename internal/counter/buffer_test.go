package counter

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/miaoledor/lolicount/internal/store"
)

// memRepo is an in-memory Repository for buffer tests, with a hook to
// make SetMulti fail on demand.
type memRepo struct {
	mu       sync.Mutex
	data     map[string]int64
	failNext bool
	sets     int
}

func newMemRepo() *memRepo {
	return &memRepo{data: make(map[string]int64)}
}

func (m *memRepo) Get(_ context.Context, name string) (store.Counter, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.data[name]
	return store.Counter{Name: name, Num: v}, ok, nil
}

func (m *memRepo) GetAll(_ context.Context) ([]store.Counter, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]store.Counter, 0, len(m.data))
	for k, v := range m.data {
		out = append(out, store.Counter{Name: k, Num: v})
	}
	return out, nil
}

func (m *memRepo) Set(_ context.Context, name string, num int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[name] = num
	return nil
}

func (m *memRepo) SetMulti(_ context.Context, items []store.Counter) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sets++
	if m.failNext {
		m.failNext = false
		return errFailed
	}
	for _, it := range items {
		m.data[it.Name] = it.Num
	}
	return nil
}

var errFailed = errFailedErr{}

type errFailedErr struct{}

func (errFailedErr) Error() string { return "simulated SetMulti failure" }

func newTestBuffer(t *testing.T, repo store.Repository) *Buffer {
	t.Helper()
	return New(repo, zerolog.Nop(), 1)
}

func TestIncrAndGet(t *testing.T) {
	repo := newMemRepo()
	b := newTestBuffer(t, repo)
	if err := b.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer b.Stop()

	for i := int64(1); i <= 5; i++ {
		v, err := b.Incr(context.Background(), "a")
		if err != nil {
			t.Fatalf("Incr: %v", err)
		}
		if v != i {
			t.Errorf("Incr a = %d want %d", v, i)
		}
	}
	got, _ := b.Get(context.Background(), "a")
	if got != 5 {
		t.Errorf("Get a = %d want 5", got)
	}
}

func TestFlushPersists(t *testing.T) {
	repo := newMemRepo()
	b := newTestBuffer(t, repo)
	b.Start(context.Background())
	b.Incr(context.Background(), "x")
	b.Incr(context.Background(), "x")
	b.Incr(context.Background(), "y")
	// flush is private; trigger via Stop (final flush).
	b.Stop()

	if repo.data["x"] != 2 || repo.data["y"] != 1 {
		t.Errorf("after flush: x=%d y=%d want 2,1", repo.data["x"], repo.data["y"])
	}
}

func TestFlushFailurePreservesCache(t *testing.T) {
	repo := newMemRepo()
	b := newTestBuffer(t, repo)
	b.Start(context.Background())

	b.Incr(context.Background(), "a") // cache a=1
	repo.failNext = true
	// Trigger a flush that will fail. Use the private flush via a manual
	// call path: we can't call flush directly, so stop+restart the loop
	// is awkward. Instead drive flush by reaching in through the unexported
	// method (same package).
	if err := b.flush(context.Background()); err == nil {
		t.Fatal("expected flush to fail")
	}
	// After failed flush, a should still be 1 in cache (merged back).
	got, _ := b.Get(context.Background(), "a")
	if got != 1 {
		t.Errorf("after failed flush Get a = %d want 1", got)
	}
	// A subsequent Incr should continue from 1.
	v, _ := b.Incr(context.Background(), "a")
	if v != 2 {
		t.Errorf("Incr a after fail = %d want 2", v)
	}
	b.Stop()
}

// Concurrent increments to the same name must not lose any: the final
// value equals the number of Incr calls.
func TestConcurrentIncrNoLoss(t *testing.T) {
	repo := newMemRepo()
	b := newTestBuffer(t, repo)
	b.Start(context.Background())
	defer b.Stop()

	const n = 1000
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			b.Incr(context.Background(), "c")
		}()
	}
	wg.Wait()
	got, _ := b.Get(context.Background(), "c")
	if got != n {
		t.Errorf("got %d want %d", got, n)
	}
}

// flush during a concurrent Incr must not lose the in-flight increment.
func TestFlushDuringIncrNoLoss(t *testing.T) {
	repo := newMemRepo()
	b := newTestBuffer(t, repo)
	b.Start(context.Background())

	const n = 500
	var wg sync.WaitGroup
	wg.Add(n + 1)
	go func() {
		defer wg.Done()
		// flush repeatedly while increments happen.
		for i := 0; i < 20; i++ {
			b.flush(context.Background())
			time.Sleep(time.Millisecond)
		}
	}()
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			b.Incr(context.Background(), "d")
		}()
	}
	wg.Wait()
	b.Stop() // final flush before reading, so cache+repo are consistent
	got, _ := b.Get(context.Background(), "d")
	if got != n {
		t.Errorf("got %d want %d", got, n)
	}
}

// Over capacity, new names degrade to read-only (no increment).
func TestCapacityDegradeReadOnly(t *testing.T) {
	repo := newMemRepo()
	b := newTestBuffer(t, repo)
	b.Start(context.Background())
	defer b.Stop()

	// Fill cache to capacity.
	for i := 0; i < maxCacheEntries; i++ {
		b.Incr(context.Background(), fmt.Sprintf("n%d", i))
	}
	// A pre-existing name still increments.
	v, _ := b.Incr(context.Background(), "n0")
	if v != 2 {
		t.Errorf("n0 second incr = %d want 2", v)
	}
	// A brand-new name over capacity degrades: returns 0, not 1.
	v, _ = b.Incr(context.Background(), "brand-new")
	if v != 0 {
		t.Errorf("over-cap new name = %d want 0 (read-only)", v)
	}
}

