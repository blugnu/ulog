package ulog

import "testing"

func TestLevelsString(t *testing.T) {
	testcases := []struct {
		name  string
		level Level
	}{
		{name: "Trace", level: TraceLevel},
		{name: "Debug", level: DebugLevel},
		{name: "Info", level: InfoLevel},
		{name: "Warn", level: WarnLevel},
		{name: "Error", level: ErrorLevel},
		{name: "Fatal", level: FatalLevel},
		{name: "<not set>", level: Level(0)},
		{name: "<invalid level (-1)>", level: Level(-1)},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// ACT
			got := tc.level.String()

			// ASSERT
			wanted := tc.name
			if wanted != got {
				t.Errorf("wanted %v, got %v", wanted, got)
			}
		})
	}
}
