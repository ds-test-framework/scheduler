package testlib

import (
	"github.com/ds-test-framework/scheduler/apiserver"
	"github.com/ds-test-framework/scheduler/config"
	"github.com/ds-test-framework/scheduler/context"
	"github.com/ds-test-framework/scheduler/dispatcher"
	"github.com/ds-test-framework/scheduler/log"
	"github.com/ds-test-framework/scheduler/types"
)

// TestingServer is used to run the scheduler tool for unit testing
type TestingServer struct {
	apiserver  *apiserver.APIServer
	dispatcher *dispatcher.Dispatcher
	ctx        *context.RootContext
	eventCh    chan *types.Event

	doneCh chan string

	testCases      map[string]*TestCase
	executionState *executionState
	ReportStore    *TestCaseReportStore
	*types.BaseService
}

// NewTestingServer instantiates TestingServer
// testcases are passed as arguments
func NewTestingServer(config *config.Config, testcases []*TestCase) (*TestingServer, error) {
	log.Init(config.LogConfig)
	ctx := context.NewRootContext(config, log.DefaultLogger)

	server := &TestingServer{
		apiserver:      nil,
		dispatcher:     dispatcher.NewDispatcher(ctx),
		ctx:            ctx,
		eventCh:        ctx.EventQueue.Subscribe("testingServer"),
		doneCh:         make(chan string),
		testCases:      make(map[string]*TestCase),
		executionState: newExecutionState(),
		ReportStore:    NewTestCaseReportStore(),
		BaseService:    types.NewBaseService("TestingServer", log.DefaultLogger),
	}
	for _, t := range testcases {
		server.testCases[t.Name] = t
	}

	server.apiserver = apiserver.NewAPIServer(ctx, server)

	for _, t := range testcases {
		t.Logger = server.Logger.With(log.LogParams{"testcase": t.Name})
	}
	return server, nil
}

// Start starts the TestingServer and implements Service
func (srv *TestingServer) Start() {
	srv.StartRunning()
	srv.apiserver.Start()
	srv.ctx.Start()
	srv.execute()

	// Just keep running until asked to stop
	// For dashboard purposes
	<-srv.QuitCh()
}

func (srv *TestingServer) Done() chan string {
	return srv.doneCh
}

// Stop stops the TestingServer and implements Service
func (srv *TestingServer) Stop() {
	srv.StopRunning()
	srv.apiserver.Stop()
	srv.ctx.Stop()
}
