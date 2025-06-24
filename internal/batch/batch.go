package batch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/portal26/event_receiver_service/internal/config"
	"github.com/portal26/event_receiver_service/internal/metrics"
	"github.com/portal26/event_receiver_service/internal/model"
)

type BatchProcessor struct {
	cfg       *config.Config
	client    *s3.Client
	events    chan model.Event
	batch     []model.Event
	batchSize int
	mu        sync.Mutex
	lastFlush time.Time
	stop      chan struct{}
	wg        sync.WaitGroup
}

func NewBatchProcessor(cfg *config.Config) *BatchProcessor {
	awsCfg, err := awsConfig.LoadDefaultConfig(context.Background(), awsConfig.WithRegion(cfg.AWSRegion))
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	return &BatchProcessor{
		cfg:       cfg,
		client:    s3.NewFromConfig(awsCfg),
		events:    make(chan model.Event, 1000),
		batch:     make([]model.Event, 0, 1000),
		batchSize: cfg.BatchSizeMB * 1024 * 1024,
		lastFlush: time.Now(),
		stop:      make(chan struct{}),
	}
}

func (bp *BatchProcessor) Start() {
	bp.wg.Add(1)
	go bp.process()
}

func (bp *BatchProcessor) Stop() {
	close(bp.stop)
	bp.wg.Wait()
	bp.flush()
}

func (bp *BatchProcessor) AddEvent(event model.Event) {
	bp.events <- event
}

func (bp *BatchProcessor) process() {
	defer bp.wg.Done()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case event := <-bp.events:
			bp.mu.Lock()
			bp.batch = append(bp.batch, event)
			bp.mu.Unlock()

			if bp.shouldFlush() {
				bp.flush()
			}
		case <-ticker.C:
			if time.Since(bp.lastFlush) >= time.Duration(bp.cfg.BatchTimeoutSeconds)*time.Second {
				bp.flush()
			}
		case <-bp.stop:
			return
		}
	}
}
func (bp *BatchProcessor) shouldFlush() bool {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	if len(bp.batch) == 0 {
		return false
	}

	var size int
	for _, event := range bp.batch {
		data, _ := json.Marshal(event)
		size += len(data) + 1 // +1 for newline
	}
	return size >= bp.batchSize
}

func (bp *BatchProcessor) flush() {
	bp.mu.Lock()
	if len(bp.batch) == 0 {
		bp.mu.Unlock()
		return
	}

	batch := make([]model.Event, len(bp.batch))
	copy(batch, bp.batch)
	bp.batch = bp.batch[:0]
	bp.mu.Unlock()

	var buf bytes.Buffer
	for _, event := range batch {
		data, _ := json.Marshal(event)
		buf.Write(data)
		buf.WriteByte('\n')
	}

	key := fmt.Sprintf("events/batch_%d.jsonl", time.Now().Unix())
	_, err := bp.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(bp.cfg.S3Bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(buf.Bytes()),
	})
	if err != nil {
		metrics.S3Errors.Inc()
		log.Printf("Failed to write to S3: %v", err)
		return
	}

	metrics.S3Writes.Inc()
	log.Printf("Flushed batch to S3: %s (%d bytes)", key, buf.Len())
	bp.lastFlush = time.Now()
}
