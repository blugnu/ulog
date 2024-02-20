package ulog

import (
	"bytes"
	"testing"

	"github.com/blugnu/test"
)

func TestStdioTransport(t *testing.T) {
	// ARRANGE
	testcases := []struct {
		scenario string
		exec     func(t *testing.T)
	}{
		{scenario: "Stdio factory returns a transport",
			exec: func(t *testing.T) {
				// ARRANGE
				w := &bytes.Buffer{}

				// ACT
				result, err := StdioTransport(w)()

				// ASSERT
				test.Error(t, err).IsNil()
				if result, ok := test.IsType[*stdioTransport](t, result); ok {
					test.Value(t, result.Writer).Equals(w)
				}
			},
		},
		{scenario: "log writes to the writer",
			exec: func(t *testing.T) {
				// ARRANGE
				w := &bytes.Buffer{}
				sut := &stdioTransport{}
				sut.init(w)

				// ACT
				sut.log([]byte("test"))

				// ASSERT
				test.Value(t, w.String()).Equals("test\n")
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.exec(t)
		})
	}
}
