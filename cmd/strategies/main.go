package strategies

import (
	"fmt"

	"github.com/ds-test-framework/scheduler/config"
	"github.com/ds-test-framework/scheduler/context"
	"github.com/ds-test-framework/scheduler/log"
	"github.com/ds-test-framework/scheduler/strategies"
	"github.com/ds-test-framework/scheduler/util"
	"github.com/spf13/cobra"
)

func StrategiesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "strategy [strategy_name]",
		Short: "Run the scheduler with the specified strategy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			termCh := util.Term()

			conf, err := config.ParseConfig(config.ConfigPath)
			if err != nil {
				return fmt.Errorf("failed to parse config: %s", err)
			}
			log.Init(conf.LogConfig)
			ctx := context.NewRootContext(conf, log.DefaultLogger)

			strategy, err := strategies.GetStrategy(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to initialize strategy: %s", err)
			}

			ctx.Start()
			strategy.Start()

			<-termCh
			strategy.Stop()
			ctx.Stop()
			return nil
		},
	}
}
