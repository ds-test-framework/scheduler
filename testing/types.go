package testing

import (
	"errors"
	"time"

	"github.com/ds-test-framework/scheduler/log"
	"github.com/ds-test-framework/scheduler/types"
)

type TestCaseCtx interface {
	Done() chan bool
	Timeout() time.Duration
}

type TestCase interface {
	Initialize(*types.ReplicaStore, *log.Logger) (TestCaseCtx, error)
	HandleMessage(*types.Message) (bool, []*types.Message)
	HandleEvent(*types.Event)
	HandleLogMessage(*types.ReplicaLog)
	Assert() error

	Name() string
}

type BaseTestCaseCtx struct {
	doneCh  chan bool
	timeout time.Duration
}

func NewBaseTestCaseCtx(timeout time.Duration) *BaseTestCaseCtx {
	return &BaseTestCaseCtx{
		timeout: timeout,
		doneCh:  make(chan bool, 1),
	}
}

// Call only once
func (ctx *BaseTestCaseCtx) SetDone() {
	close(ctx.doneCh)
}

func (ctx *BaseTestCaseCtx) Done() chan bool {
	return ctx.doneCh
}

func (ctx *BaseTestCaseCtx) Timeout() time.Duration {
	return ctx.timeout
}

type BaseTestCase struct {
	Ctx    *BaseTestCaseCtx
	Logger *log.Logger
	name   string
}

func NewBaseTestCase(name string, timeout time.Duration) *BaseTestCase {
	return &BaseTestCase{
		Ctx:  NewBaseTestCaseCtx(timeout),
		name: name,
	}
}
func (b *BaseTestCase) Initialize(_ *types.ReplicaStore, logger *log.Logger) (TestCaseCtx, error) {
	b.Logger = logger
	return b.Ctx, nil
}
func (b *BaseTestCase) HandleMessage(_ *types.Message) (bool, []*types.Message) {
	return false, []*types.Message{}
}
func (b *BaseTestCase) HandleEvent(_ *types.Event) {
}
func (b *BaseTestCase) HandleLogMessage(_ *types.ReplicaLog) {
}
func (b *BaseTestCase) Assert() error {
	return errors.New("not implemented")
}
func (b *BaseTestCase) Name() string {
	return b.name
}
