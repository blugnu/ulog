package ulog

import (
	"testing"

	"github.com/blugnu/test"
)

func Test_exit(t *testing.T) {
	// ARRANGE
	exitWasCalled := false
	exitCode := 0

	og := ExitFn
	defer func() { ExitFn = og }()
	ExitFn = func(code int) {
		exitWasCalled = true
		exitCode = code
	}

	// ACT
	exit(42)

	// ASSERT
	test.IsTrue(t, exitWasCalled, "calls exit func")
	test.Value(t, exitCode, "exit code set").Equals(42)
}
