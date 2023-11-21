package microbatch

type RetryableError struct {
	Message string
}

func (r *RetryableError) Error() string {
	return r.Message
}
