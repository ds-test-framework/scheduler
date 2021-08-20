package testlib

import (
	"sync"
	"time"

	"github.com/ds-test-framework/scheduler/context"
	"github.com/ds-test-framework/scheduler/log"
)

type executionState struct {
	allowEvents bool
	curCtx      *Context
	report      *TestCaseReport
	lock        *sync.Mutex
}

func newExecutionState() *executionState {
	return &executionState{
		allowEvents: false,
		curCtx:      nil,
		report:      nil,
		lock:        new(sync.Mutex),
	}
}

func (e *executionState) Block() {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.allowEvents = false
}

func (e *executionState) Unblock() {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.allowEvents = true
}

func (e *executionState) NewTestCase(ctx *context.RootContext, testcase string) {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.curCtx = NewContext(ctx)
	e.report = NewTestCaseReport(testcase)
}

func (e *executionState) CurCtx() *Context {
	e.lock.Lock()
	defer e.lock.Unlock()
	return e.curCtx
}

func (e *executionState) Report() *TestCaseReport {
	e.lock.Lock()
	defer e.lock.Unlock()
	return e.report
}

func (e *executionState) CanAllowEvents() bool {
	e.lock.Lock()
	defer e.lock.Unlock()
	return e.allowEvents
}

func (srv *TestingServer) execute() {
	srv.Logger.Info("Waiting for all replicas to connect...")
	for {
		if srv.ctx.Replicas.Count() == srv.ctx.Config.NumReplicas {
			break
		}
	}
	srv.Logger.Info("All replicas connected.")
	go srv.pollEvents()

	for _, testcase := range srv.testCases {
		testcaseLogger := srv.Logger.With(log.LogParams{"testcase": testcase.Name})
		testcase.Logger = testcaseLogger
		testcaseLogger.Info("Starting testcase")
		testcaseLogger.Debug("Waiting for replicas to be ready")
		for {
			if srv.ctx.Replicas.NumReady() == srv.ctx.Config.NumReplicas {
				break
			}
		}
		testcaseLogger.Debug("Replicas are ready")

		// Setup
		testcaseLogger.Debug("Setting up testcase")
		srv.executionState.NewTestCase(srv.ctx, testcase.Name)
		testcase.Setup(srv.executionState.CurCtx())

		// Wait for completion or timeout
		testcaseLogger.Debug("Waiting for completion")
		srv.executionState.Unblock()
		select {
		case <-testcase.doneCh:
		case <-time.After(testcase.Timeout):
		case <-srv.QuitCh():
			return
		}

		// Stopping further processing of events
		srv.executionState.Block()

		// Finalize report
		testcaseLogger.Debug("Finalizing")
		report := srv.executionState.Report()
		ctx := srv.executionState.CurCtx()
		report.FinalState = testcase.curState
		report.EventDAG = ctx.EventDAG
		report.Assertion = testcase.Assert()
		srv.reportStore.AddReport(report)

		// Reset the servers and flush the queues after waiting for some time
		//
	}

}

func (srv *TestingServer) pollEvents() {
	for {
		if !srv.executionState.CanAllowEvents() {
			continue
		}
		select {
		case <-srv.eventCh:
			// Add event to context and feed context to testcase
			// Record outgoing messages and state transitions
			// Dispatch the messages
		case <-srv.QuitCh():
			return
		}
	}
}
