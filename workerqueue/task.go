package workerqueue

import (
	"context"
	"errors"
	"sync"
)

// Task struct is just a wrapper for sync.WaitGroup
type Task struct {
	wg sync.WaitGroup
}

// Run method should be used when you want to run a job in background
// but still ensure it is completed before program exits
func (t *Task) Run(job Job) {
	t.wg.Add(1)

	go func() {
		defer t.wg.Done()

		job.Process()
	}()
}

// Shutdown method will wait for all the go routines associated with the tracker to
// complete or context to expire
func (t *Task) Shutdown(ctx context.Context) error {

	// Create a channel to signal when the waitgroup is finished.
	ch := make(chan struct{})

	// Create a goroutine to wait for all other goroutines to
	// be done then close the channel to unblock the select.
	go func() {
		t.wg.Wait()
		close(ch)
	}()

	// Block this function from returning. Wait for either the
	// waitgroup to finish or the context to expire.
	select {
	case <-ch:
		return nil
	case <-ctx.Done():
		return errors.New("timeout")
	}
}
