package ulog

import (
	"testing"

	"github.com/blugnu/test"
)

func Test_entry(t *testing.T) {
	// ARRANGE
	testcases := []struct {
		scenario string
		exec     func(t *testing.T)
	}{
		{scenario: "String/no fields",
			exec: func(t *testing.T) {
				// ARRANGE
				e := entry{logcontext: &logcontext{}, Level: InfoLevel, Message: "entry"}

				// ACT
				result := e.String()

				// ASSERT
				test.That(t, result).Equals("level=info message=\"entry\" fields=[<none>]")
			},
		},
		{scenario: "String/one field",
			exec: func(t *testing.T) {
				// ARRANGE
				e := entry{logcontext: &logcontext{fields: &fields{m: map[string]any{"field": "value"}}}, Level: InfoLevel, Message: "entry"}

				// ACT
				result := e.String()

				// ASSERT
				test.That(t, result).Equals("level=info message=\"entry\" fields=[\"field\"=\"value\"]")
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.exec(t)
		})
	}
}
