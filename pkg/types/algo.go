package types

type RunObj struct {
	Ch chan bool
}

type AlgoDriver interface {
	Init()
	OutChan() chan *MessageWrapper
	InChan() chan *MessageWrapper
	Destroy()
	StartRun(int) *RunObj
	StopRun()
}
