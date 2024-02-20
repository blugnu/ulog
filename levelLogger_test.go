package ulog

import (
	"testing"

	"github.com/blugnu/test"
)

func TestLevelLogger(t *testing.T) {
	// ARRANGE
	logger, mock := NewMock()

	testcases := []struct {
		scenario string
		exec     func(t *testing.T)
	}{
		{scenario: "AtLevel/info",
			exec: func(t *testing.T) {
				// ARRANGE
				levelled := logger.AtLevel(InfoLevel)
				mock.ExpectInfo(WithMessage("message"))
				mock.ExpectInfo(WithMessage("message: with arg"))

				// ACT
				levelled.Log("message")
				levelled.Logf("message: %s", "with arg")

				// ASSERT
				test.That(t, mock.ExpectationsWereMet()).IsNil()
			},
		},
		{scenario: "AtLevel/debug",
			exec: func(t *testing.T) {
				// ARRANGE
				levelled := logger.AtLevel(DebugLevel)
				mock.ExpectDebug(WithMessage("message"))
				mock.ExpectDebug(WithMessage("message: with arg"))

				// ACT
				levelled.Log("message")
				levelled.Logf("message: %s", "with arg")

				// ASSERT
				test.That(t, mock.ExpectationsWereMet()).IsNil()
			},
		},
		{scenario: "WithField",
			exec: func(t *testing.T) {
				// ARRANGE
				mock.ExpectEntry(
					AtLevel(InfoLevel),
					WithMessage("info message"),
					WithFieldValue("key", "value"),
				)

				sut := logger.AtLevel(InfoLevel)

				// ACT
				sut.WithField("key", "value").
					Log("info message")

				// ASSERT
				test.That(t, mock.ExpectationsWereMet()).IsNil()
			},
		},
		{scenario: "WithFields",
			exec: func(t *testing.T) {
				// ARRANGE
				mock.ExpectEntry(
					AtLevel(InfoLevel),
					WithMessage("info message"),
					WithFieldValues(map[string]string{
						"key1": "value1",
						"key2": "value2",
					}),
				)

				sut := logger.AtLevel(InfoLevel)

				// ACT
				sut.WithFields(map[string]any{
					"key1": "value1",
					"key2": "value2",
				}).Log("info message")

				// ASSERT
				test.That(t, mock.ExpectationsWereMet()).IsNil()
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			// ARRANGE
			mock.Reset()

			// ACT
			tc.exec(t)
		})
	}
}
