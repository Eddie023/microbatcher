package microbatch

import (
	"context"
	"errors"
	"log/slog"
)

const MAX_RETRY_ATTEMPTS = 3

func ProcessWithRetry(ctx context.Context, f BatchProcessor, job Job, retryNum int) (JobResult, error) {
	for i := 0; i < retryNum; i++ {
		select {
		case <-ctx.Done():
			return JobResult{}, ctx.Err()
		default:
		}
		result, err := f.Process(job)
		if err == nil {
			return result, nil
		}

		if err != nil {
			var serr *RetryableError
			if errors.As(err, &serr) {
				slog.Error("job failed", "job_id", job.Id, "task_input", job.Task, "error_msg", err.Error(), "error_type", "retryable", "status", "retrying", "attempt", i)

				continue
			}

			return JobResult{}, err
		}
	}

	return JobResult{}, errors.New("MAX_RETRY_EXCEEDED")
}
