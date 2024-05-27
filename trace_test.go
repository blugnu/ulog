package ulog

import (
	"context"
	"testing"

	"github.com/blugnu/test"
)

func TestDebug(t *testing.T) {
	t.Run("no-op by default", func(t *testing.T) {
		// ACT
		_, stderr := test.CaptureOutput(t, func() {
			trace(context.Background(), "foo")
		})

		// ASSERT
		stderr.IsEmpty()
	})

	t.Run("when trace enabled", func(t *testing.T) {
		// ARRANGE
		og := traceFn
		defer func() { traceFn = og }()

		// ACT
		EnableTraceLogs(nil)
		_, stderr := test.CaptureOutput(t, func() {
			trace("foo")
		})

		// ASSERT
		stderr.Contains("ULOG:TRACE foo")
	})
}
