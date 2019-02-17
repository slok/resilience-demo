package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/slok/goresilience"
	"github.com/slok/goresilience/bulkhead"
	"github.com/slok/goresilience/metrics"
)

const (
	exp02ResTimeout = 1 * time.Second
	exp02ResWorkers = 60
)

func mainExperiment02() {
	log.Printf("Experiment 02 (Bulkhead) (%d workers) (%s timeout)\n", exp02ResWorkers, exp02ResTimeout)

	h := getServerHandler()
	reg := prometheus.NewRegistry()

	go http.ListenAndServe(addrMetrics, promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	http.ListenAndServe(addr, registerMiddlewares02(reg, h))
}

func registerMiddlewares02(reg *prometheus.Registry, next http.Handler) http.Handler {
	// Create our resilience pattern using a goresilience bulkhead.
	runner := goresilience.RunnerChain(
		metrics.NewMiddleware("exp02", metrics.NewPrometheusRecorder(reg)),
		bulkhead.NewMiddleware(bulkhead.Config{
			Workers:     exp02ResWorkers,
			MaxWaitTime: exp02ResTimeout,
		}),
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		err := runner.Run(r.Context(), func(_ context.Context) error {
			next.ServeHTTP(w, r)
			return nil
		})

		if err != nil {
			log.Printf("request dropped (%s): %s\n", time.Since(start), err)
			w.WriteHeader(http.StatusTooManyRequests)
		}
	})
}
