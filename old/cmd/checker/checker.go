package checker

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ds-test-framework/scheduler/checker"
	"github.com/ds-test-framework/scheduler/log"
	"github.com/ds-test-framework/scheduler/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	configPath string
	verbose    bool
)

func CheckerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "checker",
		Short: "Checker tests a distributed system with the provided strategy",
		Long:  "Framework to test arbitrary distributed system with a range of strategies to choose from",
		RunE: func(cmd *cobra.Command, args []string) error {
			termCh := make(chan os.Signal, 1)
			signal.Notify(termCh, os.Interrupt, syscall.SIGTERM)

			ctx, cancel := context.WithCancel(context.Background())

			go func() {
				oscall := <-termCh
				log.Info(fmt.Sprintf("Received syscall: %s", oscall.String()))
				cancel()
			}()

			viper.AddConfigPath(configPath)
			runConfig := viper.New()
			runConfig.SetDefault("runs", 1000)
			runConfig.SetDefault("time", 5)
			viper.SetDefault("run", runConfig)

			err := viper.ReadInConfig()
			if err != nil {
				return errors.New("could not read config, check config path")
			}

			config := viper.GetViper()
			loglevel := "info"
			if verbose {
				loglevel = "debug"
			}
			log.Init(config.Sub("log"), loglevel)

			context := types.NewContext(config, log.DefaultLogger)

			checkerI := checker.NewChecker(context)
			checkerI.Run(ctx)

			log.Destroy()
			return nil
		},
	}
	cmd.PersistentFlags().StringVar(&configPath, "config", "", "Path to directory containing config file")
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Display verbose output")
	cmd.MarkPersistentFlagRequired("config")
	return cmd
}
