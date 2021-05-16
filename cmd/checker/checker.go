package checker

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ds-test-framework/scheduler/pkg/algos"
	"github.com/ds-test-framework/scheduler/pkg/algos/common"
	"github.com/ds-test-framework/scheduler/pkg/apiserver"
	"github.com/ds-test-framework/scheduler/pkg/log"
	"github.com/ds-test-framework/scheduler/pkg/strategies"
	"github.com/ds-test-framework/scheduler/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	configPath string
	verbose    bool
)

type Checker struct {
	engine    types.StrategyEngine
	driver    types.AlgoDriver
	ctx       *types.Context
	apiServer *apiserver.APIServer
	stopChan  chan bool
}

func NewChecker(ctx *types.Context) *Checker {
	checker := &Checker{
		ctx:      ctx,
		stopChan: make(chan bool, 2),
	}
	apiS := apiserver.NewAPIServer(ctx)
	engine, err := strategies.GetStrategyEngine(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}
	workload, err := algos.GetWorkloadInjector(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}
	driver := common.NewCommonDriver(ctx, workload)

	checker.engine = engine
	checker.driver = driver
	checker.apiServer = apiS

	apiS.Start()
	driver.Start()
	go engine.Start()

	return checker
}

func (c *Checker) Run(ctx context.Context) {
	go func() {
		<-ctx.Done()
		c.Stop()
	}()
	c.run()
}

func (c *Checker) run() {

	ok, err := c.driver.Ready()
	if err != nil {
		c.Stop()
		return
	}
	if !ok {
		return
	}
	log.Debug("Driver ready")

	config := c.ctx.Config("run")

	runs := config.GetInt("run.runs")
	runTime := config.GetInt("run.time")
	for i := 0; i < runs; i++ {
		c.ctx.SetRun(i)
		runObj, err := c.driver.StartRun(i)
		if err != nil {
			c.Stop()
			log.Fatal("Error starting run: " + err.Error())
			return
		}
		log.Debug(fmt.Sprintf("Started run %d", i))
		select {
		case <-c.stopChan:
			return
		case <-time.After(time.Duration(runTime) * time.Second):
		case <-runObj.Ch:
		}
		c.driver.StopRun()
		c.engine.Reset()
	}
	c.Stop()
}

func (c *Checker) Stop() {
	log.Debug("Stopping checker")
	c.apiServer.Stop()
	c.engine.Stop()
	c.driver.Stop()
	close(c.stopChan)
}

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

			checker := NewChecker(context)
			checker.Run(ctx)

			log.Destroy()
			return nil
		},
	}
	cmd.PersistentFlags().StringVar(&configPath, "config", "", "Path to directory containing config file")
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Display verbose output")
	cmd.MarkPersistentFlagRequired("config")
	return cmd
}
