package testing

import (
	"errors"
	"time"

	"github.com/ds-test-framework/scheduler/types"
)

type TestCaseCtx interface {
	Done() chan bool
	Timeout() time.Duration
}

type TestCase interface {
	Initialize(*types.ReplicaStore) (TestCaseCtx, error)
	HandleMessage(*types.Message) (bool, []*types.Message)
	HandleStateUpdate(*types.StateUpdate)
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
	Ctx  *BaseTestCaseCtx
	name string
}

func NewBaseTestCase(name string, timeout time.Duration) *BaseTestCase {
	return &BaseTestCase{
		Ctx:  NewBaseTestCaseCtx(timeout),
		name: name,
	}
}
func (b *BaseTestCase) Initalize(_ *types.ReplicaStore) TestCaseCtx {
	return b.Ctx
}
func (b *BaseTestCase) HandleMessage(_ *types.Message) (bool, []*types.Message) {
	return false, []*types.Message{}
}
func (b *BaseTestCase) HandleStateUpdate(_ *types.StateUpdate) {
}
func (b *BaseTestCase) HandleLogMessage(_ *types.ReplicaLog) {
}
func (b *BaseTestCase) Assert() error {
	return errors.New("not implemented")
}
func (b *BaseTestCase) Name() string {
	return b.name
}
