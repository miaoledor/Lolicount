package composer

import (
	"encoding/json"
	"io/fs"
	"runtime"
	"testing"
	"time"

	"github.com/miaoledor/lolicount/assets"
	"github.com/miaoledor/lolicount/internal/imgcore/asset"
)

// benchmarkSink keeps measured results reachable so the compiler cannot
// eliminate the catalog or registry construction whose cost is the subject
// of this harness. Theme counts are expected to grow, so no absolute memory
// value is ever asserted here: snapshots are logged for before/after review.
var benchmarkSink any

// memSnapshot is one stage of the startup path, with cumulative allocations
// (short-lived bytes included) and the live heap retained after a forced GC.
type memSnapshot struct {
	Stage      string  `json:"stage"`
	DurationMS float64 `json:"duration_ms"`
	AllocMB    float64 `json:"alloc_mb"`
	AllocCount uint64  `json:"alloc_count"`
	HeapLiveMB float64 `json:"heap_live_mb"`
	HeapSysMB  float64 `json:"heap_sys_mb"`
	SysMB      float64 `json:"sys_mb"`
	Catalogs   int     `json:"catalogs"`
	Themes     int     `json:"themes"`
}

// themeSubtree opens the same embedded theme tree the production registry
// uses, so every number reported here is comparable to a real server boot.
func themeSubtree(tb testing.TB) fs.FS {
	tb.Helper()
	fsys, err := fs.Sub(assets.FS, "theme")
	if err != nil {
		tb.Fatalf("open embedded theme tree: %v", err)
	}
	return fsys
}

// measureStage runs fn between two forced collections and reports both the
// total allocated bytes (which exposes full-file reads that are immediately
// discarded) and the live heap that survives fn's result.
func measureStage(stage string, fn func() (int, int)) memSnapshot {
	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	start := time.Now()
	catalogs, themes := fn()
	elapsed := time.Since(start)
	runtime.GC()
	runtime.ReadMemStats(&after)
	const mb = 1024 * 1024
	return memSnapshot{
		Stage:      stage,
		DurationMS: float64(elapsed.Microseconds()) / 1000,
		AllocMB:    float64(after.TotalAlloc-before.TotalAlloc) / mb,
		AllocCount: after.Mallocs - before.Mallocs,
		HeapLiveMB: float64(after.HeapAlloc) / mb,
		HeapSysMB:  float64(after.HeapSys) / mb,
		SysMB:      float64(after.Sys) / mb,
		Catalogs:   catalogs,
		Themes:     themes,
	}
}

// TestMemoryResidentOnly runs in its own process (go test -run
// TestMemoryResidentOnly) so HeapSys and Sys are not inflated by a previous
// full-assembly stage in the same test binary.
func TestMemoryResidentOnly(t *testing.T) {
	snap := measureStage("resident10-isolated", func() (int, int) {
		reg, errs := NewThemeRegistryWithResident(10)
		if len(errs) > 0 {
			t.Fatalf("registry errors: %v", errs[:min(len(errs), 3)])
		}
		benchmarkSink = reg
		return 0, len(reg.List())
	})
	encoded, err := json.Marshal(snap)
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}
	t.Logf("MEM %s", encoded)
}

// TestMemoryBaseline logs the allocation and live-heap cost of the catalog
// scan and the full registry assembly. It is a measurement, not a pass/fail
// budget: theme assets change frequently and must not break the suite.
func TestMemoryBaseline(t *testing.T) {
	fsys := themeSubtree(t)
	catalogs := measureStage("catalogs", func() (int, int) {
		cats, errs := asset.LoadThemeCatalogs(fsys)
		if len(errs) > 0 {
			t.Fatalf("catalog errors: %v", errs[:min(len(errs), 3)])
		}
		benchmarkSink = cats
		return len(cats), 0
	})
	registry := measureStage("registry", func() (int, int) {
		reg, errs := NewThemeRegistry()
		if len(errs) > 0 {
			t.Fatalf("registry errors: %v", errs[:min(len(errs), 3)])
		}
		benchmarkSink = reg
		return 0, len(reg.List())
	})
	resident := measureStage("resident10", func() (int, int) {
		reg, errs := NewThemeRegistryWithResident(10)
		if len(errs) > 0 {
			t.Fatalf("registry errors: %v", errs[:min(len(errs), 3)])
		}
		benchmarkSink = reg
		return 0, len(reg.List())
	})
	for _, snap := range []memSnapshot{catalogs, registry, resident} {
		encoded, err := json.Marshal(snap)
		if err != nil {
			t.Fatalf("marshal snapshot: %v", err)
		}
		t.Logf("MEM %s", encoded)
	}
}

// BenchmarkLoadThemeCatalogs measures the startup metadata scan: its B/op
// is the clearest signal for how many image bytes the catalog build touches.
func BenchmarkLoadThemeCatalogs(b *testing.B) {
	fsys := themeSubtree(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cats, errs := asset.LoadThemeCatalogs(fsys)
		if len(cats) == 0 || len(errs) > 0 {
			b.Fatalf("catalog scan failed: %d catalogs, %d errors", len(cats), len(errs))
		}
		benchmarkSink = cats
	}
}

// BenchmarkNewThemeRegistry measures the whole startup assembly, including
// catalog scanning and the per-slot images retained by the registry.
func BenchmarkNewThemeRegistry(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reg, errs := NewThemeRegistry()
		if len(reg.List()) == 0 || len(errs) > 0 {
			b.Fatalf("registry build failed: %d errors", len(errs))
		}
		benchmarkSink = reg
	}
}

// BenchmarkNewThemeRegistryWithResident measures the production startup
// path, where only the hot themes are assembled up front.
func BenchmarkNewThemeRegistryWithResident(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reg, errs := NewThemeRegistryWithResident(10)
		if len(reg.List()) == 0 || len(errs) > 0 {
			b.Fatalf("registry build failed: %d errors", len(errs))
		}
		benchmarkSink = reg
	}
}

// BenchmarkPrepareRenderCold measures the same "wenders" theme as
// BenchmarkPrepareRender, but with every assembly evicted first. That makes
// the on-demand assembly cost directly comparable with the resident swap.
func BenchmarkPrepareRenderCold(b *testing.B) {
	reg, errs := NewThemeRegistry()
	if len(errs) > 0 {
		b.Fatalf("registry errors: %v", errs[:min(len(errs), 3)])
	}
	reg.(ResidentController).Warm(nil)
	prep := reg.(RenderPreparer)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		snapshot, err := prep.PrepareRender("wenders")
		if err != nil {
			b.Fatalf("cold prepare: %v", err)
		}
		benchmarkSink = snapshot
	}
}

// BenchmarkPrepareRender measures one request-side swap plus the private
// snapshot handed to the SVG composer.
func BenchmarkPrepareRender(b *testing.B) {
	reg, errs := NewThemeRegistry()
	if len(errs) > 0 {
		b.Fatalf("registry errors: %v", errs[:min(len(errs), 3)])
	}
	prep := reg.(RenderPreparer)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		snapshot, err := prep.PrepareRender("wenders")
		if err != nil {
			b.Fatalf("prepare: %v", err)
		}
		benchmarkSink = snapshot
	}
}
