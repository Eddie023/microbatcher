package microbatch

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
	Process(j Job) (JobResult, error)
}
