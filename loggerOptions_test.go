package ulog

import (
	"errors"
	"io"
	"testing"

	"github.com/blugnu/test"
)

// LoggerBackend returns a function that sets the backend of a logger.
//
// This option is only for use in tests when replacing the default
// stdio backend with a backend implementation with behaviour required
// to exercise specific test scenarios.
func LoggerBackend(t dispatcher) LoggerOption {
	return func(l *logger) error {
		l.backend = t
		return nil
	}
}

func TestLoggerOptions(t *testing.T) {
	// ARRANGE
	testcases := []struct {
		scenario string
		exec     func(t *testing.T)
	}{
		{scenario: "LogCallsite",
			exec: func(t *testing.T) {
				// ARRANGE
				lg := &logger{}

				// ACT
				_ = LogCallsite(true)(lg)

				// ASSERT
				test.That(t, lg.getCallsite).IsNotNil()
			},
		},
		{scenario: "LoggerBackend",
			exec: func(t *testing.T) {
				// ARRANGE
				be := &mockbackend{}
				lg := &logger{}

				// ACT
				_ = LoggerBackend(be)(lg)

				// ASSERT
				test.That(t, lg.backend).Equals(be)
			},
		},
		{scenario: "LoggerLevel",
			exec: func(t *testing.T) {
				// ARRANGE
				lg := &logger{}

				// ACT
				_ = LoggerLevel(DebugLevel)(lg)

				// ASSERT
				test.That(t, lg.Level).Equals(DebugLevel)
			},
		},

		// LoggerFormat tests
		{scenario: "LoggerFormat/factory error",
			exec: func(t *testing.T) {
				// ARRANGE
				facterr := errors.New("factory error")
				lg := &logger{}

				// ACT
				err := LoggerFormat(func() (Formatter, error) { return nil, facterr })(lg)

				// ASSERT
				test.Error(t, err).Is(facterr)
			},
		},
		{scenario: "LoggerFormat/no backend",
			exec: func(t *testing.T) {
				// ARRANGE
				mf := &mockformatter{}
				lg := &logger{}

				// ACT
				err := LoggerFormat(func() (Formatter, error) { return mf, nil })(lg)

				// ASSERT
				test.IsType[*stdioBackend](t, lg.backend)
				test.That(t, lg.backend.(*stdioBackend).Formatter).Equals(mf)
				test.Error(t, err).IsNil()
			},
		},
		{scenario: "LoggerFormat/backend rejects format",
			exec: func(t *testing.T) {
				// ARRANGE
				fmterr := errors.New("formatter error")
				mf := &mockformatter{}
				lg := &logger{backend: &mockbackend{
					setformatfn: func(f Formatter) error { return fmterr },
				}}

				// ACT
				err := LoggerFormat(func() (Formatter, error) { return mf, nil })(lg)

				// ASSERT
				test.Error(t, err).Is(fmterr)
			},
		},
		{scenario: "LoggerFormat/mux backend",
			exec: func(t *testing.T) {
				// ARRANGE
				lg := &logger{backend: &mux{}}

				// ACT
				err := LoggerFormat(func() (Formatter, error) { return nil, nil })(lg)

				// ASSERT
				test.Error(t, err).Is(ErrInvalidConfiguration)
			},
		},

		// LoggerOutput tests
		{scenario: "LoggerOutput/no backend",
			exec: func(t *testing.T) {
				// ARRANGE
				lg := &logger{}

				// ACT
				err := LoggerOutput(io.Discard)(lg)

				// ASSERT
				test.IsType[*stdioBackend](t, lg.backend)
				test.That(t, lg.backend.(*stdioBackend).Writer).Equals(io.Discard)
				test.Error(t, err).IsNil()
			},
		},
		{scenario: "LoggerOutput/backend rejects output",
			exec: func(t *testing.T) {
				// ARRANGE
				outerr := errors.New("output error")
				lg := &logger{backend: &mockbackend{
					setoutputfn: func(w io.Writer) error { return outerr },
				}}

				// ACT
				err := LoggerOutput(io.Discard)(lg)

				// ASSERT
				test.Error(t, err).Is(outerr)
			},
		},
		{scenario: "LoggerOutput/mux backend",
			exec: func(t *testing.T) {
				// ARRANGE
				lg := &logger{backend: &mux{}}

				// ACT
				err := LoggerOutput(io.Discard)(lg)

				// ASSERT
				test.Error(t, err).Is(ErrInvalidConfiguration)
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.exec(t)
		})
	}
}
