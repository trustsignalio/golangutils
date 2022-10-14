package workerqueue

// Job interface
type Job interface {
	Process() bool
}
