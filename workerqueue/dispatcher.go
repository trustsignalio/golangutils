package workerqueue

// Dispatcher struct contains the necessary data to spawn the workers and
// start each worker, it contains a worker pool channel of fixed size ie buffered
type Dispatcher struct {
	maxWorkers int
	workers    []*Worker
	workerPool chan chan Job
}

// NewDispatcher method will return a dispatcher object
func NewDispatcher(maxWorkers int) *Dispatcher {
	pool := make(chan chan Job, maxWorkers)
	dis := Dispatcher{workerPool: pool, maxWorkers: maxWorkers}
	return &dis
}

// Run method will start the workers with the given jobPool which will be
// a buffered channel
func (d *Dispatcher) Run(jobPool chan Job) {
	for i := 0; i < d.maxWorkers; i++ {
		worker := NewWorker(i, d.workerPool)
		worker.Start()
		d.workers = append(d.workers, worker)
	}
	go d.dispatch(jobPool)
}

// Stop method will stop the dispatcher and its workers
func (d *Dispatcher) Stop() {
	for _, w := range d.workers {
		w.Stop()
	}
	close(d.workerPool)
}

func (d *Dispatcher) dispatch(jobPool chan Job) {
	for {
		select {
		case job := <-jobPool: // received a job
			go d.sendJobToWorker(job)
		}
	}
}

func (d *Dispatcher) sendJobToWorker(job Job) {
	if job == nil {
		return
	}
	// Get a worker from the worker pool
	worker := <-d.workerPool

	// send job to the worker for processing
	worker <- job
}
