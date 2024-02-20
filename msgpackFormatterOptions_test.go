package ulog

import (
	"testing"

	"github.com/blugnu/test"
)

func TestMsgpackFormatterOptions(t *testing.T) {
	// ARRANGE
	sut := &msgpackfmt{}

	testcases := []struct {
		scenario string
		exec     func(t *testing.T)
	}{
		{scenario: "MsgpackFieldNames/override one field",
			exec: func(t *testing.T) {
				// ARRANGE
				keys := map[FieldId]string{TimeField: "THE TIME"}

				// ACT
				err := MsgpackKeys(keys)(sut)

				// ASSERT
				test.That(t, err).IsNil()
				test.That(t, sut.keys).Equals([numFields][]byte{
					TimeField:     []byte("\xa8THE TIME"),
					LevelField:    []byte("\xa5level"),
					MessageField:  []byte("\xa7message"),
					FileField:     []byte("\xa4file"),
					FunctionField: []byte("\xa8function"),
				})
			},
		},
		{scenario: "MsgpackFieldNames/override all fields",
			exec: func(t *testing.T) {
				// ARRANGE
				keys := map[FieldId]string{
					TimeField:     "TIME",
					LevelField:    "LEVEL",
					MessageField:  "MESSAGE",
					FileField:     "FILE",
					FunctionField: "FUNCTION",
				}

				// ACT
				err := MsgpackKeys(keys)(sut)

				// ASSERT
				test.That(t, err).IsNil()
				test.That(t, sut.keys).Equals([numFields][]byte{
					TimeField:     []byte("\xa4TIME"),
					LevelField:    []byte("\xa5LEVEL"),
					MessageField:  []byte("\xa7MESSAGE"),
					FileField:     []byte("\xa4FILE"),
					FunctionField: []byte("\xa8FUNCTION"),
				})
			},
		},
		{scenario: "MsgpackLevels/override one level",
			exec: func(t *testing.T) {
				// ARRANGE
				levels := map[Level]string{InfoLevel: "INFO"}

				// ACT
				err := MsgpackLevels(levels)(sut)

				// ASSERT
				test.That(t, err).IsNil()
				test.That(t, sut.levels).Equals([numLevels][]byte{
					TraceLevel: []byte("\xa5trace"),
					DebugLevel: []byte("\xa5debug"),
					InfoLevel:  []byte("\xa4INFO"),
					WarnLevel:  []byte("\xa7warning"),
					ErrorLevel: []byte("\xa5error"),
					FatalLevel: []byte("\xa5fatal"),
				})
			},
		},
		{scenario: "MsgpackLevels/override all levels",
			exec: func(t *testing.T) {
				// ARRANGE
				levels := map[Level]string{
					TraceLevel: "TRACE",
					DebugLevel: "DEBUG",
					InfoLevel:  "INFO",
					WarnLevel:  "WARNING",
					ErrorLevel: "ERROR",
					FatalLevel: "FATAL",
				}

				// ACT
				err := MsgpackLevels(levels)(sut)

				// ASSERT
				test.That(t, err).IsNil()
				test.That(t, sut.levels).Equals([numLevels][]byte{
					TraceLevel: []byte("\xa5TRACE"),
					DebugLevel: []byte("\xa5DEBUG"),
					InfoLevel:  []byte("\xa4INFO"),
					WarnLevel:  []byte("\xa7WARNING"),
					ErrorLevel: []byte("\xa5ERROR"),
					FatalLevel: []byte("\xa5FATAL"),
				})
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			// ARRANGE
			sut.init()

			// ACT
			tc.exec(t)
		})
	}
}
