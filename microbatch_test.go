package microbatch

import (
	"context"
	"testing"
)

type DummyProcessor struct{}

// DummyProcessor implements required batchProcessor interface.
func (d *DummyProcessor) Process(j Job) (JobResult, error) {
	return JobResult{
		JobId:  j.Id,
		Result: true,
	}, nil
}

func TestMicroBatch_RunInBatch(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		jobs []Job
	}{
		{
			name: "Should run successfully run and exit even without any jobs submitted",
			cfg: Config{
				BatchSize: 10,
				Processor: &DummyProcessor{},
				Frequency: 2,
			},
			jobs: []Job{},
		},
		{
			name: "Should successfully return job result for submitted jobs",
			cfg: Config{
				BatchSize: 2,
				Processor: &DummyProcessor{},
				Frequency: 2,
			},
			jobs: []Job{{Task: 1, Id: 1}, {Task: 2, Id: 2}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMicroBatch(tt.cfg)

			for _, j := range tt.jobs {
				m.Submit(j)
			}
			m.RunInBatch(context.TODO())
		})
	}
}
