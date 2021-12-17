package handlers

import (
	"github.com/ds-test-framework/scheduler/testlib"
	"github.com/ds-test-framework/scheduler/types"
)

// HandlerFunc type to define a conditional handler
// returns false in the second return value if the handler is not concerned about the event
type HandlerFunc func(*types.Event, *testlib.Context) ([]*types.Message, bool)

// GenericHandler constructs a Handler with just a single HandlerFunc
type GenericHandler struct {
	handlerFunc HandlerFunc
}

// NewGeneticHandler creates a new GenericHandler
func NewGenericHandler(f HandlerFunc) *GenericHandler {
	return &GenericHandler{
		handlerFunc: f,
	}
}

// HandleEvent implements Handler
func (g *GenericHandler) HandleEvent(e *types.Event, c *testlib.Context) []*types.Message {
	m, ok := g.handlerFunc(e, c)
	if !ok {
		return []*types.Message{}
	}
	return m
}

// Name implements Handler
func (g *GenericHandler) Name() string {
	return "GenericHandler"
}
