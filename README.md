# MicroBatch: Go Micro-Batching Library

## Usage 
1. Create struct that implements BatchProcessor Interface such as 
```go
type YourBatchProcessor struct {}

func (p *YourBatchProcessor) Process (job microbatch.Job) (microbatch.JobResult, error) {
    return microbatch.JobResult{
        Result: nil, 
        JobId: 1,
    }, nil 
}
```

2. Initiate a new Microbatcher using Factory function. Provide required configuration such as "batchSize" , "frequency" and your batch Processor that you created above.
```go
	mb := microbatch.NewMicroBatch(microbatch.Config{
		BatchSize: 5,
		Processor: &YourBatchProcessor{},
		Frequency: time.Second * 2,
	})
```

3. Add Jobs to your microbatcher
```go
   mb.Submit(microbatch.Job{Id: 1, Task: 10})
```

4. Start MicroBatcher
```go
   mb.RunInBatch(context.Background())
```

For a full implementation with optional configuration, please checkout the example provided in the repo.