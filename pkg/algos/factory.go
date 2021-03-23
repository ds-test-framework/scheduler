package algos

import (
	"github.com/spf13/viper"

	"github.com/zeu5/model-checker/pkg/algos/raft"
	"github.com/zeu5/model-checker/pkg/types"
)

const (
	ERR_INVALID_ALGO = "INVALID_ALGO"
)

func GetAlgoDriver(options *viper.Viper) (types.AlgoDriver, *types.Error) {
	switch options.GetString("type") {
	case "raft":
		return raft.NewRaftDriver(options), nil
	default:
		return nil, types.NewError(
			ERR_INVALID_ALGO,
			"Invalid algorithm type",
		)
	}
}
