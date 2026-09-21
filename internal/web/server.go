// Package web provides the embedded web dashboard and REST API for the OENEXA node.
//
// It serves a self-contained HTML dashboard from the embedded filesystem, exposes a
// /api/status JSON endpoint, and implements SPA (Single Page Application) fallback
// routing so that React Router deep-links resolve to index.html instead of 404.
package web

import (
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"strings"
	"time"
)

// staticFS embeds everything inside the static/ directory at compile time.
//
//go:embed static
var staticFS embed.FS

// NodeInfo carries live operational data from the running node that is
// surfaced via the /api/status endpoint.
type NodeInfo struct {
	// Version is the semantic version string of the node binary, e.g. "v0.9.0".
	Version string `json:"version"`

	// Network identifies the chain network, e.g. "mainnet" or "testnet".
	Network string `json:"network"`

	// ChainHeight is the height of the latest committed block.
	ChainHeight uint64 `json:"chain_height"`

	// PeerCount is the number of currently connected peers.
	PeerCount int `json:"peer_count"`

	// PendingTx is the number of transactions currently waiting in the mempool.
	PendingTx int `json:"pending_tx"`

	// StateRoot is the hex-encoded Sparse Merkle Trie root of the current state.
	StateRoot string `json:"state_root"`
}

// defaultNodeInfo is returned by Handler() when no live NodeInfo is provided.
var defaultNodeInfo = NodeInfo{
	Version:     "v0.9.0",
	Network:     "mainnet",
	ChainHeight: 0,
	PeerCount:   0,
	PendingTx:   0,
	StateRoot:   "0x0000000000000000000000000000000000000000000000000000000000000000",
}

// Handler returns an http.Handler that serves the embedded web dashboard.
//
// The handler uses a default [NodeInfo] for the /api/status endpoint.
// For production use with live node data prefer [ServeWithInfo].
//
// The returned mux handles:
//   - GET /api/status – JSON node status
//   - GET /* – embedded static files with SPA fallback to index.html
func Handler() http.Handler {
	return buildMux(defaultNodeInfo)
}

// ServeWithInfo returns an http.Handler identical to [Handler] but with the
// /api/status endpoint populated from the supplied [NodeInfo].
//
// Call this from your node's HTTP server so the dashboard shows real-time data:
//
//	info := web.NodeInfo{Version: node.Version(), Network: cfg.Network, ...}
//	http.Handle("/", web.ServeWithInfo(info))
func ServeWithInfo(info NodeInfo) http.Handler {
	return buildMux(info)
}

// buildMux constructs the internal ServeMux shared by [Handler] and [ServeWithInfo].
func buildMux(info NodeInfo) http.Handler {
	sub, _ := fs.Sub(staticFS, "static")
	fileServer := http.FileServer(http.FS(sub))

	mux := http.NewServeMux()

	// /api/status – lightweight JSON status endpoint.
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		addCORSHeaders(w, r)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(info)
	})

	// /* – static file server with SPA fallback and cache headers.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		addSecurityHeaders(w, r)
		addCORSHeaders(w, r)

		// Detect whether the requested path maps to a real embedded file.
		// If not, fall back to index.html so React Router can handle the route.
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		if _, err := fs.Stat(sub, path); err != nil {
			// Path does not exist in the embedded FS – serve index.html (SPA fallback).
			setCacheControl(w, "index.html")
			r2 := r.Clone(r.Context())
			r2.URL.Path = "/"
			fileServer.ServeHTTP(w, r2)
			return
		}

		setCacheControl(w, path)
		fileServer.ServeHTTP(w, r)
	})

	return mux
}

// setCacheControl sets an appropriate Cache-Control header based on the file extension.
//
//   - .js / .css: public, max-age=3600 (1 hour) – fingerprinted assets can be cached.
//   - everything else (HTML, JSON, …): no-store to ensure fresh content.
func setCacheControl(w http.ResponseWriter, path string) {
	switch {
	case strings.HasSuffix(path, ".js"), strings.HasSuffix(path, ".css"):
		w.Header().Set("Cache-Control", "public, max-age=3600, immutable")
	default:
		w.Header().Set("Cache-Control", "no-store")
	}
}

// addCORSHeaders sets permissive CORS headers suitable for local development.
// Adjust the allowed-origin list for production environments.
func addCORSHeaders(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// addSecurityHeaders sets an opinionated Content-Security-Policy and other
// hardening headers for the dashboard's HTML responses.
func addSecurityHeaders(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set(
		"Content-Security-Policy",
		"default-src 'self'; "+
			"script-src 'self' 'unsafe-inline'; "+
			"style-src 'self' 'unsafe-inline'; "+
			"img-src 'self' data:; "+
			"connect-src 'self'; "+
			"font-src 'self' data:; "+
			"frame-ancestors 'none';",
	)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Referrer-Policy", "same-origin")
}

// StatusResponse is the JSON shape returned by /api/status.
// It embeds [NodeInfo] and adds a server-side timestamp.
type StatusResponse struct {
	NodeInfo
	// Timestamp is the UTC time at which this response was generated.
	Timestamp time.Time `json:"timestamp"`
	// Online indicates that the node HTTP server is reachable.
	Online bool `json:"online"`
}
