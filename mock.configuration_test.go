package ulog

import (
	"fmt"
	"testing"
)

func TestExpectEntry(t *testing.T) {
	// ARRANGE
	logger, sut := NewMock()

	// ACT
	sut.ExpectEntry()

	// ASSERT
	t.Run("adds entry", func(t *testing.T) {
		wanted := true
		got := len(sut.(*mock).cfg.expectations) == 1
		if wanted != got {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})

	// ARRANGE
	initialExpectation := sut.(*mock).cfg.expected

	t.Run("sets initial expectation", func(t *testing.T) {
		wanted := true
		got := initialExpectation != nil
		if wanted != got {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})

	t.Run("does not change initial expectation when further entries are expected", func(t *testing.T) {
		// ACT
		logger.WithField("key", "value")

		// ASSERT
		wanted := initialExpectation
		got := sut.(*mock).cfg.expected
		if wanted != got {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})

	t.Run("calls configuration functions", func(t *testing.T) {
		// ARRANGE
		cfgWasCalled := false
		cfg := func() func(*mockentry) {
			return func(*mockentry) {
				cfgWasCalled = true
			}
		}

		// ACT
		sut.ExpectEntry(cfg())

		// ASSERT
		wanted := true
		got := cfgWasCalled
		if wanted != got {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})
}

func TestExpectLevel(t *testing.T) {
	// ARRANGE
	sut := &mockentry{}

	// ARRANGE
	testcases := []struct {
		Level
	}{
		{Level: TraceLevel},
		{Level: DebugLevel},
		{Level: InfoLevel},
		{Level: WarnLevel},
		{Level: ErrorLevel},
		{Level: FatalLevel},
	}
	for _, tc := range testcases {
		t.Run(tc.Level.String(), func(t *testing.T) {
			// ACT
			ExpectLevel(tc.Level)(sut)

			// ASSERT
			wanted := tc.Level
			got := *sut.Level
			if wanted != got {
				t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
			}
		})
	}
}

func TestExpectLevelHelpers(t *testing.T) {
	// ARRANGE
	sut := &mock{
		cfg: &expectations{
			expectations: []*expectation{},
		},
	}

	// ARRANGE
	testcases := []struct {
		Level
		fn func(...MockConfiguration)
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
			// ACT
			tc.fn()

			// ASSERT
		})
	}
}

func TestExpectField(t *testing.T) {
	// ARRANGE
	sut := &mockentry{fields: map[string]*string{}}

	// ARRANGE
	testcases := []struct {
		name string
	}{
		{name: "field_name"},
		{name: "also a field name"},
	}
	for idx, tc := range testcases {
		t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
			// ACT
			ExpectField(tc.name)(sut)

			// ASSERT
			wanted := true
			_, got := sut.fields[tc.name]
			if wanted != got {
				t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
			}
		})
	}
}

func TestExpectFieldValue(t *testing.T) {
	// ARRANGE
	sut := &mockentry{fields: map[string]*string{}}

	// ARRANGE
	testcases := []struct {
		name  string
		value string
	}{
		{name: "field_name", value: "field_value"},
		{name: "also a field name", value: "also a value"},
	}
	for idx, tc := range testcases {
		t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
			// ACT
			ExpectFieldValue(tc.name, tc.value)(sut)

			// ASSERT
			t.Run("adds field", func(t *testing.T) {
				wanted := true
				value, got := sut.fields[tc.name]
				if wanted != got {
					t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
				}

				t.Run("with correct value", func(t *testing.T) {
					wanted := tc.value
					got := *value
					if wanted != got {
						t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
					}
				})
			})
		})
	}
}

func TestExpectMessage(t *testing.T) {
	// ARRANGE
	sut := &mockentry{}

	// ARRANGE
	testcases := []struct {
		msg string
	}{
		{msg: "a test message"},
		{msg: "another test message"},
	}
	for idx, tc := range testcases {
		t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
			// ACT
			ExpectMessage(tc.msg)(sut)

			// ASSERT
			wanted := tc.msg
			got := *sut.string
			if wanted != got {
				t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
			}
		})
	}
}
