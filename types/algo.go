package types

import "github.com/spf13/viper"

// RunObj is returned by the AlgoDriver after starting a run
type RunObj struct {
	// Ch is used to signal to the scheduler that a run has ended
	Ch chan bool
}

// AlgoDriver defines an interface for interacting with an algorithm implementation.
// Instrumenting your algorithm would mean implementing this interface in `algo/<custom_algo>`
type AlgoDriver interface {
	// Start called once before starting the testing itertions.
	Start()

	// Stop called when stopping the scheduler
	Stop()

	// StartRun is called at the start of every iteration and returns an RunObj or error if the run failed to start.
	// If the run failed to start then the scheduler stops
	StartRun(int) (*RunObj, *Error)

	// StopRun called once at the end of every iteration
	StopRun()

	// Ready called once before the scheduler is run. Should block till the implementation is ready and return true.
	// If false is returned then the scheduler stops
	Ready() (bool, *Error)
}

type FaultModel struct {
	CanDropMessages bool
	Byzantine       bool
}

func NewFaultModel(config *viper.Viper) *FaultModel {

	config.SetDefault("fault_model.byzantine", true)
	config.SetDefault("fault_model.can_drop_messages", true)

	faultModel := config.Sub("fault_model")
	return &FaultModel{
		CanDropMessages: faultModel.GetBool("can_drop_messages"),
		Byzantine:       faultModel.GetBool("byzantine"),
	}
}
