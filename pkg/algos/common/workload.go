package common

import "github.com/ds-test-framework/scheduler/pkg/types"

// WorkloadInjector allows for protocol specific workloadinjection
type WorkloadInjector interface {
	SetPeerStore(*PeerStore)
	InjectWorkLoad() *types.Error
}

// Dummy that does nothing for noe
type CommonWorkloadInjector struct{}

func NewCommonWorkloadInjector() *CommonWorkloadInjector {
	return &CommonWorkloadInjector{}
}

func (c *CommonWorkloadInjector) InjectWorkLoad() *types.Error { return nil }
func (c *CommonWorkloadInjector) SetPeerStore(_ *PeerStore)    {}
