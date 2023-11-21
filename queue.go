package microbatch

import (
	"fmt"
	"sync"
)

type node struct {
	Value Job
	Next  *node
}

// MultiConsumerQueue is a simple linkedlist implementation such that
// we can insert item in Last In First Out (LIFO) fashion.
type MultiConsumerQueue struct {
	Head  *node
	mutex sync.Mutex
}

func (mcq *MultiConsumerQueue) Enqueue(value Job) {
	mcq.mutex.Lock()
	defer mcq.mutex.Unlock()

	node := &node{
		Value: value,
	}

	// the list is empty
	if mcq.Head == nil {
		mcq.Head = node

		return
	}

	current := mcq.Head
	for current.Next != nil {
		current = current.Next
	}

	current.Next = node
}

func (mcq *MultiConsumerQueue) Dequeue(batchsize int) []Job {
	mcq.mutex.Lock()
	defer mcq.mutex.Unlock()

	if mcq.Head == nil {
		return []Job{}
	}

	jobs := []Job{}
	for i := 0; i < batchsize; i++ {
		if mcq.Head == nil {
			break
		}

		value := mcq.Head.Value
		jobs = append(jobs, value)

		mcq.Head = mcq.Head.Next
	}

	return jobs
}

func (mcq *MultiConsumerQueue) Visit() {
	mcq.mutex.Lock()
	defer mcq.mutex.Unlock()

	current := mcq.Head
	for current != nil {
		fmt.Println("node is", current.Value.Task)
		current = current.Next
	}
}

func (mcq *MultiConsumerQueue) Len() int {
	mcq.mutex.Lock()
	defer mcq.mutex.Unlock()

	count := 0
	current := mcq.Head
	for current != nil {
		count++
		current = current.Next
	}

	return count
}
