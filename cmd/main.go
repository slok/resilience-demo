package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/slok/goresilience/bulkhead"
)

const (
	addr        = ":8080"
	addrMetrics = ":8083"
	workers     = 200
)

const (
	exp01 = 1
	exp02 = 2
	exp03 = 3
	exp04 = 4
	exp05 = 5
)

func main() {
	exp := flag.Int("experiment", 1, "the experiment server to run")
	flag.Parse()

	// Run the required experiment.
	switch *exp {
	case exp01:
		mainExperiment01()
	case exp02:
		mainExperiment02()
	case exp03:
		mainExperiment03()
	case exp04:
		mainExperiment04()
	case exp05:
		mainExperiment05()
	default:
		fmt.Fprintf(os.Stderr, "unknown experiment")
		os.Exit(1)
	}
}

// serverHandler returns the handler that will process the requests.
// it mimics a job by sleeping for a time. Also to mimic the congestion
// of server in the same circumstances it limits the execution to a fixed
// number of workers.
func getServerHandler() http.Handler {
	runner := bulkhead.New(bulkhead.Config{
		Workers:     workers,
		MaxWaitTime: 0,
	})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		runner.Run(r.Context(), func(_ context.Context) error {
			// Create goroutines.
			q := 50
			res := make(chan string)
			for i := 0; i < q; i++ {
				i := i
				go func() {
					// Get a random number
					s1 := rand.NewSource(time.Now().UnixNano())
					randN := rand.New(s1).Intn(99999999)
					res <- fmt.Sprintf("result-%d-%d", randN, i)
				}()
			}

			// Get results.
			results := map[string]string{}
			for i := 0; i < q; i++ {
				result := <-res
				results[fmt.Sprintf("id%d", i)] = result
			}

			// Sleep some time.
			lt := getLatency(250, 40)
			time.Sleep(lt)
			return nil
		})
	})
}

func getLatency(ms int, jitterPercent int) time.Duration {
	randMax := jitterPercent * ms / 100

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	jitter := r1.Intn(randMax)

	// Add latency or reduce randomly.
	if r1.Intn(100)%2 == 0 {
		ms -= jitter
	} else {
		ms += jitter
	}

	return time.Duration(ms) * time.Millisecond
}
