package testlib

import (
	"testing"
	"time"

	"github.com/ds-test-framework/scheduler/types"
)

func TestBuilderPattern(t *testing.T) {
	testcase := NewTestCase("dummy", 2*time.Second)
	builder := testcase.Builder()

	dummyCond := func(c *Context) bool {
		return false
	}

	dummyAction := func(c *Context) []*types.Message {
		return []*types.Message{}
	}

	new := builder.
		On(dummyCond, "new").Action(dummyAction)

	new.On(dummyCond, "new2").Action(dummyAction)

	new.
		On(dummyCond, "new1").Action(dummyAction).
		On(dummyCond, testcase.Fail().Label)
}
