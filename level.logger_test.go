package ulog

import "testing"

func TestLevelLogger(t *testing.T) {
	// ARRANGE
	logger, mock := NewMock()

	testcases := []struct {
		name   string
		level  Level
		expect func()
		act    func(LevelLogger)
	}{
		{name: "Log",
			level: InfoLevel,
			expect: func() {
				mock.ExpectInfo(ExpectMessage("message"))
			},
			act: func(sut LevelLogger) { sut.Log("message") }},
		{name: "Logf",
			level: InfoLevel,
			expect: func() {
				mock.ExpectInfo(ExpectMessage("message: arg"))
			},
			act: func(sut LevelLogger) { sut.Logf("message: %s", "arg") }},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// ARRANGE
			mock.Reset()
			tc.expect()

			// ACT
			tc.act(logger.AtLevel(tc.level))

			// ASSERT
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestLevelLogger_WithField(t *testing.T) {
	// ARRANGE
	logger, mock := NewMock()
	mock.ExpectEntry(
		ExpectLevel(InfoLevel),
		ExpectMessage("info message"),
		ExpectFieldValue("key", "value"),
	)

	sut := logger.AtLevel(InfoLevel)

	// ACT
	sut.WithField("key", "value").
		Log("info message")

	// ASSERT
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestLevelLogger_WithFields(t *testing.T) {
	// ARRANGE
	logger, mock := NewMock()
	mock.ExpectEntry(
		ExpectLevel(InfoLevel),
		ExpectMessage("info message"),
		ExpectFieldValues(map[string]string{
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
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}
