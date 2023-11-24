package microbatch

type RetryableError struct {
	Message string
}

func (r *RetryableError) Error() string {
	return r.Message
}

type JobFailerError struct {
}

func (j *JobFailerError) Error() string {
	return "job failed"
}
