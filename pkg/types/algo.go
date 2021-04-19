package types

// RunObj is returned by the AlgoDriver after starting a run
type RunObj struct {
	// Ch is used to signal to the scheduler that a run has ended
	Ch chan bool
}

// AlgoDriver defines an interface for interacting with an algorithm implementation.
// Instrumenting your algorithm would mean implementing this interface in `pkg/algo/<custom_algo>`
type AlgoDriver interface {
	// Init called once before starting the testing itertions.
	Init()

	// OutChan should return the channel which will contain the incoming intercepted messages.
	OutChan() chan *MessageWrapper

	// InChan should return the channel which listens for messages that are scheduled
	InChan() chan *MessageWrapper

	// Destroy called when stopping the scheduler
	Destroy()

	// StartRun is called at the start of every iteration and returns an RunObj or error if the run failed to start.
	// If the run failed to start then the scheduler stops
	StartRun(int) (*RunObj, *Error)

	// StopRun called once at the end of every iteration
	StopRun()

	// Ready called once before the scheduler is run. Should block till the implementation is ready and return true.
	// If false is returned then the scheduler stops
	Ready() (bool, *Error)
}
