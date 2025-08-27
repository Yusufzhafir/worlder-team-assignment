package usecase

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"

	pb "github.com/Yusufzhafir/worlder-team-assignment/common/protobuf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type DataGenerator interface {
	Connect() error
	Close()
	Start() bool
	Stop() bool
	SetFrequency(freq time.Duration)
	IsRunning() bool
	GetStats() (uint64, uint64)
	SpamRequests(req SpamRequest) map[string]interface{}
	Config(config GeneratorConfig)
	GetDetailedStats() map[string]interface{}
}
type DataGeneratorImpl struct {
	mu        sync.RWMutex
	isRunning bool
	stopChan  chan struct{}
	frequency time.Duration // interval between requests
	client    pb.IngestServiceClient
	conn      *grpc.ClientConn
	startTime time.Time
	// Configuration
	serverAddr     string
	sensorValue    float64
	sensorType     string
	id1            string
	id2            int32
	requestTimeout time.Duration

	// Stats
	totalSent   uint64
	totalFailed uint64
}

type SpamRequest struct {
	Total       int     `json:"total"`
	Concurrency int     `json:"concurrency"`
	Timeout     string  `json:"timeout"`
	Value       float64 `json:"value"`
	Type        string  `json:"type"`
	ID1         string  `json:"id1"`
	ID2         int32   `json:"id2"`
}

func NewDataGenerator(serverAddr string) DataGenerator {
	return &DataGeneratorImpl{
		serverAddr:     serverAddr,
		frequency:      time.Second, // default: 1 req/sec
		sensorValue:    10.0,
		sensorType:     "TEMP",
		id1:            "ABCDEFGH",
		id2:            1,
		requestTimeout: 1 * time.Second,
	}
}

type GeneratorConfig struct {
	Value      float64 `json:"value"`
	Type       string  `json:"type"`
	ID1        string  `json:"id1"`
	ID2        int32   `json:"id2"`
	ServerAddr string  `json:"server_addr"`
}

func (dg *DataGeneratorImpl) Config(config GeneratorConfig) {

	dg.mu.Lock()
	if config.Value != 0 {
		dg.sensorValue = config.Value
	}
	if config.Type != "" {
		dg.sensorType = config.Type
	}
	if config.ID1 != "" {
		dg.id1 = config.ID1
	}
	if config.ID2 != 0 {
		dg.id2 = config.ID2
	}
	dg.mu.Unlock()
}

func (dg *DataGeneratorImpl) Connect() error {
	dctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		dctx,
		dg.serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
	)
	if err != nil {
		return err
	}

	dg.conn = conn
	dg.client = pb.NewIngestServiceClient(conn)
	return nil
}

func (dg *DataGeneratorImpl) Close() {
	if dg.conn != nil {
		dg.conn.Close()
	}
}

func (dg *DataGeneratorImpl) Start() bool {
	dg.mu.Lock()
	defer dg.mu.Unlock()

	if dg.isRunning {
		return false // already running
	}

	dg.isRunning = true
	dg.stopChan = make(chan struct{})
	dg.startTime = time.Now()

	go dg.generateContinuously()
	return true
}

func (dg *DataGeneratorImpl) Stop() bool {
	dg.mu.Lock()
	defer dg.mu.Unlock()

	if !dg.isRunning {
		return false // not running
	}

	dg.isRunning = false
	close(dg.stopChan)
	return true
}

func (dg *DataGeneratorImpl) SetFrequency(freq time.Duration) {
	dg.mu.Lock()
	defer dg.mu.Unlock()
	dg.frequency = freq
}

func (dg *DataGeneratorImpl) IsRunning() bool {
	dg.mu.RLock()
	defer dg.mu.RUnlock()
	return dg.isRunning
}

func (dg *DataGeneratorImpl) GetStats() (uint64, uint64) {
	return atomic.LoadUint64(&dg.totalSent), atomic.LoadUint64(&dg.totalFailed)
}

