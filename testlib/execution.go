package testlib

import (
	"sync"
	"time"

	"github.com/ds-test-framework/scheduler/context"
	"github.com/ds-test-framework/scheduler/log"
	"github.com/ds-test-framework/scheduler/types"
)

type executionState struct {
	allowEvents bool
	curCtx      *Context
	report      *TestCaseReport
	testcase    *TestCase
	lock        *sync.Mutex
}

func newExecutionState() *executionState {
	return &executionState{
		allowEvents: false,
		curCtx:      nil,
		report:      nil,
		testcase:    nil,
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

func (e *executionState) NewTestCase(ctx *context.RootContext, testcase *TestCase) {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.testcase = testcase
	e.report = NewTestCaseReport(testcase.Name)
	e.curCtx = NewContext(ctx, testcase, e.report)
}

func (e *executionState) CurCtx() *Context {
	e.lock.Lock()
	defer e.lock.Unlock()
	return e.curCtx
}

func (e *executionState) CurTestCase() *TestCase {
	e.lock.Lock()
	defer e.lock.Unlock()
	return e.testcase
}

func (e *executionState) CurReport() *TestCaseReport {
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
		select {
		case <-srv.QuitCh():
			return
		default:
		}
		if srv.ctx.Replicas.Count() == srv.ctx.Config.NumReplicas {
			break
		}
	}
	srv.Logger.Info("All replicas connected.")
	go srv.pollEvents()

MainLoop:
	for _, testcase := range srv.testCases {
		testcaseLogger := testcase.Logger
		testcaseLogger.Info("Starting testcase")
		testcaseLogger.Debug("Waiting for replicas to be ready")

	ReplicaReadyLoop:
		for {
			if srv.ctx.Replicas.NumReady() == srv.ctx.Config.NumReplicas {
				break ReplicaReadyLoop
			}
		}
		testcaseLogger.Debug("Replicas are ready")

		// Setup
		testcaseLogger.Debug("Setting up testcase")
		srv.executionState.NewTestCase(srv.ctx, testcase)
		err := testcase.setup(srv.executionState.CurCtx())
		if err != nil {
			testcaseLogger.With(log.LogParams{"error": err}).Error("Error setting up testcase")
			goto Finalize
		}
		// Wait for completion or timeout
		testcaseLogger.Debug("Waiting for completion")
		srv.executionState.Unblock()
		select {
		case <-testcase.doneCh:
		case <-time.After(testcase.Timeout):
			testcaseLogger.Info("Testcase timedout")
		case <-srv.QuitCh():
			break MainLoop
		}

		// Stopping further processing of events
		srv.executionState.Block()

	Finalize:
		// Finalize report
		testcaseLogger.Info("Finalizing")
		report := srv.executionState.CurReport()
		ctx := srv.executionState.CurCtx()
		report.EventDAG = ctx.EventDAG
		if testcase.aborted {
			testcaseLogger.Info("Testcase was aborted")
			report.Assertion = false
		} else {
			testcaseLogger.Info("Checking assertion")
			report.Assertion = testcase.assert(ctx)
		}
		if !report.Assertion {
			testcaseLogger.Info("Testcase failed")
		}
		srv.ReportStore.AddReport(report)

		// TODO: Figure out a way to cleanly clear the message pool
		// Reset the servers and flush the queues after waiting for some time
		srv.ctx.Reset()
		if err := srv.dispatcher.RestartAll(); err != nil {
			srv.Logger.With(log.LogParams{"error": err}).Error("Failed to restart replicas! Aborting!")
			break MainLoop
		}
	}
	close(srv.doneCh)
}

func (srv *TestingServer) pollEvents() {
	for {
		if !srv.executionState.CanAllowEvents() {
			continue
		}
		select {
		case e := <-srv.eventCh:

			ctx := srv.executionState.CurCtx()
			testcase := srv.executionState.CurTestCase()
			report := srv.executionState.CurReport()
			testcaseLogger := srv.Logger.With(log.LogParams{"testcase": testcase.Name})

			testcaseLogger.With(log.LogParams{"event_id": e.ID, "type": e.TypeS}).Debug("Stepping")
			// 1. Add event to context and feed context to testcase
			ctx.setEvent(e)
			messages := testcase.step(ctx)

			select {
			case <-testcase.doneCh:
			default:
				report.AddOutgoingMessages(messages)
				go srv.dispatchMessages(ctx, messages)
			}
		case <-srv.QuitCh():
			return
		}
	}
}

func (srv *TestingServer) dispatchMessages(ctx *Context, messages []*types.Message) {
	for _, m := range messages {
		if !ctx.MessagePool.Exists(m.ID) {
			ctx.MessagePool.Add(m)
		}
		srv.dispatcher.DispatchMessage(m)
	}
}
