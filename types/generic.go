package types

type Clonable interface {
	Clone() Clonable
}

// Generic subscriber to maintain state of the subsciber
type Subscriber struct {
	Ch chan interface{}
}
