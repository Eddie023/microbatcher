package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eddie023/microbatch"
)

type SquareProcessor struct{}

// For a given job, SquareProcessor will return the square of the job's task value.
func (ibp *SquareProcessor) Process(j microbatch.Job) (microbatch.JobResult, error) {
	taskValue, ok := j.Task.(int)
	if !ok {
		return microbatch.JobResult{}, errors.New("invalid job")
	}

	result := taskValue * taskValue

	// simulate a retryable error
	if j.Id == 4 {
		return microbatch.JobResult{}, &microbatch.RetryableError{
			Message: "api limit reached",
		}
	}

	return microbatch.JobResult{
		JobId:  j.Id,
		Result: result,
	}, nil
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	jobResult := make(chan microbatch.JobResult)
	squareProcessor := &SquareProcessor{}
	mb := microbatch.NewMicroBatch(5, squareProcessor, time.Second*2, jobResult, microbatch.WithMaxRetryAttempt(3))

	// Setup signal catching
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// start microbatcher
	go mb.Run(ctx, jobResult)

	// simuate adding jobs to our microbatcher
	for i := 0; i <= 11; i++ {
		go func(i int) {
			mb.Submit(microbatch.Job{Id: i, Task: i})
		}(i)
	}

	go func() {
		for c := range jobResult {
			slog.Info("Job completed", "job_id", c.JobId, "result", c.Result)
		}
	}()

	for {
		select {
		case <-shutdown:
			slog.Warn("Job interrupted", "message", "user interrupt")
			mb.Shutdown()
			return
		}
	}
}
