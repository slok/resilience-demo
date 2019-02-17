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

func mainExperiment05() {
	log.Println("Experiment 05 (concurrency adaptive limit)")

	h := getServerHandler()
	reg := prometheus.NewRegistry()

	go http.ListenAndServe(addrMetrics, promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	http.ListenAndServe(addr, registerMiddlewares05(reg, h))
}

func registerMiddlewares05(reg *prometheus.Registry, next http.Handler) http.Handler {
	// Create our resilience pattern using a goresilience CoDel concurrency limiter.
	runner := goresilience.RunnerChain(
		metrics.NewMiddleware("exp05", metrics.NewPrometheusRecorder(reg)),
		concurrencylimit.NewMiddleware(concurrencylimit.Config{
			Executor: execute.NewLIFO(execute.LIFOConfig{}),
			Limiter:  limit.NewAIMD(limit.AIMDConfig{}),
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
