package cmd

import (
	"github.com/ds-test-framework/scheduler/cmd/visualizer"
	"github.com/ds-test-framework/scheduler/config"
	"github.com/spf13/cobra"
)

func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scheduler",
		Short: "Scheduler tool to test/understand consensus algorithms",
	}
	cmd.PersistentFlags().StringVarP(&config.ConfigPath, "config", "c", "config.json", "Config file path")
	cmd.AddCommand(visualizer.VisualizerCmd())
	return cmd
}
