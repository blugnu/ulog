package ulog

import (
	"testing"

	"github.com/blugnu/test"
)

func TestJsonFormatterOptions(t *testing.T) {
	// ARRANGE
	testcases := []struct {
		scenario string
		exec     func(t *testing.T)
	}{
		// JSONFieldNames tests
		{scenario: "JSONFieldNames/override one field",
			exec: func(t *testing.T) {
				// ARRANGE
				f, _ := NewJSONFormatter()()

				// ACT
				err := JSONFieldNames(map[FieldId]string{TimeField: "tm"})(f.(*jsonfmt))

				// ASSERT
				test.Error(t, err).IsNil()
				test.That(t, f.(*jsonfmt).keys).Equals([numFields]string{
					TimeField:     "tm",
					LevelField:    "level",
					MessageField:  "message",
					FileField:     "file",
					FunctionField: "function",
				})
			},
		},
		{scenario: "JSONFieldNames/override all fields",
			exec: func(t *testing.T) {
				// ARRANGE
				f, _ := NewJSONFormatter()()

				// ACT
				err := JSONFieldNames(map[FieldId]string{
					TimeField:     "tm",
					LevelField:    "lv",
					MessageField:  "msg",
					FileField:     "fi",
					FunctionField: "fn",
				})(f.(*jsonfmt))

				// ASSERT
				test.Error(t, err).IsNil()
				test.That(t, f.(*jsonfmt).keys).Equals([numFields]string{
					TimeField:     "tm",
					LevelField:    "lv",
					MessageField:  "msg",
					FileField:     "fi",
					FunctionField: "fn",
				})
			},
		},

		// JSONLevelLabels tests
		{scenario: "JSONLevelLabels/override one level",
			exec: func(t *testing.T) {
				// ARRANGE
				f, _ := NewJSONFormatter()()

				// ACT
				err := JSONLevelLabels(map[Level]string{TraceLevel: "TRACE"})(f.(*jsonfmt))

				// ASSERT
				test.Error(t, err).IsNil()
				test.That(t, f.(*jsonfmt).levels).Equals([numLevels]string{
					TraceLevel: "TRACE",
					DebugLevel: "debug",
					InfoLevel:  "info",
					WarnLevel:  "warning",
					ErrorLevel: "error",
					FatalLevel: "fatal",
				})
			},
		},
		{scenario: "JSONLevelLabels/override all levels",
			exec: func(t *testing.T) {
				// ARRANGE
				f, _ := NewJSONFormatter()()

				// ACT
				err := JSONLevelLabels(map[Level]string{
					TraceLevel: "TRACE",
					DebugLevel: "DEBUG",
					InfoLevel:  "INFO",
					WarnLevel:  "WARN",
					ErrorLevel: "ERROR",
					FatalLevel: "FATAL",
				})(f.(*jsonfmt))

				// ASSERT
				test.Error(t, err).IsNil()
				test.That(t, f.(*jsonfmt).levels).Equals([numLevels]string{
					TraceLevel: "TRACE",
					DebugLevel: "DEBUG",
					InfoLevel:  "INFO",
					WarnLevel:  "WARN",
					ErrorLevel: "ERROR",
					FatalLevel: "FATAL",
				})
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.exec(t)
		})
	}
}
