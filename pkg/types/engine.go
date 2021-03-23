package types

type StrategyEngine interface {
	Reset()
	Run() *Error
	Stop()
	SetChannels(chan *MessageWrapper, chan *MessageWrapper)
}
