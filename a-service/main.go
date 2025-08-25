// tools/spam_unary/main.go
package main

import (
	"context"
	"flag"
	"log"
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/Yusufzhafir/worlder-team-assignment/common/protobuf" // adjust if needed
)

func main() {
	addr := flag.String("addr", "localhost:50051", "gRPC server address")
	total := flag.Int("n", 10000, "total number of requests")
	concurrency := flag.Int("c", 100, "number of concurrent workers")
	timeout := flag.Duration("timeout", 2*time.Second, "per-request timeout")
	val := flag.Float64("value", 10, "sensor value")
	typ := flag.String("type", "TEMP", "sensor type")
	id1 := flag.String("id1", "ABCDEFGH", "id1 (8 chars, uppercase)")
	id2 := flag.Int("id2", 1, "id2 (int)")
	flag.Parse()

	// 1) Dial once; let HTTP/2 multiplex for concurrency.
	dctx, cancelDial := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelDial()
	conn, err := grpc.DialContext(
		dctx,
		*addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
	)
	if err != nil {
		log.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()
	client := pb.NewIngestServiceClient(conn)

	// 2) Prepare work queue.
	jobs := make(chan int, *total)
	for i := 0; i < *total; i++ {
		jobs <- i
	}
	close(jobs)

	var ok, fail uint64
	latencies := make([]time.Duration, 0, *total)
	var latMu sync.Mutex

	startWall := time.Now()

	var wg sync.WaitGroup
	worker := func() {
		defer wg.Done()
		for range jobs {
			reqCtx, cancel := context.WithTimeout(context.Background(), *timeout)
			t0 := time.Now()

			_, err := client.Readings(reqCtx, &pb.SensorReading{
				Value:       *val,
				SensorType:  *typ,
				Id1:         *id1,
				Id2:         int32(*id2), // adjust type if your proto uses int64
				TimestampMs: time.Now().UnixMilli(),
			})
			cancel()

			dur := time.Since(t0)
			latMu.Lock()
			latencies = append(latencies, dur)
			latMu.Unlock()

			if err != nil {
				atomic.AddUint64(&fail, 1)
			} else {
				atomic.AddUint64(&ok, 1)
			}
		}
	}

	// 3) Spin up workers.
	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go worker()
	}
	wg.Wait()

	wall := time.Since(startWall)
	totalReq := float64(*total)
	rps := totalReq / wall.Seconds()

	// 4) Latency percentiles.
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	p := func(q float64) time.Duration {
		if len(latencies) == 0 {
			return 0
		}
		i := int(math.Ceil(q*float64(len(latencies))) - 1)
		if i < 0 {
			i = 0
		}
		if i >= len(latencies) {
			i = len(latencies) - 1
		}
		return latencies[i]
	}

	log.Printf("=== RESULTS ===")
	log.Printf("addr=%s  total=%d  concurrency=%d  timeout=%s",
		*addr, *total, *concurrency, timeout.String())
	log.Printf("ok=%d  fail=%d  wall=%s  throughput=%.0f req/s",
		ok, fail, wall, rps)
	if len(latencies) > 0 {
		log.Printf("latency: p50=%s  p90=%s  p95=%s  p99=%s  max=%s",
			p(0.50), p(0.90), p(0.95), p(0.99), latencies[len(latencies)-1])
	}
}
