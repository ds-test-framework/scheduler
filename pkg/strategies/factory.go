package strategies

import (
	"github.com/spf13/viper"

	"github.com/zeu5/model-checker/pkg/strategies/random"
	"github.com/zeu5/model-checker/pkg/strategies/timeout"
	"github.com/zeu5/model-checker/pkg/types"
)

const (
	ERR_INVALID_STRATEGY = "INVALID_STRATEGY"
)

func GetStrategyEngine(options *viper.Viper) (types.StrategyEngine, *types.Error) {
	switch options.GetString("type") {
	case "timeout":
		return timeout.NewTimeoutEngine(options), nil
	case "random":
		return random.NewRandomScheduler(), nil
	default:
		return nil, types.NewError(
			ERR_INVALID_STRATEGY,
			"Invalid strategy type provided",
		)
	}
}
