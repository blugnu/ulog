package ulog

import (
	"errors"
	"fmt"
)

// expectation represents an expected log entry.
type expectation struct {
	*mockentry
	actual *mockentry
	met    bool
}

// expectations represents a set of expected log entries.
type expectations struct {
	expectations []*expectation
	unexpected   []mockentry
	expected     *expectation
	next         int
}

// ExpectationsWereMet returns an error if any of the expected entries were not
// met or if any unexpected entries were logged.
func (m *mock) ExpectationsWereMet() error {
	result := []error{}

	for _, e := range m.cfg.expectations {
		if !e.met {
			result = append(result, fmt.Errorf("expected entry not met:\n  wanted: %v\n  got   : %v", e.mockentry, e.actual))
		}
	}

	for _, entry := range m.cfg.unexpected {
		result = append(result, fmt.Errorf("unexpected entry: %v", entry))
	}

	if len(result) > 0 {
		result = append([]error{ErrExpectationsNotMet}, result...)
		return errors.Join(result...)
	}

	return nil
}
