package handlers

import (
	"github.com/ds-test-framework/scheduler/testlib"
	"github.com/ds-test-framework/scheduler/types"
)

type HandlerFunc func(*types.Event, *testlib.Context) ([]*types.Message, bool)

type GenericHandler struct {
	handlerFunc HandlerFunc
}

func NewGenericHandler(f HandlerFunc) *GenericHandler {
	return &GenericHandler{
		handlerFunc: f,
	}
}

func (g *GenericHandler) HandleEvent(e *types.Event, c *testlib.Context) []*types.Message {
	m, ok := g.handlerFunc(e, c)
	if !ok {
		return []*types.Message{}
	}
	return m
}

func (g *GenericHandler) Name() string {
	return "GenericHandler"
}
