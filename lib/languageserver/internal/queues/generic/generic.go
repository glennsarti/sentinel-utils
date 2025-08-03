package generic

import (
	"context"
	"errors"
	"sync"
)

func NewGenericQueue[J any](
	maxWorkers int,
	maxJobs int,
	processor func(job J) error,
) *GenericQueue[J] {
	return &GenericQueue[J]{
		maxWorkers: maxWorkers,
		maxJobs:    maxJobs,
		processor:  processor,
	}
}

type GenericQueue[J any] struct {
	maxWorkers int
	maxJobs    int
	processor  func(job J) error
	jobs       chan J
	running    bool
}

func (gq *GenericQueue[J]) StartAsync(ctx context.Context) error {
	if gq.running {
		return errors.New("queue already running")
	}
	go gq.Start(ctx) //nolint:errcheck
	return nil
}

func (gq *GenericQueue[J]) Start(_ context.Context) error {
	if gq.running {
		return errors.New("queue already running")
	}
	gq.running = true
	cancelChan := make(chan struct{})

	wg := &sync.WaitGroup{}
	gq.jobs = make(chan J, gq.maxJobs)

	for range gq.maxWorkers {
		wg.Add(1)
		go gq.executor(gq.jobs, cancelChan)
	}

	wg.Wait()

	close(cancelChan)
	gq.running = false
	return nil
}

// TryEnqueue tries to enqueue a job to the given job channel. Returns true if
// the operation was successful, and false if enqueuing would not have been
// possible without blocking. Job is not enqueued in the latter case.
func (gq *GenericQueue[J]) Enqueue(job J) bool {
	if !gq.running {
		return false
	}
	select {
	case gq.jobs <- job:
		return true
	default:
		return false
	}
}

func (gq *GenericQueue[J]) executor(jobChan <-chan J, cancelChan <-chan struct{}) {
	for {
		select {
		case <-cancelChan:
			return

		case job := <-jobChan:
			// TODO: Should really handle errors here.
			_ = gq.processor(job)
		}
	}
}
