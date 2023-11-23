package ulog

import (
	"errors"
	"testing"
)

func addr(s string) *string {
	return &s
}

func Test_entry_matches(t *testing.T) {
	// ARRANGE
	info := InfoLevel
	warn := WarnLevel
	msg := "message"
	othermsg := "other message"
	flds := map[string]*string{"key": addr("value")}
	otherfldsk := map[string]*string{"other key": addr("value")}
	otherfldsv := map[string]*string{"key": addr("other value")}

	testcases := []struct {
		name   string
		sut    *mockentry
		target *mockentry
		result bool
	}{
		{name: "zero-value vs level (info)", sut: &mockentry{}, target: &mockentry{Level: &info}, result: true},
		{name: "zero-value vs level (warn)", sut: &mockentry{}, target: &mockentry{Level: &warn}, result: true},
		{name: "zero-value vs message", sut: &mockentry{}, target: &mockentry{string: &msg}, result: true},
		{name: "zero-value vs fields", sut: &mockentry{}, target: &mockentry{fields: flds}, result: true},
		{name: "zero-value vs level and message", sut: &mockentry{}, target: &mockentry{Level: &info, string: &msg}, result: true},
		{name: "zero-value vs level, message and fields", sut: &mockentry{}, target: &mockentry{Level: &info, string: &msg, fields: flds}, result: true},
		{name: "specified level vs same level", sut: &mockentry{Level: &info}, target: &mockentry{Level: &info}, result: true},
		{name: "specified level vs different level", sut: &mockentry{Level: &info}, target: &mockentry{Level: &warn}, result: false},
		{name: "specified message vs same message", sut: &mockentry{string: &msg}, target: &mockentry{string: &msg}, result: true},
		{name: "specified message vs different message", sut: &mockentry{string: &msg}, target: &mockentry{string: &othermsg}, result: false},
		{name: "specified fields vs same keys, different values", sut: &mockentry{fields: flds}, target: &mockentry{fields: otherfldsv}, result: false},
		{name: "specified fields vs different keys", sut: &mockentry{fields: flds}, target: &mockentry{fields: otherfldsk}, result: false},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// ACT
			got := tc.sut.matches(tc.target)

			// ASSERT
			wanted := tc.result
			if wanted != got {
				t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
			}
		})
	}
}

func Test_entry_String(t *testing.T) {
	// ARRANGE
	info := InfoLevel
	msg := "message"
	flds := map[string]*string{"key": addr("value")}

	testcases := []struct {
		name   string
		sut    *mockentry
		result string
	}{
		{name: "zero-value", sut: &mockentry{}, result: "level=<any> message=<any> fields=<any>"},
		{name: "specified level", sut: &mockentry{Level: &info}, result: "level=Info message=<any> fields=<any>"},
		{name: "specified message", sut: &mockentry{string: &msg}, result: "level=<any> message=\"message\" fields=<any>"},
		{name: "specified fields", sut: &mockentry{fields: flds}, result: "level=<any> message=<any> fields=[\"key\"=\"value\"]"},
		{name: "specified level and message", sut: &mockentry{Level: &info, string: &msg}, result: "level=Info message=\"message\" fields=<any>"},
		{name: "specified level, message and fields", sut: &mockentry{Level: &info, string: &msg, fields: flds}, result: "level=Info message=\"message\" fields=[\"key\"=\"value\"]"},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// ACT
			got := tc.sut.String()

			// ASSERT
			wanted := tc.result
			if wanted != got {
				t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
			}
		})
	}
}

func TestExpectationsWereMet(t *testing.T) {
	// ARRANGE
	testcases := []struct {
		name    string
		setup   func(MockLog)
		logging func(Logger)
		result  error
	}{
		{name: "nothing expected, nothing logged", setup: func(m MockLog) {}, logging: func(logger Logger) {}, result: nil},
		{name: "nothing expected, info logged", setup: func(m MockLog) {}, logging: func(logger Logger) { logger.Info("info message") }, result: ErrExpectationsNotMet},
		{name: "one expected, nothing logged", setup: func(m MockLog) { m.ExpectEntry() }, logging: func(logger Logger) {}, result: ErrExpectationsNotMet},
		{name: "one expected, one logged", setup: func(m MockLog) { m.ExpectEntry() }, logging: func(logger Logger) { logger.Info("info message") }, result: nil},
		{name: "info expected, debug logged", setup: func(m MockLog) { m.ExpectEntry(ExpectLevel(InfoLevel)) }, logging: func(logger Logger) { logger.Debug("debug message") }, result: ErrExpectationsNotMet},
		{name: "field expected with any value, no fields", setup: func(m MockLog) { m.ExpectEntry(ExpectField("field")) }, logging: func(logger Logger) { logger.Debug("debug message") }, result: ErrExpectationsNotMet},
		{name: "field expected with any value, field with different key", setup: func(m MockLog) { m.ExpectEntry(ExpectField("field")) }, logging: func(logger Logger) { logger.WithField("other", nil).Debug("debug message") }, result: ErrExpectationsNotMet},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// ARRANGE
			logger, sut := NewMock()
			tc.setup(sut)

			// ACT
			tc.logging(logger)

			// ASSERT
			wanted := tc.result
			got := sut.ExpectationsWereMet()
			if !errors.Is(got, wanted) {
				t.Errorf("\nwanted %v\ngot    %v", wanted, got)
			}
		})
	}
}
