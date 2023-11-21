// microbatch processes many tuples per iteration rather than just once based on batch size
package microbatch

import (
	"context"
	"log/slog"
	"time"
)

type Config struct {
	// size of each batch cycle
	BatchSize int
	// Processor should implement BatchProcessor interface
	Processor BatchProcessor
	// how often the system processes or dispatches batches of task
	Frequency time.Duration
}

// Optional configurations
type Options struct {
	MaxRetryAttempts int
}

type MicroBatch struct {
	BatchSize        int
	Frequency        time.Duration
	Processor        BatchProcessor
	ShutdownChan     chan struct{}
	cron             time.Ticker
	jobs             *MultiConsumerQueue
	maxRetryAttempts int
}

// NewMicroBatch initiates new micro batcher with provided config.
func NewMicroBatch(cfg Config, opts ...func(o *Options)) *MicroBatch {

	o := Options{
		MaxRetryAttempts: MAX_RETRY_ATTEMPTS,
	}

	for _, opt := range opts {
		opt(&o)
	}

	m := MicroBatch{
		BatchSize:        cfg.BatchSize,
		maxRetryAttempts: o.MaxRetryAttempts,
		Frequency:        time.Second * 5,
		Processor:        cfg.Processor,
		cron:             *time.NewTicker(cfg.Frequency),
		jobs:             &MultiConsumerQueue{},
		ShutdownChan:     make(chan struct{}),
	}

	return &m
}

// Add new Job to the MicroBatcher
func (m *MicroBatch) Submit(j Job) {
	m.jobs.Enqueue(j)
}

// Shutdown method will close our microbatcher from accepting any new jobs.
// This can be used to provide contextual information such as close after certain time or
// close when user interrupts
func (m *MicroBatch) Shutdown() {
	close(m.ShutdownChan)
}

// Retrieve items from queue in batches
// batches are generated based on configured BatchSize
func (m *MicroBatch) generateBatch() []Job {
	batchJobs := m.jobs.Dequeue(m.BatchSize)

	return batchJobs
}

// RunInBatch triggers the new microbatcher that based on the configured Frequency
// will periodically process the accepted Jobs in batch accoording to configured BatchSize
func (m *MicroBatch) RunInBatch(ctx context.Context) {
	slog.Info("New microbatch started", "batch_size", m.BatchSize, "frequency", m.Frequency)
	for {
		select {
		case <-m.cron.C:
			// create batch
			batchedJobs := m.generateBatch()

			if len(batchedJobs) == 0 {
				slog.Info("Successfully completed all accepted jobs")
				close(m.ShutdownChan)
				return
			}

			for _, job := range batchedJobs {
				jobResult, err := ProcessWithRetry(ctx, m.Processor, job, m.maxRetryAttempts)
				if err != nil {
					// FATAL: unhandled error occurred
					slog.Error("Job failed", "job_id", job.Id, "task_input", job.Task, "error msg", err.Error(), "status", "skipping")
					continue
				}

				slog.Info("Job completed", "job_id", job.Id, "task_input", job.Task, "result", jobResult.Result)
			}
		}
	}
}

// WithMaxRetryAttempt will set the retryable errors maximum try to provided value.
func WithMaxRetryAttempt(num int) func(o *Options) {
	return func(o *Options) {
		o.MaxRetryAttempts = num
	}
}
