package testlib

import (
	"github.com/ds-test-framework/scheduler/apiserver"
	"github.com/ds-test-framework/scheduler/config"
	"github.com/ds-test-framework/scheduler/context"
	"github.com/ds-test-framework/scheduler/dispatcher"
	"github.com/ds-test-framework/scheduler/log"
	"github.com/ds-test-framework/scheduler/types"
)

type TestingServer struct {
	apiserver  *apiserver.APIServer
	dispatcher *dispatcher.Dispatcher
	ctx        *context.RootContext
	messagesCh chan *types.Message
	eventCh    chan *types.Event

	testCases      []*TestCase
	executionState *executionState
	reportStore    *TestCaseReportStore
	*types.BaseService
}

func NewTestingServer(config *config.Config, testcases []*TestCase) (*TestingServer, error) {
	log.Init(config.LogConfig)
	ctx := context.NewRootContext(config, log.DefaultLogger)

	server := &TestingServer{
		apiserver:      nil,
		dispatcher:     dispatcher.NewDispatcher(ctx),
		ctx:            ctx,
		messagesCh:     ctx.MessageQueue.Subscribe("testingServer"),
		eventCh:        ctx.EventQueue.Subscribe("testingServer"),
		testCases:      testcases,
		executionState: newExecutionState(),
		reportStore:    NewTestCaseReportStore(config.ReportStoreConfig),
		BaseService:    types.NewBaseService("TestingServer", log.DefaultLogger),
	}

	server.apiserver = apiserver.NewAPIServer(ctx, server)

	for _, t := range testcases {
		t.Logger = server.Logger.With(log.LogParams{"testcase": t.Name})
	}
	return server, nil
}

func (srv *TestingServer) Start() {
	srv.StartRunning()
	srv.apiserver.Start()
	srv.ctx.Start()
	srv.execute()
}

func (srv *TestingServer) poll() {
	for {
		select {
		case m := <-srv.messagesCh:
			err := srv.dispatcher.DispatchMessage(m)
			if err != nil {
				srv.Logger.With(log.LogParams{
					"message_id": m.ID,
					"err":        err.Error(),
				}).Error("error dispatching message")
			}
		case <-srv.QuitCh():
			return
		}
	}
}

func (srv *TestingServer) Stop() {
	srv.StopRunning()
	srv.apiserver.Stop()
	srv.ctx.Stop()
}
