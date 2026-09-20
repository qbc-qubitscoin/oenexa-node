// Package metrics registers and exposes Prometheus metrics for the OEN node.
package metrics

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ─────────────────────────────────────────────────────────────────────────────
// Gauges and counters
// ─────────────────────────────────────────────────────────────────────────────

var (
	// ChainHeight is the current confirmed block height.
	ChainHeight = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "oen",
		Name:      "chain_height",
		Help:      "Current confirmed block height.",
	})

	// PeerCount is the number of currently connected P2P peers.
	PeerCount = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "oen",
		Name:      "peer_count",
		Help:      "Number of connected P2P peers.",
	})

	// MempoolSize is the number of pending transactions in the mempool.
	MempoolSize = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "oen",
		Name:      "mempool_size",
		Help:      "Number of pending transactions in the mempool.",
	})

	// BlocksProduced is the total blocks produced by this validator.
	BlocksProduced = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "oen",
		Name:      "blocks_produced_total",
		Help:      "Total blocks produced by this validator since startup.",
	})

	// TxProcessed is the total transactions included in committed blocks.
	TxProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "oen",
		Name:      "transactions_processed_total",
		Help:      "Total transactions included in committed blocks since startup.",
	})

	// FeeBurned is the cumulative oenexa burned via base-fee since startup.
	FeeBurned = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "oen",
		Name:      "fee_burned_oenexa_total",
		Help:      "Cumulative oenexa burned as base fee since startup.",
	})

	// BaseFee is the current block base fee in oenexa/gas.
	BaseFee = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "oen",
		Name:      "base_fee_oenexa",
		Help:      "Current block base fee in oenexa per gas.",
	})

	// BlockProductionDuration tracks block-building latency.
	BlockProductionDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: "oen",
		Name:      "block_production_duration_seconds",
		Help:      "Time taken to build and commit each block.",
		Buckets:   prometheus.DefBuckets,
	})

	// RPCRequests is the total JSON-RPC requests received.
	RPCRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "oen",
		Name:      "rpc_requests_total",
		Help:      "Total JSON-RPC requests received, labeled by method.",
	}, []string{"method"})

	// RPCErrors is the total JSON-RPC errors returned.
	RPCErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "oen",
		Name:      "rpc_errors_total",
		Help:      "Total JSON-RPC errors returned, labeled by method.",
	}, []string{"method"})
)

// ─────────────────────────────────────────────────────────────────────────────
// Server
// ─────────────────────────────────────────────────────────────────────────────

// Serve starts an HTTP server that exposes Prometheus metrics at /metrics.
// It blocks until ctx is canceled.
func Serve(ctx context.Context, addr string) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
	}()

	log.Printf("[metrics] Prometheus endpoint listening on https://%s/metrics", addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("[metrics] server error: %v", err)
	}
}
