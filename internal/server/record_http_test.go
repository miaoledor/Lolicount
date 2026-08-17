package server

import (
	"context"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"

	"github.com/miaoledor/lolicount/internal/config"
	"github.com/miaoledor/lolicount/internal/counter"
	"github.com/miaoledor/lolicount/internal/store"
	"github.com/miaoledor/lolicount/internal/theme"
	"github.com/rs/zerolog"
)

// TestRecordAgreesOverRealHTTP is a regression test for the M3 bug where
// /record/@:name returned 0 (or a stale value) while /@:name kept
// incrementing correctly.
//
// Root cause: Fiber/fasthttp route params can reference a per-request
// buffer that the runtime reuses across requests. counterHandler stored
// that string as a map key inside counter.Buffer; the next request
// overwrote the buffer, corrupting the cached key (e.g. "d6" became
// "ec"), so the read-only /record lookup missed the cache and fell back
// to an empty store. The fix clones the param before caching.
//
// app.Test runs handlers synchronously on a single fasthttp context, so
// it cannot reproduce the buffer-reuse race; this test drives the server
// over a real TCP listener with an http.Client to exercise the same
// buffer-reuse path as production.
func TestRecordAgreesOverRealHTTP(t *testing.T) {
	th := &theme.Theme{Name: "loli", Frames: []theme.Frame{{Width: 10, Height: 20, Data: "data:image/gif;base64,QQ"}}}
	reg := &stubRegistry{themes: map[string]*theme.Theme{"loli": th}}

	repo, err := store.NewSQLite(context.Background(), ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if c, ok := repo.(interface{ Close() error }); ok {
			c.Close()
		}
	})
	buf := counter.New(repo, zerolog.Nop(), 3600)
	if err := buf.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(buf.Stop)

	cfg := &config.Config{Host: "127.0.0.1", Port: 0, DBInterval: 10, RateLimitIPPerSec: 10000, RateLimitIPPerMin: 100000, RateLimitNamePerSec: 10000}
	s := New(cfg, zerolog.Nop(), reg, buf, nil)
	t.Cleanup(func() {
		s.ipLimiter.Stop()
		s.nameLimiter.Stop()
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go s.app.Listener(ln)
	base := "http://" + ln.Addr().String()
	client := &http.Client{}

	// A fresh name: each incr then record must agree and increment.
	for i := 1; i <= 5; i++ {
		resp, err := client.Get(base + "/@regress?theme=loli")
		if err != nil {
			t.Fatalf("incr %d: %v", i, err)
		}
		io.ReadAll(resp.Body)
		resp.Body.Close()

		rec, err := client.Get(base + "/record/@regress")
		if err != nil {
			t.Fatalf("record %d: %v", i, err)
		}
		body, _ := io.ReadAll(rec.Body)
		rec.Body.Close()
		want := `"num":` + itoa(i)
		if !strings.Contains(string(body), want) {
			t.Fatalf("iter %d: record body %q does not contain %s (param-buffer reuse regression)", i, string(body), want)
		}
	}
}

// itoa is a small local int->string to avoid importing strconv just for
// one call site.
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b []byte
	for i > 0 {
		b = append([]byte{byte('0' + i%10)}, b...)
		i /= 10
	}
	return string(b)
}
