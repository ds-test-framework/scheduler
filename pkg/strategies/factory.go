package strategies

import (
	"github.com/ds-test-framework/scheduler/pkg/strategies/nop"
	"github.com/ds-test-framework/scheduler/pkg/strategies/random"
	"github.com/ds-test-framework/scheduler/pkg/strategies/timeout"
	"github.com/ds-test-framework/scheduler/pkg/strategies/ttest"
	"github.com/ds-test-framework/scheduler/pkg/types"
)

const (
	ERR_INVALID_STRATEGY = "INVALID_STRATEGY"
)

func GetStrategyEngine(ctx *types.Context) (types.StrategyEngine, *types.Error) {
	switch ctx.Config("engine").GetString("type") {
	case "timeout":
		return timeout.NewTimeoutEngine(ctx), nil
	case "random":
		return random.NewRandomScheduler(ctx), nil
	case "no-op":
		return nop.NewNopScheduler(ctx), nil
	// Tendermint testing scheduler
	case "ttest":
		return ttest.NewTTestScheduler(ctx), nil
	default:
		return nil, types.NewError(
			ERR_INVALID_STRATEGY,
			"Invalid strategy type provided",
		)
	}
}
