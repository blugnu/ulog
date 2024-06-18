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
		{scenario: "FromContext()/no logger",
			exec: func(t *testing.T) {
				// ACT
				result := FromContext(ctx)

				// ASSERT
				test.That(t, result).Equals(noop.logger)
			},
		},
		{scenario: "FromContext()/no logger/NoopIfNotPresent",
			exec: func(t *testing.T) {
				// ACT
				result := FromContext(ctx, NoopIfNotPresent)

				// ASSERT
				test.That(t, result).Equals(noop.logger)
			},
		},
		{scenario: "FromContext/no logger/PanicIfNotPresent",
			exec: func(t *testing.T) {
				// ARRANGE ASSERT
				defer test.ExpectPanic(ErrNoLoggerInContext).Assert(t)

				// ACT
				_ = FromContext(ctx, PanicIfNotPresent)
			},
		},
		{scenario: "FromContext/no logger/NilIfNotPresent",
			exec: func(t *testing.T) {
				// ACT
				result := FromContext(ctx, NilIfNotPresent)

				// ASSERT
				test.That(t, result).IsNil()
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
