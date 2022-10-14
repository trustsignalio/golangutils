package workerqueue

// Worker struct holds the information regarding the worker
type Worker struct {
	ID         int
	JobChannel chan Job
	WorkerPool chan chan Job
	QuitChan   chan bool
}

// NewWorker method will create a worker object and return it
func NewWorker(id int, workerPool chan chan Job) *Worker {
	worker := &Worker{
		ID:         id,
		JobChannel: make(chan Job),
		WorkerPool: workerPool,
		QuitChan:   make(chan bool)}
	return worker
}

func (w *Worker) startWorker() {
	for {
		// Add ourselves to the work queue
		w.WorkerPool <- w.JobChannel

		select {
		case work := <-w.JobChannel:
			// Receive a work request
			work.Process()

		case <-w.QuitChan:
			// We have been asked to stop the processing
			return
		}
	}
}

// Start method will start the worker
func (w *Worker) Start() {
	go w.startWorker()
}

// Stop method will stop the worker
func (w *Worker) Stop() {
	go func() {
		// Send a stop request in quit channel
		w.QuitChan <- true
	}()
}
