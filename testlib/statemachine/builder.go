package statemachine

import "github.com/ds-test-framework/scheduler/testlib/handlers"

// StateMachineBuilder implements a builder pattern for creating the testcase state machine
type StateMachineBuilder struct {
	stateMachine *StateMachine
	curState     *State
}

// On can be used to create a transition relation between states based on the specified condition
func (s StateMachineBuilder) On(cond handlers.Condition, stateLabel string) StateMachineBuilder {
	next, ok := s.stateMachine.getState(stateLabel)
	if !ok {
		next = s.stateMachine.newState(stateLabel)
	}
	s.curState.Transitions[next.Label] = cond
	return StateMachineBuilder{
		stateMachine: s.stateMachine,
		curState:     next,
	}
}

func (s StateMachineBuilder) MarkSuccess() StateMachineBuilder {
	s.curState.Success = true
	return s
}
