package ulog

import (
	"context"
	"testing"

	"github.com/blugnu/test"
)

func TestFromContext(t *testing.T) {
	// ARRANGE
	ctx := context.Background()

	// when retrieving or putting a logger from/into a context, it is
	// enriched with the context so we need a stand-in logger with a
	// no-op enrich function
	lg := &logcontext{
		ctx: ctx,
		logger: &logger{
			enrich: func(og *logcontext, _ context.Context) *logcontext { return og },
		},
	}

	testcases := []struct {
		scenario string
		exec     func(t *testing.T)
	}{
		{scenario: "FromContext()/no logger in context",
			exec: func(t *testing.T) {
				// ARRANGE
				// ACT
				result := FromContext(ctx)

				// ASSERT
				test.That(t, result).Equals(noop.logger)
			},
		},
		{scenario: "FromContext(Required)/no logger in context",
			exec: func(t *testing.T) {
				// ARRANGE
				defer test.ExpectPanic(ErrNoLoggerInContext).Assert(t)

				// ACT
				_ = FromContext(ctx, Required)
			},
		},
		{scenario: "FromContext()/logger in context",
			exec: func(t *testing.T) {
				// ARRANGE
				ctx = context.WithValue(ctx, loggerKey, lg)

				// ACT
				result := FromContext(ctx)

				// ASSERT
				test.That(t, result).Equals(lg)
			},
		},
		{scenario: "ContextWithLogger()",
			exec: func(t *testing.T) {
				// ACT
				result := ContextWithLogger(ctx, lg)

				// ASSERT
				test.That(t, result.Value(loggerKey)).Equals(lg)
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.exec(t)
		})
	}
}
