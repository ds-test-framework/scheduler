package common

import "github.com/ds-test-framework/scheduler/pkg/types"

// WorkloadInjector allows for protocol specific workloadinjection
type WorkloadInjector interface {
	InjectWorkLoad() *types.Error
}
