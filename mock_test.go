package ulog

import (
	"testing"

	"github.com/blugnu/test"
)

func TestNewMock(t *testing.T) {
	// ACT
	result, _ := NewMock()

	// ASSERT
	test.That(t, result).IsNotNil()
	if ctx, ok := test.IsType[*logcontext](t, result); ok {
		test.IsType[*mock](t, ctx.logger.backend)
		test.That(t, ctx.logger.Level).Equals(TraceLevel)
	}
}

func TestMock(t *testing.T) {
	// ARRANGE
	testcases := []struct {
		scenario string
		exec     func(t *testing.T)
	}{
		// dispatch tests
		{scenario: "dispatch/no expectations",
			exec: func(t *testing.T) {
				// ARRANGE
				sut := &mock{}

				// ACT
				sut.dispatch(entry{logcontext: &logcontext{}, Level: InfoLevel, Message: "entry"})

				// ASSERT
				test.That(t, len(sut.additional)).Equals(1)
			},
		},
		{scenario: "dispatch/one expected/one dispatched/matches",
			exec: func(t *testing.T) {
				// ARRANGE
				elv := InfoLevel
				emsg := "entry"
				e := &MockEntry{level: &elv, string: &emsg}
				sut := &mock{idx: 0, expecting: []*MockEntry{e}, expected: e}

				// ACT
				sut.dispatch(entry{logcontext: &logcontext{}, Level: InfoLevel, Message: "entry"})

				// ASSERT
				test.That(t, len(sut.additional)).Equals(0)
				test.That(t, sut.idx).Equals(1)
				test.That(t, sut.expected).IsNil()
				test.That(t, e.ok).Equals(true)
			},
		},
		{scenario: "dispatch/one expected/one dispatched/does not match",
			exec: func(t *testing.T) {
				// ARRANGE
				elv := InfoLevel
				emsg := "entry"
				e := &MockEntry{level: &elv, string: &emsg}
				sut := &mock{idx: 0, expecting: []*MockEntry{e}, expected: e}

				// ACT
				sut.dispatch(entry{logcontext: &logcontext{}, Level: InfoLevel, Message: "other entry"})

				// ASSERT
				test.That(t, len(sut.additional)).Equals(0)
				test.That(t, sut.idx).Equals(1)
				test.That(t, sut.expected).IsNil()
				test.That(t, e.ok).Equals(false)
			},
		},
		{scenario: "dispatch/one expected/two dispatched/first matches",
			exec: func(t *testing.T) {
				// ARRANGE
				elv := InfoLevel
				emsg := "entry"
				e := &MockEntry{level: &elv, string: &emsg}
				sut := &mock{idx: 0, expecting: []*MockEntry{e}, expected: e}

				// ACT
				sut.dispatch(entry{logcontext: &logcontext{}, Level: InfoLevel, Message: "entry"})
				sut.dispatch(entry{logcontext: &logcontext{}, Level: InfoLevel, Message: "additional"})

				// ASSERT
				test.That(t, len(sut.additional)).Equals(1)
				test.That(t, sut.idx).Equals(1)
				test.That(t, sut.expected).IsNil()
				test.That(t, e.ok).Equals(true)
			},
		},
		{scenario: "dispatch/one expected/two dispatched/first does not match",
			exec: func(t *testing.T) {
				// ARRANGE
				elv := InfoLevel
				emsg := "entry"
				e := &MockEntry{level: &elv, string: &emsg}
				sut := &mock{idx: 0, expecting: []*MockEntry{e}, expected: e}

				// ACT
				sut.dispatch(entry{logcontext: &logcontext{}, Level: InfoLevel, Message: "other entry"})
				sut.dispatch(entry{logcontext: &logcontext{}, Level: InfoLevel, Message: "additional"})

				// ASSERT
				test.That(t, len(sut.additional)).Equals(1)
				test.That(t, sut.idx).Equals(1)
				test.That(t, sut.expected).IsNil()
				test.That(t, e.ok).Equals(false)
			},
		},
		{scenario: "dispatch/two expected/one dispatched/matches",
			exec: func(t *testing.T) {
				// ARRANGE
				elv := InfoLevel
				emsg := "entry"
				e1 := &MockEntry{level: &elv, string: &emsg}
				e2 := &MockEntry{level: &elv, string: &emsg}
				sut := &mock{idx: 0, expecting: []*MockEntry{e1, e2}, expected: e1}

				// ACT
				sut.dispatch(entry{logcontext: &logcontext{}, Level: InfoLevel, Message: "entry"})

				// ASSERT
				test.That(t, len(sut.additional)).Equals(0)
				test.That(t, sut.idx).Equals(1)
				test.That(t, sut.expected).Equals(e2)
				test.That(t, e1.ok).Equals(true)
				test.That(t, e2.ok).Equals(false)
				test.That(t, e2.actual).IsNil()
			},
		},

		// ExpectationsWereMet tests
		{scenario: "ExpectationsWereMet/one expected/all expectations met",
			exec: func(t *testing.T) {
				// ARRANGE
				elv := InfoLevel
				emsg := "entry"
				e := &MockEntry{level: &elv, string: &emsg}
				sut := &mock{idx: 0, expecting: []*MockEntry{e}, expected: e}
				sut.dispatch(entry{logcontext: &logcontext{}, Level: InfoLevel, Message: "entry"})

				// ACT
				result := sut.ExpectationsWereMet()

				// ASSERT
				test.That(t, result).IsNil()
			},
		},
		{scenario: "ExpectationsWereMet/one expected/level expectation not met",
			exec: func(t *testing.T) {
				// ARRANGE
				elv := InfoLevel
				emsg := "entry"
				e := &MockEntry{level: &elv, string: &emsg}
				sut := &mock{idx: 0, expecting: []*MockEntry{e}, expected: e}
				sut.dispatch(entry{logcontext: &logcontext{}, Level: DebugLevel, Message: "entry"})

				// ACT
				result := sut.ExpectationsWereMet()

				// ASSERT
				test.Error(t, result).Is(ErrMalformedLogEntry)
			},
		},
		{scenario: "ExpectationsWereMet/one expected/message expectation not met",
			exec: func(t *testing.T) {
				// ARRANGE
				elv := InfoLevel
				emsg := "entry"
				e := &MockEntry{level: &elv, string: &emsg}
				sut := &mock{idx: 0, expecting: []*MockEntry{e}, expected: e}
				sut.dispatch(entry{logcontext: &logcontext{}, Level: InfoLevel, Message: "some other entry"})

				// ACT
				result := sut.ExpectationsWereMet()

				// ASSERT
				test.Error(t, result).Is(ErrMalformedLogEntry)
			},
		},
		{scenario: "ExpectationsWereMet/one expected/fields expectation not met",
			exec: func(t *testing.T) {
				// ARRANGE
				elv := InfoLevel
				emsg := "entry"
				e := &MockEntry{level: &elv, string: &emsg, fields: map[string]*string{"key": nil}}
				sut := &mock{idx: 0, expecting: []*MockEntry{e}, expected: e}
				sut.dispatch(entry{logcontext: &logcontext{}, Level: InfoLevel, Message: "entry"})

				// ACT
				result := sut.ExpectationsWereMet()

				// ASSERT
				test.Error(t, result).Is(ErrMalformedLogEntry)
			},
		},
		{scenario: "ExpectationsWereMet/additional log entries",
			exec: func(t *testing.T) {
				// ARRANGE
				sut := &mock{}
				sut.dispatch(entry{logcontext: &logcontext{}, Level: InfoLevel, Message: "entry"})

				// ACT
				result := sut.ExpectationsWereMet()

				// ASSERT
				test.Error(t, result).Is(ErrUnexpectedLogEntry)
			},
		},
		{scenario: "ExpectationsWereMet/missing log entries",
			exec: func(t *testing.T) {
				// ARRANGE
				elv := InfoLevel
				emsg := "entry"
				e := &MockEntry{level: &elv, string: &emsg, fields: map[string]*string{"key": nil}}
				sut := &mock{idx: 0, expecting: []*MockEntry{e}, expected: e}

				// ACT
				result := sut.ExpectationsWereMet()

				// ASSERT
				test.Error(t, result).Is(ErrMissingExpectedLogEntry)
			},
		},

		// Reset tests
		{scenario: "Reset",
			exec: func(t *testing.T) {
				// ARRANGE
				sut := &mock{idx: 1, expecting: []*MockEntry{{}}, additional: []entry{{}}, expected: &MockEntry{}}

				// ACT
				sut.Reset()

				// ASSERT
				test.That(t, sut.idx).Equals(0)
				test.That(t, len(sut.expecting)).Equals(0)
				test.That(t, len(sut.additional)).Equals(0)
				test.That(t, sut.expected).IsNil()
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.exec(t)
		})
	}
}
