package dispatcher

// Job defines an interface that allows different types of jobs to be implemented with their own processing logic
type Job interface {
	Process() error
}
