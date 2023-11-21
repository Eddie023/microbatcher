package microbatch

import (
	"testing"
)

func TestMultiConsumerQueue_Enqueue(t *testing.T) {
	tests := []struct {
		name   string
		values []Job
		want   int
	}{
		{
			name: "Should be able to add values to the queue",
			values: []Job{
				{Task: 1, Id: 1},
				{Task: 2, Id: 2},
				{Task: 3, Id: 3},
				{Task: 4, Id: 4},
			},
			want: 4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mcq := &MultiConsumerQueue{}

			for _, j := range tt.values {
				mcq.Enqueue(j)
			}

			if tt.want != mcq.Len() {
				t.Fatalf("enqueue failed: want='%d' got='%d'", tt.want, mcq.Len())
			}

		})
	}
}

func TestMultiConsumerQueue_Dequeue(t *testing.T) {
	tests := []struct {
		name       string
		values     []Job
		batchSize  int
		want       int
		checkItems bool
	}{
		{
			name: "Dequeue with batchsize 1 should remove single item",
			values: []Job{
				{Task: 1, Id: 1},
				{Task: 2, Id: 2},
				{Task: 3, Id: 3},
				{Task: 4, Id: 4},
			},
			batchSize: 1,
			want:      3,
		},
		{
			name: "Dequeue with batchsize greater than total item should remove all item",
			values: []Job{
				{Task: 1, Id: 1},
				{Task: 2, Id: 2},
				{Task: 3, Id: 3},
				{Task: 4, Id: 4},
			},
			batchSize: 10,
			want:      0,
		},
		{
			name: "Dequeue with batchsize 2 should remove first two item of the queue",
			values: []Job{
				{Task: 1, Id: 1},
				{Task: 2, Id: 2},
				{Task: 3, Id: 3},
				{Task: 4, Id: 4},
			},
			batchSize:  2,
			want:       2,
			checkItems: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mcq := &MultiConsumerQueue{}

			for _, j := range tt.values {
				mcq.Enqueue(j)
			}

			got := mcq.Dequeue(tt.batchSize)

			if tt.want != mcq.Len() {
				t.Fatalf("dequeue failed: want='%d' got='%d'", tt.want, mcq.Len())
			}

			if tt.checkItems {
				gotWrongItems := true
				for _, j := range got {
					for i := 0; i <= tt.batchSize; i++ {
						if j == tt.values[i] {
							gotWrongItems = false
						}
					}
				}

				if gotWrongItems {
					t.Fatalf("dequeued wrong items: want='%v' got=%v", tt.values[:tt.batchSize], got)
				}
			}
		})
	}
}
