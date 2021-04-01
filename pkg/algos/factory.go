package algos

import (
	"github.com/spf13/viper"

	"github.com/ds-test-framework/scheduler/pkg/algos/common"
	"github.com/ds-test-framework/scheduler/pkg/algos/raft"
	"github.com/ds-test-framework/scheduler/pkg/types"
)

const (
	ErrInvalidAlgo = "INVALID_ALGO"
)

func GetAlgoDriver(options *viper.Viper) (types.AlgoDriver, *types.Error) {
	switch options.GetString("type") {
	case "raft":
		return raft.NewRaftDriver(options), nil
	case "common":
		h, err := GetWorkloadInjector(options)
		if err != nil {
			return nil, err
		}
		return common.NewCommonDriver(options, h), nil
	default:
		return nil, types.NewError(
			ErrInvalidAlgo,
			"Invalid algorithm type",
		)
	}
}

func GetWorkloadInjector(options *viper.Viper) (common.WorkloadInjector, *types.Error) {
	switch options.GetString("algo") {
	case "common":
		return common.NewCommonWorkloadInjector(), nil
	default:
		return nil, types.NewError(
			ErrInvalidAlgo,
			"Invalid directive handler type",
		)
	}
}
