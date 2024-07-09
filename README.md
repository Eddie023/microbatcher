# MicroBatcher: Go Micro-Batching Library

1. Simple Microbatching example written in golang. 

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

2. Initiate a new Microbatcher using Factory function. Provide required  such as "batchSize" , "frequency" and your batch Processor that you created above.
```go
  mb := microbatch.NewMicroBatch(batchSize, yourBatchProcessor, frequency)
```

3. Create a channel to get your successfully ran job results. 
```go
  jobResult := make(chan microbatch.JobResult{})
```

4. Start MicroBatcher and pass your job result channel. 
```go
  mb.Run(context.Background(), jobResult)
```

3. Add Jobs to your microbatcher
```go
  mb.Submit(microbatch.Job{Id: 1, Task: 10})
```


For a full implementation with optional configuration, please checkout the example provided in the repo.
