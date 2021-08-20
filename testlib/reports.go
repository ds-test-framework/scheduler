package testlib

import (
	"sync"

	"github.com/ds-test-framework/scheduler/types"
)

type TestCaseReport struct {
	OutgoingMessages []*types.Message
	EventDAG         *types.EventDAG
	Run              []*State
	FinalState       *State
	Assertion        bool
	TestCase         string
}

func NewTestCaseReport(testcase string) *TestCaseReport {
	return &TestCaseReport{
		OutgoingMessages: make([]*types.Message, 0),
		EventDAG:         nil,
		Run:              make([]*State, 0),
		FinalState:       nil,
		Assertion:        false,
		TestCase:         testcase,
	}
}

type TestCaseReportStore struct {
	reports map[string]*TestCaseReport
	lock    *sync.Mutex
}

func NewTestCaseReportStore() *TestCaseReportStore {
	return &TestCaseReportStore{
		reports: make(map[string]*TestCaseReport),
		lock:    new(sync.Mutex),
	}
}

func (s *TestCaseReportStore) AddReport(r *TestCaseReport) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.reports[r.TestCase] = r
}
