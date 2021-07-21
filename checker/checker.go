package checker

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/ds-test-framework/scheduler/algos"
	"github.com/ds-test-framework/scheduler/algos/common"
	"github.com/ds-test-framework/scheduler/apiserver"
	"github.com/ds-test-framework/scheduler/log"
	"github.com/ds-test-framework/scheduler/strategies"
	"github.com/ds-test-framework/scheduler/types"
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
		log.With(map[string]interface{}{
			"error": err.Error(),
		}).Error("Error with driver getting ready")
		c.Stop()
		return
	}
	if !ok {
		return
	}
	log.Debug("Driver ready")

	config := c.ctx.Config("run")
	runs := config.GetInt("runs")
	runTime := config.GetInt("time")
	log.With(map[string]interface{}{
		"no_runs":   strconv.Itoa(runs),
		"wait_time": strconv.Itoa(runTime),
	}).Debug("Starting testing loop")
	for i := 0; i < runs; i++ {
		c.ctx.SetRun(&types.Run{
			Id:    i,
			Label: fmt.Sprintf("run_%d", i),
		})
		runObj, err := c.driver.StartRun(i)
		if err != nil {
			log.With(map[string]interface{}{
				"error": err.Error(),
			}).Error("Error starting run")
			c.Stop()
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
