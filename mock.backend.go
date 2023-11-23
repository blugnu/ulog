package ulog

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
)

// NewMock returns a logger with a listener that sends log entries to a
// mock listener implementation.
//
// Unlike other backend factories which are used to configure a Logger
// initialised by the NewLogger() factory, NewMock() initialises and
// returns a complete Logger, as well as initialising the backend used
// by that Logger to configure and test expectations.
func NewMock() (Logger, MockLog) {
	mock := &mock{cfg: &expectations{}}
	logger, ic, _ := initLogger(context.Background(), LoggerBackend(mock))
	logger.Level = TraceLevel
	return ic, mock
}

// mock is a listener implementation providing methods to configure
// and test expectations about the log entries that will be emitted.
type mock struct {
	cfg *expectations
}

// Reset clears all expectations and errors, allowing the mock to be
// reused.
func (m *mock) Reset() {
	m.cfg = &expectations{}
}

// dispatch satisfies the backend interface, recording a log entry
// and testing whether it matches the expected entry.
func (m *mock) dispatch(e entry) {
	me := mockentry{&e.Level, &e.Message, nil}
	if e.fields != nil {
		me.fields = map[string]*string{}
		for k, v := range e.fields.m {
			s := fieldValue(v)
			me.fields[k] = &s
		}
	}

	if m.cfg.expected == nil {
		m.cfg.unexpected = append(m.cfg.unexpected, me)
		return
	}

	m.cfg.expected.met = m.cfg.expected.mockentry.matches(&me)
	m.cfg.expected.actual = &me

	m.cfg.next++
	if m.cfg.next == len(m.cfg.expectations) {
		m.cfg.expected = nil
		return
	}
	m.cfg.expected = m.cfg.expectations[m.cfg.next]
}

var boolstr = map[bool]string{true: "true", false: "false"}

// fieldValue returns a string representation of the supplied value.
func fieldValue(v any) string {
	switch v := v.(type) {
	case bool:
		return boolstr[v]
	case int:
		return strconv.FormatInt(int64(v), 10)
	case string:
		return v
	default:
		if stringable, ok := v.(fmt.Stringer); ok {
			return stringable.String()
		}
	}
	b, _ := json.Marshal(v)
	return string(b)
}
