package ulog

import (
	"context"
	"testing"

	"github.com/blugnu/test"
)

func TestFromContext(t *testing.T) {
	// ARRANGE
	ctx := context.Background()

	// when retrieving a logger from a context (i.e. a logcontext), it is
	// enriched with the context from which it was retrieved, so we need a
	// stand-in, referencing a logger with a no-op enrich function
	logger := &logcontext{
		ctx: ctx,
		logger: &logger{
			enrich: func(og *logcontext, _ context.Context) *logcontext { return og },
		},
	}

	testcases := []struct {
		name     string
		required bool
		option   ContextLoggerOption
		Logger
		result Logger
		panic  *test.Panic
	}{
		{name: "no logger,ContextLoggerRequired==false,ContextLoggerOption==not set",
			required: false,
			result:   noop.logger,
		},
		{name: "no logger,ContextLoggerRequired==false,ContextLoggerOption==Required",
			required: false,
			option:   Required,
			panic:    test.ExpectPanic(ErrNoLoggerInContext),
		},
		{name: "no logger,ContextLoggerRequired==true,ContextLoggerOption==not set",
			required: true,
			panic:    test.ExpectPanic(ErrNoLoggerInContext),
		},
		{name: "no logger,ContextLoggerRequired==true,ContextLoggerOption==Required",
			required: true,
			option:   Required,
			panic:    test.ExpectPanic(ErrNoLoggerInContext),
		},
		{name: "logger present",
			Logger: logger,
			result: logger,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// ARRANGE
			if tc.Logger != nil {
				ctx = context.WithValue(ctx, loggerKey, tc.Logger)
			}
			ContextLoggerRequired = tc.required
			defer tc.panic.IsRecovered(t)

			// ACT
			got := FromContext(ctx, tc.option)

			// ASSERT
			wanted := tc.result
			if wanted != got {
				t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
			}
		})
	}
}

func TestContextWithLogger(t *testing.T) {
	// ARRANGE
	ctx := context.Background()

	// when placing a logger in a context (i.e. a logcontext), it is
	// first enriched with the context to which it is being added, so we
	// need a stand-in, referencing a logger with a no-op enrich function
	logger := &logcontext{
		ctx: ctx,
		logger: &logger{
			enrich: func(og *logcontext, _ context.Context) *logcontext { return og },
		},
	}

	// ACT
	ctx = ContextWithLogger(ctx, logger)

	// ASSERT
	wanted := logger
	got := ctx.Value(loggerKey)
	if wanted != got {
		t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
	}
}
