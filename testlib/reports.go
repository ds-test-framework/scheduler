package testlib

import (
	"sync"

	"github.com/ds-test-framework/scheduler/types"
)

// TestCaseReport stores all the necesary information about the execution of the testcase
type TestCaseReport struct {
	// OutgoingMessages stores the messages delivered in an ordered list
	OutgoingMessages []*types.Message `json:"messages"`
	// EventDAG stores the events as a directed acyclic graph
	EventDAG *types.EventDAG `json:"event_dag"`
	// Assertion stores the response of the assert call of the testcase
	Assertion bool `json:"assert"`
	// TestCase name of the testcase
	TestCase string `json:"testcase"`
	// Log of messages
	Log *ReportLog `json:"log"`

	lock *sync.Mutex `json:"-"`
}

type ReportLog struct {
	Messages []ReportMessage `json:"messages"`
	lock     *sync.Mutex
}

func NewReportLog() *ReportLog {
	return &ReportLog{
		Messages: make([]ReportMessage, 0),
		lock:     new(sync.Mutex),
	}
}

func (l *ReportLog) AddMessage(message string, params map[string]interface{}) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.Messages = append(l.Messages, ReportMessage{})
}

type ReportMessage struct {
	Message string                 `json:"message"`
	Params  map[string]interface{} `json:"params"`
}

// NewTestCaseReport instantiates the testcase report structure
func NewTestCaseReport(testcase string) *TestCaseReport {
	return &TestCaseReport{
		OutgoingMessages: make([]*types.Message, 0),
		EventDAG:         nil,
		Assertion:        false,
		TestCase:         testcase,

		lock: new(sync.Mutex),
	}
}

// AddOutgoingMessages adds to the OutgoingMessages attribute
func (r *TestCaseReport) AddOutgoingMessages(messages []*types.Message) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.OutgoingMessages = append(r.OutgoingMessages, messages...)
}

// TestCaseReportStore stores the reports as a map with testcase names as keys
type TestCaseReportStore struct {
	reports map[string]*TestCaseReport
	lock    *sync.Mutex
}

// NewTestCaseReportStore intantitates a new TestCaseReportStore with the specified config
func NewTestCaseReportStore() *TestCaseReportStore {
	return &TestCaseReportStore{
		reports: make(map[string]*TestCaseReport),
		lock:    new(sync.Mutex),
	}
}

// AddReport adds a report to the store
func (s *TestCaseReportStore) AddReport(r *TestCaseReport) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.reports[r.TestCase] = r
}

// Fetch the report with the testcase name
// Second return argument is false if the report does not exist
func (s *TestCaseReportStore) GetReport(name string) (*TestCaseReport, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()
	r, ok := s.reports[name]
	return r, ok
}
