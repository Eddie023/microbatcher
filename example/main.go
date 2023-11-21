package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
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

	processor := &SquareProcessor{}
	mb := microbatch.NewMicroBatch(microbatch.Config{
		BatchSize: 5,
		Processor: processor,
		Frequency: time.Second * 2,
	}, microbatch.WithMaxRetryAttempt(5))

	// simuate adding jobs to our microbatcher
	for i := 0; i <= 11; i++ {
		mb.Submit(microbatch.Job{Id: i, Task: i})
	}

	// attempt to add invalid job as well
	mb.Submit(microbatch.Job{
		Task: "invalid",
		Id:   12,
	})

	// start microbatcher
	go mb.RunInBatch(ctx)

	for {
		select {
		case <-mb.ShutdownChan:
			return
		case <-c:
			slog.Warn("Job interrupted", "message", "user interrupt")
			mb.Shutdown()
		}
	}
}
