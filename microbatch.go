// microbatch processes many tuples per iteration rather than just once based on batch size
package microbatch

import (
	"context"
	"log/slog"
	"time"
)

type Job struct {
	Task any
	Id   int
}

type JobResult struct {
	JobId  int
	Result any
}

// It is the responsibility of user to validate the provided job.
type BatchProcessor interface {
	Process(Job) (JobResult, error)
}

type MicroBatch struct {
	batchSize        int
	frequency        time.Duration
	processor        BatchProcessor
	ShutdownChan     chan struct{}
	cron             time.Ticker
	jobs             *multiConsumerQueue
	jobResult        chan JobResult
	maxRetryAttempts int
}

// Optional configurations
type Options struct {
	MaxRetryAttempts int
}

// WithMaxRetryAttempt will set the retryable errors maximum try to provided value.
func WithMaxRetryAttempt(num int) func(o *Options) {
	return func(o *Options) {
		o.MaxRetryAttempts = num
	}
}

// NewMicroBatch initiates new micro batcher with provided config.
func NewMicroBatch(batchSize int, processor BatchProcessor, frequency time.Duration, jobResult chan JobResult, opts ...func(o *Options)) *MicroBatch {
	o := Options{
		MaxRetryAttempts: MAX_RETRY_ATTEMPTS,
	}

	for _, opt := range opts {
		opt(&o)
	}

	return &MicroBatch{
		batchSize:        batchSize,
		processor:        processor,
		frequency:        frequency,
		cron:             *time.NewTicker(frequency),
		jobs:             &multiConsumerQueue{},
		jobResult:        jobResult,
		ShutdownChan:     make(chan struct{}),
		maxRetryAttempts: o.MaxRetryAttempts,
	}
}

// Run triggers the new microbatcher that based on the configured Frequency
// will periodically process the accepted Jobs in batch accoording to configured BatchSize
func (m *MicroBatch) Run(ctx context.Context, jobResult chan<- JobResult) {
	slog.Info("New microbatch started", "batch_size", m.batchSize, "frequency", m.frequency)
	for {
		select {
		case <-ctx.Done():
			m.Shutdown()
		case <-m.cron.C:
			batchedJobs := m.generateBatch(m.batchSize)

			slog.Info("Processing", "time", time.Now(), "remaining_jobs", m.jobs.Len())

			m.process(batchedJobs)
		}
	}
}

// Add new Job to the MicroBatcher
func (m *MicroBatch) Submit(j Job) {
	m.jobs.Enqueue(Job{
		Task: j.Task,
		Id:   j.Id,
	})
}

// Shutdown method will close our microbatcher from accepting any new jobs.
// This can be used to provide contextual information such as close after certain time or
// close when user interrupts
// gracefully shutdown after all previously accepted Jobs are processed
func (m *MicroBatch) Shutdown() {
	// send shutdown signal
	slog.Info("shutdown signal received... processing remaining jobs")

	if m.jobs.Len() > 0 {
		batchedJobs := m.generateBatch(m.jobs.Len())
		m.process(batchedJobs)
	}

	slog.Info("no remaining jobs...gracefully shutting down")
}

// Retrieve items from queue in batches
// batches are generated based on configured BatchSize
func (m *MicroBatch) generateBatch(batchSize int) []Job {
	batchJobs := m.jobs.Dequeue(batchSize)

	return batchJobs
}

func (m *MicroBatch) process(jobs []Job) {
	for _, job := range jobs {
		res, err := processWithRetry(m.processor, job, m.maxRetryAttempts)
		if err != nil {
			// jobResult <- err
			continue
		}

		m.jobResult <- res
	}
}
