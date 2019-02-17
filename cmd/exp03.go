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
	"github.com/slok/goresilience/circuitbreaker"
	"github.com/slok/goresilience/metrics"
)

const (
	exp03ResTimeout = 1 * time.Second
	exp03ResWorkers = 60
)

func mainExperiment03() {
	log.Printf("Experiment 03 (Bulkhead + circuit breaker) (%d workers) (%s timeout)\n", exp03ResWorkers, exp03ResTimeout)

	h := getServerHandler()
	reg := prometheus.NewRegistry()

	go http.ListenAndServe(addrMetrics, promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	http.ListenAndServe(addr, registerMiddlewares03(reg, h))
}

func registerMiddlewares03(reg *prometheus.Registry, next http.Handler) http.Handler {
	// Create our resilience pattern using a goresilience chain composed
	// by a bulkhead and a circuitbreaker.
	runner := goresilience.RunnerChain(
		metrics.NewMiddleware("exp03", metrics.NewPrometheusRecorder(reg)),
		circuitbreaker.NewMiddleware(circuitbreaker.Config{}),
		bulkhead.NewMiddleware(bulkhead.Config{
			Workers:     exp03ResWorkers,
			MaxWaitTime: exp03ResTimeout,
		}),
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		r = r.WithContext(ctx)

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
