package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/viper"

	"github.com/ds-test-framework/scheduler/pkg/algos"
	"github.com/ds-test-framework/scheduler/pkg/logger"
	"github.com/ds-test-framework/scheduler/pkg/strategies"
	"github.com/ds-test-framework/scheduler/pkg/types"
)

var (
	root *string = flag.String("config", "", "Config path")
)

type Checker struct {
	engine        types.StrategyEngine
	driver        types.AlgoDriver
	config        *viper.Viper
	engineManager *strategies.EngineManager
	stopChan      chan bool
}

func NewChecker(c *viper.Viper) *Checker {
	checker := &Checker{
		config:   c,
		stopChan: make(chan bool, 2),
	}
	engine, err := strategies.GetStrategyEngine(c.Sub("engine"))
	if err != nil {
		logger.Fatal(err.Error())
	}
	driver, err := algos.GetAlgoDriver(c.Sub("driver"))
	if err != nil {
		logger.Fatal(err.Error())
	}
	engineManager := strategies.NewEngineManager(engine, driver.OutChan(), driver.InChan())
	checker.engine = engine
	checker.driver = driver
	checker.engineManager = engineManager

	driver.Init()
	go engine.Run()
	go engineManager.Run()

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
	logger.Debug("Driver ready")

	runs := c.config.GetInt("run.runs")
	runTime := c.config.GetInt("run.time")
	for i := 0; i < runs; i++ {
		c.engineManager.FlushChannels()
		c.engineManager.SetRun(i)
		runObj, err := c.driver.StartRun(i)
		if err != nil {
			c.Stop()
			logger.Fatal("Error starting run: " + err.Error())
			return
		}
		logger.Debug(fmt.Sprintf("Started run %d", i))
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
	logger.Debug("Stopping checker")
	c.engine.Stop()
	c.engineManager.Stop()
	c.driver.Destroy()
	close(c.stopChan)
}

func main() {

	termCh := make(chan os.Signal, 1)
	signal.Notify(termCh, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		oscall := <-termCh
		log.Printf("Received syscall: %#v", oscall)
		cancel()
	}()

	flag.Parse()

	viper.AddConfigPath(*root)
	runConfig := viper.New()
	runConfig.SetDefault("runs", 1000)
	runConfig.SetDefault("time", 5)
	viper.SetDefault("run", runConfig)

	err := viper.ReadInConfig()
	if err != nil {
		logger.Fatal(err.Error())
	}

	config := viper.GetViper()
	logger.Init(config.Sub("log"))

	checker := NewChecker(config)
	checker.Run(ctx)

	logger.Destroy()
}
