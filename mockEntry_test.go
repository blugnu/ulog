package ulog

import (
	"testing"

	"github.com/blugnu/test"
)

func addrOf[T any](s T) *T {
	return &s
}

func TestMockEntry(t *testing.T) {
	// ARRANGE
	testcases := []struct {
		scenario string
		exec     func(t *testing.T)
	}{
		{scenario: "matches}",
			exec: func(t *testing.T) {
				// ARRANGE
				target := entry{
					Level:   InfoLevel,
					Message: "entry",
					logcontext: &logcontext{fields: &fields{m: map[string]any{
						"key":            "value",
						"additional key": "additional value",
					}}},
				}

				testcases := []struct {
					scenario string
					exec     func(t *testing.T)
				}{
					{scenario: "any level/any message/any fields",
						exec: func(t *testing.T) {
							// ARRANGE
							e := &MockEntry{}

							// ACT
							result := e.matches(target)

							// ASSERT
							test.IsTrue(t, result)
						},
					},
					{scenario: "level set/matches",
						exec: func(t *testing.T) {
							// ARRANGE
							e := &MockEntry{level: addrOf(InfoLevel)}

							// ACT
							result := e.matches(target)

							// ASSERT
							test.IsTrue(t, result)
						},
					},
					{scenario: "level set/different",
						exec: func(t *testing.T) {
							// ARRANGE
							e := &MockEntry{level: addrOf(DebugLevel)}

							// ACT
							result := e.matches(target)

							// ASSERT
							test.IsFalse(t, result)
						},
					},
					{scenario: "message set/matches",
						exec: func(t *testing.T) {
							// ARRANGE
							e := &MockEntry{string: addrOf("entry")}

							// ACT
							result := e.matches(target)

							// ASSERT
							test.IsTrue(t, result)
						},
					},
					{scenario: "message set/different",
						exec: func(t *testing.T) {
							// ARRANGE
							e := &MockEntry{string: addrOf("different")}

							// ACT
							result := e.matches(target)

							// ASSERT
							test.IsFalse(t, result)
						},
					},
					{scenario: "fields/key set/present",
						exec: func(t *testing.T) {
							// ARRANGE
							e := &MockEntry{fields: map[string]*string{"key": nil}}

							// ACT
							result := e.matches(target)

							// ASSERT
							test.IsTrue(t, result)
						},
					},
					{scenario: "fields/key set/not present",
						exec: func(t *testing.T) {
							// ARRANGE
							e := &MockEntry{fields: map[string]*string{"other key": nil}}

							// ACT
							result := e.matches(target)

							// ASSERT
							test.IsFalse(t, result)
						},
					},
					{scenario: "fields/key and value set/present",
						exec: func(t *testing.T) {
							// ARRANGE
							e := &MockEntry{fields: map[string]*string{"key": addrOf("value")}}

							// ACT
							result := e.matches(target)

							// ASSERT
							test.IsTrue(t, result)
						},
					},
					{scenario: "fields/key and value set/different value",
						exec: func(t *testing.T) {
							// ARRANGE
							e := &MockEntry{fields: map[string]*string{"key": addrOf("other value")}}

							// ACT
							result := e.matches(target)

							// ASSERT
							test.IsFalse(t, result)
						},
					},
					{scenario: "fields/more keys expected than presents",
						exec: func(t *testing.T) {
							// ARRANGE
							e := &MockEntry{fields: map[string]*string{
								"key":   addrOf("value"),
								"key+":  nil,
								"key++": nil,
							}}

							// ACT
							result := e.matches(target)

							// ASSERT
							test.IsFalse(t, result)
						},
					},
				}
				for _, tc := range testcases {
					t.Run(tc.scenario, func(t *testing.T) {
						tc.exec(t)
					})
				}
			},
		},

		// String tests
		{scenario: "String/zero-value",
			exec: func(t *testing.T) {
				// ARRANGE
				e := &MockEntry{}

				// ACT
				result := e.String()

				// ASSERT
				test.That(t, result).Equals("level=<any> message=<any> fields=<any>")
			},
		},
		{scenario: "String/level set",
			exec: func(t *testing.T) {
				// ARRANGE
				e := &MockEntry{level: addrOf(InfoLevel)}

				// ACT
				result := e.String()

				// ASSERT
				test.That(t, result).Equals("level=info message=<any> fields=<any>")
			},
		},
		{scenario: "String/message set",
			exec: func(t *testing.T) {
				// ARRANGE
				e := &MockEntry{string: addrOf("entry")}

				// ACT
				result := e.String()

				// ASSERT
				test.That(t, result).Equals("level=<any> message=\"entry\" fields=<any>")
			},
		},
		{scenario: "String/fields set",
			exec: func(t *testing.T) {
				// ARRANGE
				e := &MockEntry{fields: map[string]*string{"key": addrOf("value")}}

				// ACT
				result := e.String()

				// ASSERT
				test.That(t, result).Equals("level=<any> message=<any> fields=[\"key\"=\"value\"]")
			},
		},
		{scenario: "String/field key set",
			exec: func(t *testing.T) {
				// ARRANGE
				e := &MockEntry{fields: map[string]*string{"key": nil}}

				// ACT
				result := e.String()

				// ASSERT
				test.That(t, result).Equals("level=<any> message=<any> fields=[\"key\"=<any>]")
			},
		},
		{scenario: "String/level, message and fields set",
			exec: func(t *testing.T) {
				// ARRANGE
				e := &MockEntry{level: addrOf(InfoLevel), string: addrOf("entry"), fields: map[string]*string{"key": addrOf("value")}}

				// ACT
				result := e.String()

				// ASSERT
				test.That(t, result).Equals("level=info message=\"entry\" fields=[\"key\"=\"value\"]")
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.exec(t)
		})
	}
}
