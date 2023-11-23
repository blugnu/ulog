package ulog

import (
	"path"
	"testing"

	"github.com/blugnu/test"
)

// This test calls callsite() which will return the first runtime
// frame that is not in the ulog package.  Since this test is
// itself part of that package, the result will be the function
// in the go test runner that calls this function.
//
// We can't know the exact line number of that function, so
// for the will make a reasonable expectation that it is at least
// a positive integer.
//
// We can be a bit more confident of the expected filename
// and function, though even these could conceivably change
// in a future iteration of the go test tooling.  This should
// be the first thing to check in the event of this test suddenly
// starting to fail following an update of the tool chain.
func TestCallsite(t *testing.T) {
	// ACT
	got := caller()

	// ASSERT
	if got.line <= 0 {
		t.Errorf("expected line > 0, got %d", got.line)
	}

	t.Run("file", func(t *testing.T) {
		_, got := path.Split(got.file)
		test.Equal(t, "testing.go", got)
	})
	t.Run("function", func(t *testing.T) {
		test.Equal(t, "testing.tRunner", got.function)
	})

	t.Run("fails", func(t *testing.T) {
		// ARRANGE

		// we force a failure by setting the minimum caller depth
		// to a value that is greater than the number of frames
		// in the stack
		og := ulogframes
		defer func() { ulogframes = og }()
		ulogframes = 1000

		// ACT
		got := caller()

		// ASSERT
		if got != nil {
			t.Errorf("expected nil, got %#v", got)
		}
	})
}
