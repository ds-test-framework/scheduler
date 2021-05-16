package algos

import (
	"github.com/ds-test-framework/scheduler/pkg/algos/common"
	"github.com/ds-test-framework/scheduler/pkg/algos/tendermint"
	"github.com/ds-test-framework/scheduler/pkg/types"
)

const (
	ErrInvalidAlgo = "INVALID_ALGO"
)

func GetWorkloadInjector(ctx *types.Context) (common.WorkloadInjector, *types.Error) {
	options := ctx.Config("driver")
	switch options.GetString("algo") {
	case "tendermint":
		return tendermint.NewWorkloadInjector(ctx), nil
	default:
		return nil, types.NewError(
			ErrInvalidAlgo,
			"Invalid directive handler type",
		)
	}
}
