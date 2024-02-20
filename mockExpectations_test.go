package ulog

import (
	"testing"

	"github.com/blugnu/test"
)

func TestExpectEntry(t *testing.T) {
	// ARRANGE
	sut := &mock{}
	refCalled := false
	ref := func(*MockEntry) { refCalled = true }

	// ACT
	sut.ExpectEntry(ref)

	// ASSERT
	test.That(t, len(sut.expecting)).Equals(1)
	test.That(t, sut.expected).Equals(&MockEntry{fields: map[string]*string{}})
	test.IsTrue(t, refCalled)

	t.Run("level helpers", func(t *testing.T) {
		// ARRANGE
		testcases := []struct {
			Level
			fn func(...EntryExpectation)
		}{
			{Level: TraceLevel, fn: sut.ExpectTrace},
			{Level: DebugLevel, fn: sut.ExpectDebug},
			{Level: InfoLevel, fn: sut.ExpectInfo},
			{Level: WarnLevel, fn: sut.ExpectWarn},
			{Level: ErrorLevel, fn: sut.ExpectError},
			{Level: FatalLevel, fn: sut.ExpectFatal},
		}
		for _, tc := range testcases {
			t.Run(tc.Level.String(), func(t *testing.T) {
				// ARRANGE
				// overwrite the existing sut with a new one to reset the state;
				// (we can't just new up an entirely new sut because the test
				// cases above reference the functions on the original sut)
				*sut = mock{}

				// ACT
				tc.fn()

				// ASSERT
				test.That(t, len(sut.expecting)).Equals(1)
				test.That(t, sut.expected).Equals(&MockEntry{level: &tc.Level, fields: map[string]*string{}})
			})
		}
	})
}

func TestExpectedEntryRefinements(t *testing.T) {
	// ARRANGE
	var sut *MockEntry

	testcases := []struct {
		scenario string
		exec     func(t *testing.T)
	}{
		{scenario: "AtLevel",
			exec: func(t *testing.T) {
				for _, lvl := range Levels {
					t.Run(lvl.String(), func(t *testing.T) {
						// ACT
						AtLevel(lvl)(sut)

						// ASSERT
						test.That(t, sut.level).Equals(&lvl)
					})
				}
			},
		},
		{scenario: "WithField",
			exec: func(t *testing.T) {
				// ACT
				WithField("key")(sut)

				// ASSERT
				test.Map(t, sut.fields).Equals(map[string]*string{"key": nil})
			},
		},
		{scenario: "WithFields",
			exec: func(t *testing.T) {
				// ACT
				WithFields("key1", "key2")(sut)

				// ASSERT
				test.Map(t, sut.fields).Equals(map[string]*string{"key1": nil, "key2": nil})
			},
		},
		{scenario: "WithFieldValue",
			exec: func(t *testing.T) {
				// ACT
				WithFieldValue("key", "value")(sut)

				// ASSERT
				test.Map(t, sut.fields).Equals(map[string]*string{"key": addrOf("value")})
			},
		},
		{scenario: "WithFieldValues",
			exec: func(t *testing.T) {
				// ACT
				WithFieldValues(map[string]string{"key1": "value1", "key2": "value2"})(sut)

				// ASSERT
				test.Map(t, sut.fields).Equals(map[string]*string{"key1": addrOf("value1"), "key2": addrOf("value2")})
			},
		},
		{scenario: "WithMessage",
			exec: func(t *testing.T) {
				// ACT
				WithMessage("message")(sut)

				// ASSERT
				test.That(t, sut.string).Equals(addrOf("message"))
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			// ARRANGE
			sut = &MockEntry{fields: map[string]*string{}}

			// ACT
			tc.exec(t)
		})
	}
}
