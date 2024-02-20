package ulog

import (
	"testing"

	"github.com/blugnu/test"
)

func TestLevelsString(t *testing.T) {
	testcases := []struct {
		name  string
		level Level
	}{
		{name: "trace", level: TraceLevel},
		{name: "debug", level: DebugLevel},
		{name: "info", level: InfoLevel},
		{name: "warn", level: WarnLevel},
		{name: "ERROR", level: ErrorLevel},
		{name: "FATAL", level: FatalLevel},
		{name: "<not set>", level: Level(0)},
		{name: "<invalid level (-1)>", level: Level(-1)},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// ACT
			result := tc.level.String()

			// ASSERT
			test.Value(t, result).Equals(tc.name)
		})
	}
}
