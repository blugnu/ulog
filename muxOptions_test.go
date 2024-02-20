package ulog

import (
	"errors"
	"testing"

	"github.com/blugnu/test"
)

func TestMuxFormat(t *testing.T) {
	// ARRANGE
	testcases := []struct {
		scenario string
		exec     func(t *testing.T)
	}{
		{scenario: "formats registered successfully",
			exec: func(t *testing.T) {
				// ARRANGE
				mx := &mux{formats: map[string]*formatref{}}

				// ACT
				err1 := MuxFormat("id1", func() (Formatter, error) { return &mockformatter{}, nil })(mx)
				err2 := MuxFormat("id2", func() (Formatter, error) { return &mockformatter{}, nil })(mx)

				// ASSERT
				test.Error(t, err1).IsNil()
				test.Error(t, err2).IsNil()
				test.That(t, mx.formats).Equals(map[string]*formatref{
					"id1": {0, &mockformatter{}},
					"id2": {1, &mockformatter{}},
				})
			},
		},
		{scenario: "duplicate id",
			exec: func(t *testing.T) {
				// ARRANGE
				mx := &mux{formats: map[string]*formatref{}}
				_ = MuxFormat("id", func() (Formatter, error) { return &mockformatter{}, nil })(mx)

				// ACT
				err := MuxFormat("id", func() (Formatter, error) { return &mockformatter{}, nil })(mx)

				// ASSERT
				test.Error(t, err).Is(ErrFormatAlreadyRegistered)
			},
		},
		{scenario: "factory error",
			exec: func(t *testing.T) {
				// ARRANGE
				mx := &mux{formats: map[string]*formatref{}}
				facterr := errors.New("factory error")

				// ACT
				err := MuxFormat("id", func() (Formatter, error) { return nil, facterr })(mx)

				// ASSERT
				test.Error(t, err).Is(facterr)
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.exec(t)
		})
	}
}
