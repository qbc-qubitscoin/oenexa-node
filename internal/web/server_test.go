package web_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/oenexa/oenexa/internal/web"
)

// ─── helpers ────────────────────────────────────────────────────────────────

func newGet(t *testing.T, h http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// ─── Handler() backward-compat ───────────────────────────────────────────────

func TestHandler_RootReturnsIndex(t *testing.T) {
	h := web.Handler()
	if h == nil {
		t.Fatal("web.Handler returned nil")
	}

	rec := newGet(t, h, "/")

	if rec.Code != http.StatusOK {
		t.Errorf("GET /: want 200, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("Content-Type: want text/html, got %q", ct)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "OENEXA") {
		t.Errorf("body missing 'OENEXA': got %.200s", body)
	}
}

func TestHandler_ExplicitIndexHtml(t *testing.T) {
	h := web.Handler()

	rec := newGet(t, h, "/index.html")

	// Go's http.FileServer redirects /index.html → / (301 Moved Permanently).
	if rec.Code != http.StatusMovedPermanently && rec.Code != http.StatusOK {
		t.Errorf("GET /index.html: want 301 or 200, got %d", rec.Code)
	}
}

func TestHandler_HeadMethodSupported(t *testing.T) {
	h := web.Handler()

	req := httptest.NewRequest(http.MethodHead, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("HEAD /: want 200, got %d", rec.Code)
	}
}

func TestHandler_PostMethodNotAllowed(t *testing.T) {
	h := web.Handler()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"test":1}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST /: want 405, got %d", rec.Code)
	}
}

// ─── /api/status ─────────────────────────────────────────────────────────────

func TestAPIStatus_ReturnsJSON(t *testing.T) {
	h := web.Handler()
	rec := newGet(t, h, "/api/status")

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/status: want 200, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type: want application/json, got %q", ct)
	}

	var payload map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("body is not valid JSON: %v", err)
	}

	for _, key := range []string{"version", "network", "chain_height", "peer_count"} {
		if _, ok := payload[key]; !ok {
			t.Errorf("JSON response missing key %q", key)
		}
	}
}

func TestAPIStatus_WithNodeInfo(t *testing.T) {
	info := web.NodeInfo{
		Version:     "v1.2.3",
		Network:     "testnet",
		ChainHeight: 42,
		PeerCount:   7,
		PendingTx:   3,
		StateRoot:   "0xdeadbeef",
	}

	h := web.ServeWithInfo(info)
	rec := newGet(t, h, "/api/status")

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/status: want 200, got %d", rec.Code)
	}

	var payload map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("body is not valid JSON: %v", err)
	}

	if got := payload["version"]; got != "v1.2.3" {
		t.Errorf("version: want v1.2.3, got %v", got)
	}
	if got := payload["network"]; got != "testnet" {
		t.Errorf("network: want testnet, got %v", got)
	}
	// JSON numbers decode to float64 by default.
	if got := payload["chain_height"]; got != float64(42) {
		t.Errorf("chain_height: want 42, got %v", got)
	}
	if got := payload["peer_count"]; got != float64(7) {
		t.Errorf("peer_count: want 7, got %v", got)
	}
	if got := payload["state_root"]; got != "0xdeadbeef" {
		t.Errorf("state_root: want 0xdeadbeef, got %v", got)
	}
}

func TestAPIStatus_PostNotAllowed(t *testing.T) {
	h := web.Handler()

	req := httptest.NewRequest(http.MethodPost, "/api/status", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST /api/status: want 405, got %d", rec.Code)
	}
}

// ─── SPA fallback ────────────────────────────────────────────────────────────

func TestSPAFallback_DeepRoute(t *testing.T) {
	h := web.Handler()
	rec := newGet(t, h, "/some/deep/spa/route")

	if rec.Code != http.StatusOK {
		t.Errorf("GET /some/deep/spa/route: want 200 (SPA fallback), got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("SPA fallback Content-Type: want text/html, got %q", ct)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "OENEXA") {
		t.Errorf("SPA fallback body missing 'OENEXA'")
	}
}

func TestSPAFallback_UnknownFile(t *testing.T) {
	h := web.Handler()
	// A path that looks like a non-existent JS file — SPA handler should still
	// return index.html (200) rather than 404, because any non-asset path is
	// treated as an SPA route.
	rec := newGet(t, h, "/nonexistent-page")

	if rec.Code != http.StatusOK {
		t.Errorf("GET /nonexistent-page: want 200 (SPA fallback), got %d", rec.Code)
	}
}

// ─── Cache-Control headers ───────────────────────────────────────────────────

func TestCacheControl_JsFile(t *testing.T) {
	h := web.Handler()
	// The embedded assets directory contains a .js file; request it directly.
	rec := newGet(t, h, "/assets/index-wfj7xKuT.js")

	// The file exists, so we expect 200 with a caching header.
	if rec.Code != http.StatusOK {
		t.Fatalf("GET JS asset: want 200, got %d", rec.Code)
	}

	cc := rec.Header().Get("Cache-Control")
	if !strings.Contains(cc, "max-age") {
		t.Errorf("Cache-Control for .js: expected max-age, got %q", cc)
	}
}

func TestCacheControl_CssFile(t *testing.T) {
	h := web.Handler()
	rec := newGet(t, h, "/assets/index-DctOWvX0.css")

	if rec.Code != http.StatusOK {
		t.Fatalf("GET CSS asset: want 200, got %d", rec.Code)
	}

	cc := rec.Header().Get("Cache-Control")
	if !strings.Contains(cc, "max-age") {
		t.Errorf("Cache-Control for .css: expected max-age, got %q", cc)
	}
}

func TestCacheControl_HtmlNoCache(t *testing.T) {
	h := web.Handler()
	rec := newGet(t, h, "/")

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /: want 200, got %d", rec.Code)
	}

	cc := rec.Header().Get("Cache-Control")
	if !strings.Contains(cc, "no-store") {
		t.Errorf("Cache-Control for HTML: expected no-store, got %q", cc)
	}
}

// ─── CORS headers ────────────────────────────────────────────────────────────

func TestCORSHeaders_StaticFile(t *testing.T) {
	h := web.Handler()
	rec := newGet(t, h, "/")

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("CORS Allow-Origin: want *, got %q", got)
	}
}

func TestCORSHeaders_APIStatus(t *testing.T) {
	h := web.Handler()
	rec := newGet(t, h, "/api/status")

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("CORS Allow-Origin on /api/status: want *, got %q", got)
	}
}
