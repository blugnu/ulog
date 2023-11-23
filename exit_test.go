package ulog

import "testing"

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
	exit(1)

	// ASSERT
	t.Run("calls ExitFn", func(t *testing.T) {
		wanted := true
		got := exitWasCalled
		if wanted != got {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}

		t.Run("with exit code", func(t *testing.T) {
			wanted := 1
			got := exitCode
			if wanted != got {
				t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
			}
		})
	})
}
