package workerqueue

import (
	"context"
	"errors"
	"sync"
)

// TaskQueue should be used when you want to limit the number of tasks you want
// to process in background
type TaskQueue struct {
	queue    chan Job
	quitChan chan bool
	qlen     int
	wg       sync.WaitGroup
}

// NewTaskQueue will create a new taskqueue of buffered job channel
func NewTaskQueue(l int) *TaskQueue {
	var t = &TaskQueue{queue: make(chan Job, l), quitChan: make(chan bool, l), qlen: l}
	return t
}

// Start method will create a go routine that will process the work in background
func (t *TaskQueue) Start() {
	for i := 0; i < t.qlen; i++ {
		go func() {
			for {
				select {
				case job := <-t.queue: // received a job
					if job != nil {
						job.Process()
						t.wg.Done()
					}

				// We have been asked to stop the processing
				case <-t.quitChan:
					return
				}
			}
		}()
	}

}

// AddJob method will add the job to the queue, this method is blocking since if
// the queue is full then it will wait till a job is finished
func (t *TaskQueue) AddJob(job Job) {
	t.wg.Add(1)
	t.queue <- job
}

// Shutdown method will wait for all the go routines associated with the tracker to
// complete or context to expire
func (t *TaskQueue) Shutdown(ctx context.Context) error {

	// Create a channel to signal when the waitgroup is finished.
	ch := make(chan struct{}, 1)

	// Create a goroutine to wait for all other goroutines to
	// be done then close the channel to unblock the select.
	go func() {
		t.wg.Wait()
		close(t.queue)
		for i := 0; i < t.qlen; i++ {
			t.quitChan <- true
		}
		ch <- struct{}{}
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
