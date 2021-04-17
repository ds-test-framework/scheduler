package types

// StrategyEngine defines an interface for a testing strategy.
// One can define their own strategy in `pkg/strategies/<custom_strategy>`
type StrategyEngine interface {
	// Reset is run at the end of every iteration
	Reset()

	// Run is called once to start the engine
	Run() *Error

	// Stop is called once at the end of testing
	Stop()

	// SetChannels is called with the incoming channel and outgoing channel respectively
	// The Engine should use these message channels to reveive from and send to
	SetChannels(chan *MessageWrapper, chan *MessageWrapper)
}
