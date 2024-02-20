package ulog

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/blugnu/test"
)

func TestNewLogger(t *testing.T) {
	// ARRANGE
	testcases := []struct {
		scenario string
		exec     func(t *testing.T)
	}{
		{scenario: "option error",
			exec: func(t *testing.T) {
				// ARRANGE
				opterr := errors.New("option error")
				opt := func(*logger) error { return opterr }

				// ACT
				lg, cfn, err := NewLogger(context.Background(), opt)

				// ASSERT
				test.That(t, lg).IsNil()
				test.That(t, cfn).IsNotNil()
				test.Error(t, err).Is(opterr)

				t.Run("close func does not panic", func(t *testing.T) {
					defer test.ExpectPanic(nil).Assert(t)
					cfn()
				})
			},
		},
		{scenario: "no backend option",
			exec: func(t *testing.T) {
				// ACT
				lg, cfn, err := NewLogger(context.Background())

				// ASSERT
				test.That(t, lg).IsNotNil()
				test.That(t, cfn).IsNotNil()
				test.That(t, err).IsNil()

				if backend, ok := test.IsType[*stdioBackend](t, lg.(*logcontext).logger.backend); ok {
					if formatter, ok := test.IsType[*logfmt](t, backend.Formatter); ok {
						lf, _ := LogfmtFormatter()()
						wanted := lf.(*logfmt)
						test.That(t, formatter).Equals(wanted)
					}
					test.That(t, backend.Writer).Equals(any(os.Stdout).(io.Writer))
				}

				t.Run("close func does not panic", func(t *testing.T) {
					defer test.ExpectPanic(nil).Assert(t)
					cfn()
				})
			},
		},
		{scenario: "backend that fails to start",
			exec: func(t *testing.T) {
				// ARRANGE
				beerr := errors.New("backend error")
				be := &mockbackend{
					startfn: func() (func(), error) { return nil, beerr },
				}

				// ACT
				lg, cfn, err := NewLogger(context.Background(), LoggerBackend(be))

				// ASSERT
				test.IsNil(t, lg, "logger")
				test.IsNotNil(t, cfn, "close function")
				test.Error(t, err).Is(beerr)

				t.Run("close func does not panic", func(t *testing.T) {
					defer test.ExpectPanic(nil).Assert(t)
					cfn()
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

func TestLogger(t *testing.T) {
	// ARRANGE
	ctx := context.Background()
	lg := &logger{}

	// register a no-op enrichment func to ensure coverage by exercising
	// the re-assignment of the enrichment function ref
	//
	// TODO: replace enrichment function with an enrichment interface so that a more meaningful test can be written
	og := enrichment
	defer func() { enrichment = og }()
	enrichment = []EnrichmentFunc{func(context.Context) map[string]any { return nil }}

	testcases := []struct {
		scenario string
		exec     func(t *testing.T)
	}{
		// exit
		{scenario: "exit",
			exec: func(t *testing.T) {
				// ARRANGE
				closeCalled := false
				exitCalled := false
				exitCode := 0

				ogexit := ExitFn
				defer func() { ExitFn = ogexit }()
				ExitFn = func(i int) { exitCalled = true; exitCode = i }

				ogclose := lg.closeFn
				defer func() { lg.closeFn = ogclose }()
				lg.closeFn = func() { closeCalled = true }

				// ACT
				lg.exit(42)

				// ASSERT
				test.IsTrue(t, closeCalled, "close func is called")
				test.IsTrue(t, exitCalled, "exit func is called")
				test.That(t, exitCode, "exit code").Equals(42)
			},
		},

		// init tests
		{scenario: "init/valid",
			exec: func(t *testing.T) {
				// ARRANGE
				optApplied := false
				opt := func(*logger) error { optApplied = true; return nil }

				// ACT
				ic, err := lg.init(ctx, opt)

				// ASSERT
				test.That(t, ic).IsNotNil()
				test.That(t, err).IsNil()

				test.IsTrue(t, optApplied, "options applied")

				test.That(t, lg.Level, "default level").Equals(InfoLevel)
				test.That(t, ic, "initial context").Equals(&logcontext{ctx: ctx, logger: lg, exitCode: 1})
			},
		},
		{scenario: "init/option error",
			exec: func(t *testing.T) {
				// ARRANGE
				opterr := errors.New("option error")
				opt := func(*logger) error { return opterr }

				// ACT
				ic, err := lg.init(ctx, opt)

				// ASSERT
				test.That(t, ic).IsNil()
				test.Error(t, err).Is(opterr)
			},
		},

		// isLevelEnabled
		{scenario: "isLevelEnabled",
			exec: func(t *testing.T) {
				// ARRANGE
				og := lg.Level
				defer func() { lg.Level = og }()

				for _, lg.Level = range Levels {
					for _, el := range Levels {
						t.Run(fmt.Sprintf("logger@%s/entry@%s", lg.Level, el), func(t *testing.T) {
							// ACT
							got := lg.isLevelEnabled(ctx, el)

							// ASSERT
							test.That(t, got).Equals(el <= lg.Level)
						})
					}
				}
			},
		},

		// log tests
		{scenario: "log",
			exec: func(t *testing.T) {
				// ARRANGE
				isDispatched := false

				og := lg.backend
				defer func() { lg.backend = og }()
				lg.backend = &mockbackend{dispatchfn: func(e entry) { isDispatched = true }}

				testcases := []struct {
					scenario string
					exec     func(t *testing.T)
				}{
					{scenario: "noop entry",
						exec: func(t *testing.T) {
							// ACT
							lg.log(noop.entry)

							// ASSERT
							test.IsFalse(t, isDispatched, "entry dispatched")
						},
					},
					{scenario: "valid entry",
						exec: func(t *testing.T) {
							// ACT
							lg.log(entry{
								logcontext: &logcontext{logger: lg},
							})

							// ASSERT
							test.IsTrue(t, isDispatched, "entry dispatched")
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

		// noEnrichment
		{scenario: "noEnrichment",
			exec: func(t *testing.T) {
				// ARRANGE
				type key int
				ctx := context.WithValue(ctx, key(1), "value")
				logctx := &logcontext{
					logger:     lg,
					dispatcher: &mockdispatcher{},
				}

				// ACT
				result := lg.noEnrichment(logctx, ctx)

				// ASSERT
				test.Value(t, result).DoesNotEqual(logctx, "new log context")
				test.That(t, result.ctx, "ctx").Equals(ctx)
				test.That(t, result.dispatcher, "dispatcher").Equals(logctx.dispatcher)
				test.That(t, result.fields, "fields").IsNil()
				test.That(t, result.xfields, "xfields").IsNil()
			},
		},

		// withEnrichment
		{scenario: "withEnrichment",
			exec: func(t *testing.T) {
				// ARRANGE
				type key int
				ctx := context.WithValue(ctx, key(1), "value")
				logctx := &logcontext{
					logger:     lg,
					dispatcher: &mockdispatcher{},
				}
				og := enrichment
				defer func() { enrichment = og }()
				RegisterEnrichment(func(context.Context) map[string]any { return map[string]any{"key": "value"} })

				// ACT
				result := lg.withEnrichment(logctx, ctx)

				// ASSERT
				test.Value(t, result).DoesNotEqual(logctx, "new log context")
				test.That(t, result.ctx, "ctx").Equals(ctx)
				test.That(t, result.dispatcher, "dispatcher").Equals(logctx.dispatcher)
				test.That(t, result.fields, "fields").IsNil()
				test.That(t, result.xfields, "xfields").Equals(map[string]any{"key": "value"})
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.exec(t)
		})
	}
}
