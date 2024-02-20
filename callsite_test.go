package ulog

import (
	"path"
	"testing"

	"github.com/blugnu/test"
)

// This test calls caller() which will return the first runtime
// frame that is not in the ulog package.  Since this test is
// itself part of that package, the result will be the function
// in the go test runner that calls this function.
//
// We can't know the exact line number of that function, so will
// make a reasonable expectation that it is at least a positive
// integer.
//
// We can be a bit more confident of the expected filename and
// function, though even these could conceivably change in a future
// iteration of the go test tooling.  This should be the first
// thing to check should this test start failing following a go
// update.
func TestCaller(t *testing.T) {
	// ARRANGE
	testcases := []struct {
		scenario string
		exec     func(t *testing.T)
	}{
		{scenario: "successful",
			exec: func(t *testing.T) {
				// ACT
				got := caller()

				// ASSERT
				_, filename := path.Split(got.file)
				function := got.function

				test.IsTrue(t, got.line > 0, "got.line > 0")
				test.That(t, filename).Equals("testing.go")
				test.That(t, function).Equals("testing.tRunner")
			},
		},
		{scenario: "unable to determine caller",
			exec: func(t *testing.T) {
				// ARRANGE

				// we force a failure by setting the number of ulog frames to
				// a value that is greater than the number of frames in the
				// entire call stack
				og := ulogframes
				defer func() { ulogframes = og }()
				ulogframes = 1000

				// ACT
				got := caller()

				// ASSERT
				test.That(t, got).IsNil()
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.exec(t)
		})
	}
}
