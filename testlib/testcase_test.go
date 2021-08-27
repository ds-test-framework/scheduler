package testlib

import (
	"testing"
	"time"

	"github.com/ds-test-framework/scheduler/types"
)

func TestBuilderPatter(t *testing.T) {
	testcase := NewTestCase("dummy", 2*time.Second)
	builder := testcase.Builder()

	dummyCond := func(c *Context) bool {
		return false
	}

	dummyAction := func(c *Context) []*types.Message {
		return []*types.Message{}
	}

	new := builder.
		On(dummyCond, "new").Do(dummyAction)
	new.
		On(dummyCond, "new1").Do(dummyAction).
		On(dummyCond, testcase.Fail().Label)
}
