package testlib

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sync"

	"github.com/ds-test-framework/scheduler/config"
	"github.com/ds-test-framework/scheduler/types"
)

// TestCaseReport stores all the necesary information about the execution of the testcase
type TestCaseReport struct {
	// OutgoingMessages stores the messages delivered in an ordered list
	OutgoingMessages []*types.Message `json:"messages"`
	// EventDAG stores the events as a directed acyclic graph
	EventDAG *types.EventDAG `json:"event_dag"`
	// Run stores the states that were transitioned during the execution of the state machine
	Run []*State `json:"run"`
	// FinalState of the state machine
	FinalState *State `json:"final_state"`
	// Assertion stores the response of the assert call of the testcase
	Assertion bool `json:"assert"`
	// TestCase name of the testcase
	TestCase string `json:"testcase"`

	lock *sync.Mutex `json:"-"`
}

// NewTestCaseReport instantiates the testcase report structure
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

// AddOutgoingMessages adds to the OutgoingMessages attribute
func (r *TestCaseReport) AddOutgoingMessages(messages []*types.Message) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.OutgoingMessages = append(r.OutgoingMessages, messages...)
}

// AddStateTransition adds the state to the Run attribute
func (r *TestCaseReport) AddStateTransition(s *State) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.Run = append(r.Run, s)
}

// TestCaseReportStore stores the reports as a map with testcase names as keys
type TestCaseReportStore struct {
	reports   map[string]*TestCaseReport
	lock      *sync.Mutex
	storePath string
}

// NewTestCaseReportStore intantitates a new TestCaseReportStore with the specified config
func NewTestCaseReportStore(conf config.ReportStoreConfig) *TestCaseReportStore {
	return &TestCaseReportStore{
		reports:   make(map[string]*TestCaseReport),
		lock:      new(sync.Mutex),
		storePath: conf.Path,
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

// Save stores the report in the configured path
func (s *TestCaseReportStore) Save() error {
	info, err := os.Stat(s.storePath)
	if err != nil {
		return fmt.Errorf("could not open specified path: %s", err)
	}
	var file *os.File
	if info.IsDir() {
		file, err = os.Create(path.Join(s.storePath, "report.json"))
		if err != nil {
			return fmt.Errorf("could not create report file: %s", err)
		}
	} else {
		file, err = os.Create(s.storePath)
		if err != nil {
			return fmt.Errorf("could not open the report file: %s", err)
		}
	}
	s.lock.Lock()
	reportsS, err := json.Marshal(s.reports)
	if err != nil {
		return fmt.Errorf("failed to marshal report: %s", err)
	}
	s.lock.Unlock()
	_, err = file.Write(reportsS)
	if err != nil {
		return fmt.Errorf("failed to write report to file: %s", err)
	}
	if err = file.Close(); err != nil {
		return fmt.Errorf("failed to close report file: %s", err)
	}
	return nil
}
