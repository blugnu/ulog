package ulog

import (
	"testing"

	"github.com/blugnu/test"
)

func TestDebug(t *testing.T) {
	t.Run("no-op by default", func(t *testing.T) {
		// ACT
		_, stderr := test.CaptureOutput(t, func() {
			trace("foo")
		})

		// ASSERT
		stderr.IsEmpty()
	})

	t.Run("when trace enabled", func(t *testing.T) {
		// ARRANGE
		og := traceFn
		defer func() { traceFn = og }()

		// ACT
		EnableTrace()
		_, stderr := test.CaptureOutput(t, func() {
			trace("foo")
		})

		// ASSERT
		stderr.Contains("[ulog trace] foo")
	})
}
