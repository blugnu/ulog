package ulog

import (
	"context"
	"errors"
	"fmt"
)

// NewMock returns a logger with a listener that sends log entries to a
// mock listener implementation.
//
// Unlike other backend factories which are used to configure a Logger
// initialised by the NewLogger() factory, NewMock() initialises and
// returns a complete Logger, as well as initialising the backend used
// by that Logger to configure and test expectations.
func NewMock() (Logger, MockLog) {
	mock := &mock{}
	logger := &logger{backend: mock}
	ic, _ := logger.init(context.Background(),
		LoggerLevel(TraceLevel),
	)
	return ic, mock
}

// mock is a listener implementation providing methods to configure
// and test expectations about the log entries that will be emitted.
type mock struct {
	expecting  []*MockEntry // the expected entries
	additional []entry      // any entries that were logged in addition to any expected entries
	expected   *MockEntry   // pointer to the next expected entry (nil if no expectations have been set)
	idx        int          // the index (in expecting) of the next expected entry; if expecting is nil then idx is undefined
}

// dispatch satisfies the backend interface, recording a log entry
// and testing whether it matches the expected entry.
func (m *mock) dispatch(e entry) {
	if m.expected == nil {
		m.additional = append(m.additional, e)
		return
	}

	m.expected.ok = m.expected.matches(e)
	m.expected.actual = &e

	m.idx++
	if m.idx == len(m.expecting) {
		m.expected = nil
		return
	}
	m.expected = m.expecting[m.idx]
}

// ExpectationsWereMet returns an error if any of the expected entries were not
// met or if any unexpected entries were logged.
func (m *mock) ExpectationsWereMet() error {
	result := []error{}

	for _, e := range m.expecting {
		switch {
		case e.actual == nil:
			result = append(result, fmt.Errorf("%w:\n  wanted: %v", ErrMissingExpectedLogEntry, e))
		case !e.ok:
			result = append(result, fmt.Errorf("%w:\n  wanted: %v\n  got   : %v", ErrMalformedLogEntry, e, *e.actual))
		default:
			continue
		}
	}

	for _, e := range m.additional {
		result = append(result, fmt.Errorf("%w: %v", ErrUnexpectedLogEntry, e))
	}

	if len(result) > 0 {
		result = append([]error{ErrExpectationsNotMet}, result...)
		return errors.Join(result...)
	}

	return nil
}

// Reset clears all expectations and errors, allowing the mock to be
// reused.
func (m *mock) Reset() {
	*m = mock{}
}
