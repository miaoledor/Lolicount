package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestSanitizeBackslashEscapeCounter covers the milkdown/remark
// serialization bug where "&" is escaped as "\&" in markdown source,
// producing counter URLs like:
//
//	/@demo?theme=lian\&fsize=16\&scale=1\&unshowf=true\&mode=seq
//
// Without the sanitize middleware, fasthttp keeps the backslash in the
// value (theme="lian\", fsize="16\"), so theme fails themename
// validation and fsize fails int bind -> 400. The middleware rewrites
// "\&" -> "&" before binding, restoring the original intent.
func TestSanitizeBackslashEscapeCounter(t *testing.T) {
	s := newCounterServer(t)

	cases := []struct {
		name string
		url  string
	}{
		{
			name: "full buggy url",
			url:  "/@demo?theme=lian\\&fsize=16\\&scale=1\\&unshowf=true\\&mode=seq",
		},
		{
			name: "theme only with trailing backslash",
			url:  "/@demo?theme=lian\\&fsize=16",
		},
		{
			name: "random theme with backslash escape",
			url:  "/@demo?theme=random\\&mode=seq",
		},
		{
			name: "number param with backslash escape",
			url:  "/@demo?theme=lian\\&number=42",
		},
		{
			name: "position params x y with backslash escape",
			url:  "/@demo?theme=lian\\&x=10\\&y=20",
		},
		{
			name: "scale float with backslash escape",
			url:  "/@demo?theme=lian\\&scale=1.5\\&fsize=20",
		},
		{
			name: "single param trailing backslash only",
			url:  "/@demo?theme=lian\\&mode=random",
		},
		{
			name: "many params all backslash escaped",
			url:  "/@demo?theme=lian\\&fsize=16\\&scale=1\\&unshowf=true\\&mode=seq\\&x=0\\&y=0\\&rx=0.5\\&ry=0.5\\&number=7",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.url, nil)
			resp, err := s.app.Test(req)
			if err != nil {
				t.Fatalf("app.Test: %v", err)
			}
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("status: got %d want 200 (url=%s)", resp.StatusCode, tc.url)
			}
			if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "image/svg+xml") {
				t.Errorf("Content-Type: got %q want image/svg+xml", ct)
			}
		})
	}
}

// TestSanitizeBackslashEscapeGetAlias verifies the /get/@:name
// compatibility alias is also repaired.
func TestSanitizeBackslashEscapeGetAlias(t *testing.T) {
	s := newCounterServer(t)
	req := httptest.NewRequest(http.MethodGet, "/get/@demo?theme=lian\\&fsize=16\\&mode=seq", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want 200", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "image/svg+xml") {
		t.Errorf("Content-Type: got %q want image/svg+xml", ct)
	}
}

// TestSanitizeBackslashEscapeNoOpForCleanQuery ensures the middleware
// does not alter queries that have no backslash (the common path).
func TestSanitizeBackslashEscapeNoOpForCleanQuery(t *testing.T) {
	s := newCounterServer(t)
	req := httptest.NewRequest(http.MethodGet, "/@demo?theme=lian&fsize=16&scale=1&unshowf=true&mode=seq", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want 200", resp.StatusCode)
	}
}

// TestSanitizeBackslashEscapePreservesTheme verifies that after the
// middleware repairs the query, the requested theme is actually used
// (not silently falling back to the default). We compare the SVG body
// against the same request without backslashes — they must match.
func TestSanitizeBackslashEscapePreservesTheme(t *testing.T) {
	s := newCounterServer(t)

	clean := httptest.NewRequest(http.MethodGet, "/@demo?theme=lian&fsize=16&mode=seq", nil)
	respClean, err := s.app.Test(clean)
	if err != nil {
		t.Fatalf("clean app.Test: %v", err)
	}
	bodyClean := readBody(t, respClean)

	buggy := httptest.NewRequest(http.MethodGet, "/@demo?theme=lian\\&fsize=16\\&mode=seq", nil)
	respBuggy, err := s.app.Test(buggy)
	if err != nil {
		t.Fatalf("buggy app.Test: %v", err)
	}
	bodyBuggy := readBody(t, respBuggy)

	// Frame index for demo is always 0, and the rendered text is the
	// fixed "0123456789", so the two SVGs must be byte-identical when
	// the middleware correctly repairs the query.
	if bodyClean != bodyBuggy {
		t.Errorf("repaired query produced different SVG than clean query:\nclean=%q\nbuggy=%q",
			trunc(bodyClean, 120), trunc(bodyBuggy, 120))
	}
}

// TestSanitizeBackslashEscapeRecordRoute verifies the record alias is
// also repaired.
func TestSanitizeBackslashEscapeRecordRoute(t *testing.T) {
	s := newCounterServer(t)
	// First, increment via the counter so a record exists.
	req1 := httptest.NewRequest(http.MethodGet, "/@rec-bs?theme=lian", nil)
	if _, err := s.app.Test(req1); err != nil {
		t.Fatalf("increment app.Test: %v", err)
	}
	// Then read via /record with a backslash-escaped query.
	req2 := httptest.NewRequest(http.MethodGet, "/record/@rec-bs?theme=lian\\&fsize=16", nil)
	resp, err := s.app.Test(req2)
	if err != nil {
		t.Fatalf("record app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("record status: got %d want 200", resp.StatusCode)
	}
}

// TestSanitizeBackslashDoesNotTouchIsolatedBackslash confirms the
// middleware only rewrites the "\&" sequence, leaving an isolated
// backslash (not before &) untouched. An isolated backslash in a theme
// value is illegal and must still 400 — the middleware must not
// silently strip it and mask a genuinely invalid request.
func TestSanitizeBackslashDoesNotTouchIsolatedBackslash(t *testing.T) {
	s := newCounterServer(t)
	// theme=lian\ren — backslash is NOT before &, so it stays.
	// theme becomes "lian\ren", which fails themename validation.
	req := httptest.NewRequest(http.MethodGet, "/@demo?theme=lian\\ren&fsize=16", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("isolated backslash should still 400: got %d", resp.StatusCode)
	}
}

// TestSanitizeBackslashEscapeNoQuery verifies the middleware handles
// requests with no query string at all (the fast-path must not panic
// or misbehave on an empty query).
func TestSanitizeBackslashEscapeNoQuery(t *testing.T) {
	s := newCounterServer(t)
	req := httptest.NewRequest(http.MethodGet, "/@demo", nil)
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want 200", resp.StatusCode)
	}
}