func (dg *DataGeneratorImpl) GetDetailedStats() map[string]interface{} {
	totalSent := atomic.LoadUint64(&dg.totalSent)
	totalFailed := atomic.LoadUint64(&dg.totalFailed)
	totalRequests := totalSent + totalFailed

	dg.mu.RLock()
	isRunning := dg.isRunning
	startTime := dg.startTime
	frequency := dg.frequency
	dg.mu.RUnlock()

	var uptimeSeconds float64
	var overallRPS float64

	if isRunning && !startTime.IsZero() {
		uptimeSeconds = time.Since(startTime).Seconds()
		if uptimeSeconds > 0 {
			overallRPS = float64(totalRequests) / uptimeSeconds
		}
	}
	return map[string]interface{}{
		"Total_Sent":        totalSent,
		"Total_Failed":      totalFailed,
		"Total_Requests":    totalRequests,
		"Uptime_Seconds":    uptimeSeconds,
		"Overall_RPS":       overallRPS,
		"Is_Running":        isRunning,
		"Configured_FreqMs": frequency.Milliseconds(),
	}
}

func (dg *DataGeneratorImpl) generateContinuously() {
	ticker := time.NewTicker(dg.frequency)
	defer ticker.Stop()

	// Keep track of the current ticker frequency
	currentFreq := dg.frequency
	for {
		select {
		case <-dg.stopChan:
			return
		case <-ticker.C:
			// Update ticker frequency if changed
			dg.mu.RLock()
			newFreq := dg.frequency
			log.Default().Printf("this is the current frequency %v", currentFreq)
			dg.mu.RUnlock()

			if newFreq != currentFreq {
				ticker.Stop()
				ticker = time.NewTicker(newFreq)
				currentFreq = newFreq
			}

			dg.sendSingleReading()
		}
	}
}

func (dg *DataGeneratorImpl) sendSingleReading() {
	ctx, cancel := context.WithTimeout(context.Background(), dg.requestTimeout)
	defer cancel()

	_, err := dg.client.Readings(ctx, &pb.SensorReading{
		Value:       dg.sensorValue,
		SensorType:  dg.sensorType,
		Id1:         dg.id1,
		Id2:         dg.id2,
		TimestampMs: time.Now().UnixMilli(),
	})

	if err != nil {
		atomic.AddUint64(&dg.totalFailed, 1)
		log.Printf("Failed to send reading: %v", err)
	} else {
		atomic.AddUint64(&dg.totalSent, 1)
	}
}

func (dg *DataGeneratorImpl) SpamRequests(req SpamRequest) map[string]interface{} {
	timeout, _ := time.ParseDuration(req.Timeout)
	if timeout == 0 {
		timeout = 2 * time.Second
	}

	jobs := make(chan int, req.Total)
	for i := 0; i < req.Total; i++ {
		jobs <- i
	}
	close(jobs)

	var ok, fail uint64
	startTime := time.Now()

	var wg sync.WaitGroup
	worker := func() {
		defer wg.Done()
		for range jobs {
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			_, err := dg.client.Readings(ctx, &pb.SensorReading{
				Value:       req.Value,
				SensorType:  req.Type,
				Id1:         req.ID1,
				Id2:         req.ID2,
				TimestampMs: time.Now().UnixMilli(),
			})
			cancel()

			if err != nil {
				atomic.AddUint64(&fail, 1)
			} else {
				atomic.AddUint64(&ok, 1)
			}
		}
	}

	for i := 0; i < req.Concurrency; i++ {
		wg.Add(1)
		go worker()
	}
	wg.Wait()

	duration := time.Since(startTime)
	rps := float64(req.Total) / duration.Seconds()

	return map[string]interface{}{
		"total_requests":      req.Total,
		"successful":          ok,
		"failed":              fail,
		"duration_ms":         duration.Milliseconds(),
		"requests_per_second": rps,
	}
}
