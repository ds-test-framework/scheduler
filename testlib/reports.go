package testlib

import (
	"sync"

	"github.com/ds-test-framework/scheduler/config"
	"github.com/ds-test-framework/scheduler/types"
)

type TestCaseReport struct {
	OutgoingMessages []*types.Message
	EventDAG         *types.EventDAG
	Run              []*State
	FinalState       *State
	Assertion        bool
	TestCase         string

	lock *sync.Mutex
}

func NewTestCaseReport(testcase string) *TestCaseReport {
	return &TestCaseReport{
		OutgoingMessages: make([]*types.Message, 0),
		EventDAG:         nil,
		Run:              make([]*State, 0),
		FinalState:       nil,
		Assertion:        false,
		TestCase:         testcase,

		lock: new(sync.Mutex),
	}
}

func (r *TestCaseReport) AddOutgoingMessages(messages []*types.Message) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.OutgoingMessages = append(r.OutgoingMessages, messages...)
}

func (r *TestCaseReport) AddStateTransition(s *State) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.Run = append(r.Run, s)
}

type TestCaseReportStore struct {
	reports   map[string]*TestCaseReport
	lock      *sync.Mutex
	storePath string
}

func NewTestCaseReportStore(conf config.ReportStoreConfig) *TestCaseReportStore {
	return &TestCaseReportStore{
		reports:   make(map[string]*TestCaseReport),
		lock:      new(sync.Mutex),
		storePath: conf.Path,
	}
}

func (s *TestCaseReportStore) AddReport(r *TestCaseReport) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.reports[r.TestCase] = r
}

func (s *TestCaseReportStore) Save() {

}
