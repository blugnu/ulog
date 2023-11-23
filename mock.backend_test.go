package ulog

import (
	"reflect"
	"testing"
	"time"

	"github.com/blugnu/test"
)

func TestNewMock(t *testing.T) {
	// ACT
	result, _ := NewMock()

	// ASSERT
	t.Run("configures expected backend", func(t *testing.T) {
		wanted := true
		_, got := result.(*logcontext).logger.backend.(*mock)
		if wanted != got {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})
}

func TestMock_Reset(t *testing.T) {
	// ARRANGE
	sut := &mock{cfg: &expectations{next: 42}}

	// ACT
	sut.Reset()

	// ASSERT
	wanted := &expectations{}
	got := sut.cfg
	if !reflect.DeepEqual(wanted, got) {
		t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
	}
}

func TestMock(t *testing.T) {
	// ARRANGE
	logger, mock := NewMock()

	// ARRANGE

	testcases := []struct {
		name  string
		setup func(MockLog)
		log   func(Logger)
		error
	}{
		{name: "no entries expected, none logged",
			setup: func(mock MockLog) {
				// NO-OP
			},
			log: func(logger Logger) {
				// NO-OP
			},
			error: nil,
		},
		{name: "1 entry expected, none logged",
			setup: func(mock MockLog) {
				mock.ExpectEntry()
			},
			log: func(logger Logger) {
				// NO-OP
			},
			error: ErrExpectationsNotMet,
		},
		{name: "no entries expected, 1 logged",
			setup: func(mock MockLog) {
				// NO-OP
			},
			log: func(logger Logger) {
				logger.Info("info message")
			},
			error: ErrExpectationsNotMet,
		},
		{name: "2 entries expected, 1 logged",
			setup: func(mock MockLog) {
				mock.ExpectEntry()
				mock.ExpectEntry()
			},
			log: func(logger Logger) {
				logger.Info("info message")
			},
			error: ErrExpectationsNotMet,
		},
		{name: "1 entry expected, 1 logged",
			setup: func(mock MockLog) {
				mock.ExpectEntry()
			},
			log: func(logger Logger) {
				logger.Info("info message")
			},
			error: nil,
		},
		{name: "1 entry expected, 2 logged",
			setup: func(mock MockLog) {
				mock.ExpectEntry()
			},
			log: func(logger Logger) {
				logger.Info("info message")
				logger.Info("info message")
			},
			error: ErrExpectationsNotMet,
		},
		{name: "1 entry expected with no fields, 1 logged with fields",
			setup: func(mock MockLog) {
				mock.ExpectEntry()
			},
			log: func(logger Logger) {
				logger.WithField("key", "value").Info("info message")
			},
			error: nil,
		},
		{name: "1 entry expected with 1 field, 1 logged with 2 fields",
			setup: func(mock MockLog) {
				mock.ExpectEntry(
					ExpectField("key"),
				)
			},
			log: func(logger Logger) {
				logger.WithFields(map[string]any{
					"key":   "value",
					"extra": "secret",
				}).Info("info message")
			},
		},
		{name: "1 entry expected with 2 fields, 1 logged with 1 field",
			setup: func(mock MockLog) {
				mock.ExpectEntry(
					ExpectField("key"),
					ExpectField("extra"),
				)
			},
			log: func(logger Logger) {
				logger.WithField("key", "value").
					Info("info message")
			},
			error: ErrExpectationsNotMet,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// ARRANGE
			tc.setup(mock)
			defer mock.Reset()

			// ACT
			tc.log(logger)

			// ASSERT
			test.ErrorIs(t, tc.error, mock.ExpectationsWereMet())
		})
	}
}

func Test_fieldValue(t *testing.T) {
	// ARRANGE
	tm := time.Date(2007, 6, 5, 4, 3, 2, 1, time.UTC)

	type json struct {
		S string
	}

	testcases := []struct {
		name   string
		value  any
		result string
	}{
		{name: "true", value: true, result: "true"},
		{name: "false", value: false, result: "false"},
		{name: "string", value: "string", result: "string"},
		{name: "stringer", value: tm, result: "2007-06-05 04:03:02.000000001 +0000 UTC"},
		{name: "int", value: 42, result: "42"},
		{name: "json", value: json{S: "string"}, result: "{\"S\":\"string\"}"},
		{name: "nil", value: nil, result: "null"},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// ACT
			got := fieldValue(tc.value)

			// ASSERT
			wanted := tc.result
			if wanted != got {
				t.Errorf("\nwanted %v\ngot    %v", wanted, got)
			}
		})
	}
}
