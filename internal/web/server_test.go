package web_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/oenexa/oenexa/internal/web"
)

func TestAPIStatus_ReturnsJSON(t *testing.T) {
	h := web.Handler()
	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET /api/status: want 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("Content-Type: want application/json, got %q", contentType)
	}

	var resp web.StatusResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if resp.Network != "mainnet" {
		t.Errorf("want network mainnet, got %q", resp.Network)
	}
}

func TestAPIStatus_WithNodeInfo(t *testing.T) {
	info := web.NodeInfo{
		Version:     "v1.2.3",
		Network:     "testnet",
		ChainHeight: 42,
		PeerCount:   5,
	}

	h := web.ServeWithInfo(info)
	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	var resp web.StatusResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if resp.Version != "v1.2.3" || resp.Network != "testnet" || resp.ChainHeight != 42 || resp.PeerCount != 5 {
		t.Errorf("NodeInfo mismatch. Got: %+v", resp.NodeInfo)
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

func TestCORSHeaders_APIStatus(t *testing.T) {
	h := web.Handler()
	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("missing CORS header")
	}
}
