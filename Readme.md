# Resilience talk

The experiment server will be run with this commands:

Run server with constrained limits:

```bash
docker run --rm -it -v $PWD:/src --network=host --memory="30m" --cpus="0.02" --name exp01 golang:1.11 /bin/bash
```

The experiments will be tested using this commands:

In background run:

```bash
echo "GET http://127.0.0.1:8000" | vegeta attack -rate=70/s | vegeta report
```

Suddenly run:

```bash
echo "GET http://127.0.0.1:8000" | vegeta attack -rate=150/s -duration=1m | vegeta report
```

Check whenever to check the state:

```bash
time curl http://127.0.0.1:8000 -v
```

## Experiment 01 (naked server)

- The server has no resilience set.
- Limit memory and CPU of the container running the server.
- Latency of 500ms on requests.

Hypothesis: The server will start increasing the latency when we reach the limits and it will collapse by CPU or killed by OOM

## Experiment 02 (Bulkhead)

Hypothesis: The server will limit the configured number of Request can process and it will timeout the ones standing on the queue to be processed, if the configuration is correct the server will not crash and will recover.

Good:

- Load shedding.
- Server recovers eventually (in ~5m).

Bad:

- It has a lot of delay.
- Lot's of tests to get the correct configuration.
- Very static configuration.

## Experiment 03 (Bulkhead + circuit breaker)

Good:

- Load shedding.
- Server recovers eventually (in ~3m).
- Fails faster than bulkhead.

Bad:

- Very static configuration.
- Lot's of tests to get the correct configuration.
- Miss request even the server is ok (until the CB realizes we are good again)

## Experiment 04 (CoDel)

## Experiment 05 (concurrency adaptive limit)
