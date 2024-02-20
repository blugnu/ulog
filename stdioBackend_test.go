package ulog

import (
	"bytes"
	"io"
	"testing"

	"github.com/blugnu/test"
)

func TestStdioBackend(t *testing.T) {
	// ARRANGE
	testcases := []struct {
		scenario string
		exec     func(t *testing.T)
	}{
		// init
		{scenario: "newStdioBackend",
			exec: func(t *testing.T) {
				// ACT
				var f Formatter = &mockformatter{}
				var w io.Writer = &bytes.Buffer{}
				result := newStdioBackend(f, w)

				// ASSERT
				test.Value(t, result.Formatter).Equals(f)
				test.Value(t, result.Writer).Equals(w)
				test.That(t, result.bufs).IsNotNil()
				test.IsType[*bytes.Buffer](t, result.bufs.Get())
			},
		},

		// dispatch tests
		{scenario: "dispatch",
			exec: func(t *testing.T) {
				// ARRANGE
				buf := &bytes.Buffer{}
				sut := newStdioBackend(&mockformatter{}, buf)

				// ACT
				sut.dispatch(entry{Message: "test"})

				// ASSERT
				test.Value(t, buf.String()).Equals("test\n")
			},
		},

		// SetFormatter tests
		{scenario: "SetFormatter",
			exec: func(t *testing.T) {
				// ARRANGE
				sut := &stdioBackend{}
				f := &mockformatter{}

				// ACT
				err := sut.SetFormatter(f)

				// ASSERT
				test.Error(t, err).IsNil()
				test.Value(t, sut.Formatter).Equals(f)
			},
		},

		// SetOutput tests
		{scenario: "SetOutput",
			exec: func(t *testing.T) {
				// ARRANGE
				sut := &stdioBackend{}
				w := &bytes.Buffer{}

				// ACT
				err := sut.SetOutput(w)

				// ASSERT
				test.Error(t, err).IsNil()
				test.Value(t, sut.Writer).Equals(w)
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.exec(t)
		})
	}
}
