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
	"github.com/miaoledor/lolicount/internal/imgcore/asset"
	"github.com/miaoledor/lolicount/internal/store"
	"github.com/rs/zerolog"
)

// TestRecordAgreesOverRealHTTP is a regression test for the M3 bug where
// /record/@:name returned 0 (or a stale value) while /@:name kept
// incrementing correctly. It drives the server over a real TCP listener
// to exercise the fasthttp param-buffer reuse path.
func TestRecordAgreesOverRealHTTP(t *testing.T) {
	th := &asset.CardTheme{Name: "lian", Frames: makeCardFrames(1)}
	reg := &stubRegistry{cards: map[string]*asset.CardTheme{"lian": th}}

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
	s := New(cfg, zerolog.Nop(), reg, nil, buf)
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

	for i := 1; i <= 5; i++ {
		resp, err := client.Get(base + "/@regress?theme=lian")
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
