package common

import "github.com/ds-test-framework/model-checker/pkg/types"

type WorkloadInjector interface {
	SetPeerStore(*PeerStore)
	InjectWorkLoad() *types.Error
}

type CommonWorkloadInjector struct{}

func NewCommonWorkloadInjector() *CommonWorkloadInjector {
	return &CommonWorkloadInjector{}
}

func (c *CommonWorkloadInjector) InjectWorkLoad() *types.Error { return nil }
func (c *CommonWorkloadInjector) SetPeerStore(_ *PeerStore)    {}
