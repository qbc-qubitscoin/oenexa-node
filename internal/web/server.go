// Package web provides the REST API for the OENEXA node.
//
// It exposes a /api/status JSON endpoint. The frontend dashboard has been
// extracted into a separate repository (oenexa-frontend) which can connect
// to this API.
package web

import (
	"encoding/json"
	"net/http"
	"time"
)

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

// Handler returns an http.Handler that serves the node API.
//
// The handler uses a default [NodeInfo] for the /api/status endpoint.
// For production use with live node data prefer [ServeWithInfo].
func Handler() http.Handler {
	return buildMux(defaultNodeInfo)
}

// ServeWithInfo returns an http.Handler identical to [Handler] but with the
// /api/status endpoint populated from the supplied [NodeInfo].
func ServeWithInfo(info NodeInfo) http.Handler {
	return buildMux(info)
}

// buildMux constructs the internal ServeMux shared by [Handler] and [ServeWithInfo].
func buildMux(info NodeInfo) http.Handler {
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
		
		resp := StatusResponse{
			NodeInfo:  info,
			Timestamp: time.Now().UTC(),
			Online:    true,
		}
		
		_ = json.NewEncoder(w).Encode(resp)
	})

	return mux
}

// addCORSHeaders sets permissive CORS headers suitable for local development.
// Adjust the allowed-origin list for production environments.
func addCORSHeaders(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
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
