package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/slok/goresilience"
	"github.com/slok/goresilience/concurrencylimit"
	"github.com/slok/goresilience/concurrencylimit/execute"
	"github.com/slok/goresilience/concurrencylimit/limit"
	"github.com/slok/goresilience/metrics"
)

const exp04ResWorkers = 60

func mainExperiment04() {
	log.Printf("Experiment 04 (CoDel) (%d workers)\n", exp04ResWorkers)

	h := getServerHandler()
	reg := prometheus.NewRegistry()

	go http.ListenAndServe(addrMetrics, promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	http.ListenAndServe(addr, registerMiddlewares04(reg, h))
}

func registerMiddlewares04(reg *prometheus.Registry, next http.Handler) http.Handler {
	// Create our resilience pattern using a goresilience CoDel concurrency limiter.
	runner := goresilience.RunnerChain(
		metrics.NewMiddleware("exp04", metrics.NewPrometheusRecorder(reg)),
		concurrencylimit.NewMiddleware(concurrencylimit.Config{
			Executor: execute.NewAdaptiveLIFOCodel(execute.AdaptiveLIFOCodelConfig{}),
			Limiter:  limit.NewStatic(exp04ResWorkers),
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
