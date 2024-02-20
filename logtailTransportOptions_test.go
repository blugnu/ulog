package ulog

import (
	"os"
	"testing"
	"time"

	"github.com/blugnu/test"
)

func TestLogtailTransportOptions(t *testing.T) {
	// ARRANGE
	var (
		bh *logtailBatchHandler
		lt = &logtail{batch: &Batch{}}
	)
	testcases := []struct {
		scenario string
		exec     func(t *testing.T)
	}{
		{scenario: "LogtailEndpoint",
			exec: func(t *testing.T) {
				// ACT
				_ = LogtailEndpoint("https://custom.endpoint.com")(lt)

				// ASSERT
				test.That(t, bh.endpoint).Equals("https://custom.endpoint.com")
			},
		},
		{scenario: "LogtailMaxBatch",
			exec: func(t *testing.T) {
				// ACT
				_ = LogtailMaxBatch(42)(lt)

				// ASSERT
				test.That(t, lt.batch.max).Equals(42)
			},
		},
		{scenario: "LogtailMaxLatency",
			exec: func(t *testing.T) {
				// ACT
				_ = LogtailMaxLatency(1 * time.Hour)(lt)

				// ASSERT
				test.That(t, lt.maxLatency).Equals(time.Hour)
			},
		},
		{scenario: "LogtailSourceToken/literal",
			exec: func(t *testing.T) {
				// ACT
				_ = LogtailSourceToken("literal_token")(lt)

				// ASSERT
				test.That(t, bh.token).Equals("literal_token")
			},
		},
		{scenario: "LogtailSourceToken/from environment",
			exec: func(t *testing.T) {
				// ARRANGE
				os.Setenv("TEST_TOKEN", "envtoken")
				defer os.Unsetenv("TEST_TOKEN")

				// ACT
				_ = LogtailSourceToken("TEST_TOKEN")(lt)

				// ASSERT
				test.That(t, bh.token).Equals("envtoken")
			},
		},
		{scenario: "LogtailSourceToken/from file",
			exec: func(t *testing.T) {
				// ACT
				_ = LogtailSourceToken("testdata/token.txt")(lt)

				// ASSERT
				test.That(t, bh.token).Equals("file_token")
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			// ARRANGE
			bh = newLogtailBatchHandler()
			lt.batch.init(bh, 16)

			// ACT
			tc.exec(t)
		})
	}
}
