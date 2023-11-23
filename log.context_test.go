package ulog

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/blugnu/test"
)

func TestLogContext_fromContext_returnsReceiver(t *testing.T) {
	// ARRANGE
	type key int

	ctx := context.Background()
	sut := &logcontext{ctx: ctx, logger: &logger{getCallsite: noCaller}}

	// ARRANGE
	testcases := []struct {
		name     string
		ctx      context.Context
		enriched bool
		result   bool
	}{
		{name: "og.ctx == ctx, enriched", ctx: ctx, enriched: true, result: true},
		{name: "og.ctx == ctx, not enriched", ctx: ctx, result: true},
		{name: "og.ctx != ctx, enriched", ctx: context.WithValue(ctx, key(1), 0), enriched: true, result: false},
		{name: "og.ctx != ctx, not enriched", ctx: context.WithValue(ctx, key(1), 0), result: false},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// ARRANGE
			sut.logger.enrich = sut.logger.noEnrichment
			if tc.enriched {
				sut.logger.enrich = sut.logger.withEnrichment
			}

			// ACT
			result := sut.fromContext(tc.ctx)

			// ASSERT
			wanted := tc.result
			got := result == sut
			if wanted != got {
				t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
			}
		})
	}
}

func TestLogContext_makeEntry(t *testing.T) {
	// ARRANGE
	ctx := context.Background()
	og := now
	defer func() { now = og }()
	now = func() time.Time { return time.Date(2010, 9, 8, 7, 6, 5, 432100, time.UTC) }

	type args struct {
		Level
		string
		withArg bool
	}
	type result struct {
		entry
		*fields
		xfields map[string]any
	}

	testcases := []struct {
		name   string
		fields map[string]any
		args
		result result
	}{
		{name: "level disabled", args: args{DebugLevel, "debug", false}, result: result{entry: noop.entry}},
		{name: "no fields", args: args{InfoLevel, "info", false},
			result: result{
				entry: entry{Time: now().UTC(), Level: InfoLevel, Message: "info"},
			},
		},
		{name: "fields merged", fields: map[string]any{"key": "value"}, args: args{InfoLevel, "info", false},
			result: result{
				entry:   entry{Time: now().UTC(), Level: InfoLevel, Message: "info"},
				fields:  &fields{m: map[string]any{"key": "value"}},
				xfields: nil,
			},
		},
		{name: "level disabled, makeEntryf", args: args{DebugLevel, "debug with %s", true}, result: result{entry: noop.entry}},
		{name: "no fields, makeEntryf", args: args{InfoLevel, "info with %s", true},
			result: result{
				entry: entry{Time: now().UTC(), Level: InfoLevel, Message: "info with arg"},
			},
		},
		{name: "fields merged, makeEntryf", fields: map[string]any{"key": "value"}, args: args{InfoLevel, "info with %s", true},
			result: result{
				entry:   entry{Time: now().UTC(), Level: InfoLevel, Message: "info with arg"},
				fields:  &fields{m: map[string]any{"key": "value"}},
				xfields: nil,
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// ARRANGE
			_, sut, _ := initLogger(ctx)
			if !tc.result.entry.noop {
				tc.result.logcontext = sut
			}
			sut.xfields = tc.fields

			// ACT
			var got entry
			if tc.withArg {
				got = sut.makeEntryf(tc.args.Level, tc.args.string, "arg")
			} else {
				got = sut.makeEntry(tc.args.Level, tc.args.string)
			}

			// ASSERT
			wanted := tc.result.entry
			if !reflect.DeepEqual(wanted, got) {
				t.Errorf("\nwanted %s\ngot    %s", wanted, got)
			}

		})
	}
}

func TestLogContext(t *testing.T) {
	// ARRANGE
	mock := &mockdispatcher{}
	logger, sut, _ := initLogger(context.Background(), LoggerBackend(mock))
	et := time.Date(2010, 9, 8, 7, 6, 5, 432100, time.UTC)

	{
		og := now
		defer func() { now = og }()
	}
	{
		og := ExitFn
		defer func() { ExitFn = og }()
	}
	now = func() time.Time { return et }
	closeWasCalled := false
	exitWasCalled := false
	logger.closeFn = func() { closeWasCalled = true }
	ExitFn = func(int) { exitWasCalled = true }

	t.Run("Log/Logf", func(t *testing.T) {
		for _, level := range Levels {
			testcases := []struct {
				levelEnabled bool
				entry        entry
				fn           func(Logger)
			}{
				{levelEnabled: false, fn: func(l Logger) { l.Log(level, "log") }, entry: noop.entry},
				{levelEnabled: false, fn: func(l Logger) { l.Logf(level, "log with %s", "arg") }, entry: noop.entry},
				{levelEnabled: true, fn: func(l Logger) { l.Log(level, "log") }, entry: entry{logcontext: sut, Time: et, Level: level, Message: "log"}},
				{levelEnabled: true, fn: func(l Logger) { l.Logf(level, "log with %s", "arg") }, entry: entry{logcontext: sut, Time: et, Level: level, Message: "log with arg"}},
			}
			for _, tc := range testcases {
				t.Run(fmt.Sprintf("Log(%s, ...) (level enabled %v)", level, tc.levelEnabled), func(t *testing.T) {
					// ARRANGE
					mock.Reset()
					logger.enabled = func(context.Context, Level) bool { return tc.levelEnabled }

					// ACT
					tc.fn(sut)

					// ASSERT
					wanted := tc.entry
					got := mock.entry
					if !reflect.DeepEqual(wanted, got) {
						t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
					}
				})
			}
		}
	})

	t.Run("Trace/Debug/Info/Warn", func(t *testing.T) {
		testcases := []struct {
			name         string
			fn           func(string)
			levelEnabled bool
			entry        entry
			callsExit    bool
		}{
			{name: "Trace", fn: sut.Trace, levelEnabled: false, entry: noop.entry},
			{name: "Debug", fn: sut.Debug, levelEnabled: false, entry: noop.entry},
			{name: "Info", fn: sut.Info, levelEnabled: false, entry: noop.entry},
			{name: "Warn", fn: sut.Warn, levelEnabled: false, entry: noop.entry},
			{name: "Trace", fn: sut.Trace, levelEnabled: true, entry: entry{logcontext: sut, Time: et, Level: TraceLevel, Message: "log"}},
			{name: "Debug", fn: sut.Debug, levelEnabled: true, entry: entry{logcontext: sut, Time: et, Level: DebugLevel, Message: "log"}},
			{name: "Info", fn: sut.Info, levelEnabled: true, entry: entry{logcontext: sut, Time: et, Level: InfoLevel, Message: "log"}},
			{name: "Warn", fn: sut.Warn, levelEnabled: true, entry: entry{logcontext: sut, Time: et, Level: WarnLevel, Message: "log"}},
		}
		for _, tc := range testcases {
			t.Run(fmt.Sprintf("%s (level enabled %v)", tc.name, tc.levelEnabled), func(t *testing.T) {
				// ARRANGE
				exitWasCalled = false
				mock.Reset()
				logger.enabled = func(context.Context, Level) bool { return tc.levelEnabled }

				// ACT
				tc.fn("log")

				// ASSERT
				wanted := tc.entry
				got := mock.entry
				if !reflect.DeepEqual(wanted, got) {
					t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
				}

				t.Run("closes the logger", func(t *testing.T) {
					wanted := false
					got := closeWasCalled
					if wanted != got {
						t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
					}
				})

				t.Run("calls exit()", func(t *testing.T) {
					wanted := tc.callsExit
					got := exitWasCalled
					if wanted != got {
						t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
					}
				})
			})
		}
	})

	t.Run("Tracef/Debugf/Infof/Warnf/Errorf/Fatalf", func(t *testing.T) {
		testcases := []struct {
			name         string
			fn           func(string, ...any)
			levelEnabled bool
			entry        entry
			closesLogger bool
			callsExit    bool
		}{
			{name: "Tracef", fn: sut.Tracef, levelEnabled: false, entry: noop.entry},
			{name: "Debugf", fn: sut.Debugf, levelEnabled: false, entry: noop.entry},
			{name: "Infof", fn: sut.Infof, levelEnabled: false, entry: noop.entry},
			{name: "Warnf", fn: sut.Warnf, levelEnabled: false, entry: noop.entry},
			{name: "Errorf", fn: sut.Errorf, levelEnabled: false, entry: noop.entry},
			{name: "Fatalf", fn: sut.Fatalf, levelEnabled: false, entry: noop.entry, closesLogger: true, callsExit: true},
			{name: "Tracef", fn: sut.Tracef, levelEnabled: true, entry: entry{logcontext: sut, Time: et, Level: TraceLevel, Message: "log with arg"}},
			{name: "Debugf", fn: sut.Debugf, levelEnabled: true, entry: entry{logcontext: sut, Time: et, Level: DebugLevel, Message: "log with arg"}},
			{name: "Infof", fn: sut.Infof, levelEnabled: true, entry: entry{logcontext: sut, Time: et, Level: InfoLevel, Message: "log with arg"}},
			{name: "Warnf", fn: sut.Warnf, levelEnabled: true, entry: entry{logcontext: sut, Time: et, Level: WarnLevel, Message: "log with arg"}},
			{name: "Errorf", fn: sut.Errorf, levelEnabled: true, entry: entry{logcontext: sut, Time: et, Level: ErrorLevel, Message: "log with arg"}},
			{name: "Fatalf", fn: sut.Fatalf, levelEnabled: true, entry: entry{logcontext: sut, Time: et, Level: FatalLevel, Message: "log with arg"}, closesLogger: true, callsExit: true},
		}
		for _, tc := range testcases {
			t.Run(fmt.Sprintf("%s (level enabled %v)", tc.name, tc.levelEnabled), func(t *testing.T) {
				// ARRANGE
				closeWasCalled = false
				exitWasCalled = false
				mock.Reset()
				logger.enabled = func(context.Context, Level) bool { return tc.levelEnabled }

				// ACT
				tc.fn("log with %s", "arg")

				// ASSERT
				wanted := tc.entry
				got := mock.entry
				if !reflect.DeepEqual(wanted, got) {
					t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
				}

				t.Run("closes the logger", func(t *testing.T) {
					test.Equal(t, tc.closesLogger, closeWasCalled)
				})

				t.Run("calls exit()", func(t *testing.T) {
					test.Equal(t, tc.callsExit, exitWasCalled)
				})
			})
		}
	})

	t.Run("AtLevel", func(t *testing.T) {
		t.Run("when level disabled", func(t *testing.T) {
			// ARRANGE
			logger.enabled = func(ctx context.Context, l Level) bool { return false }

			// ACT
			got := sut.AtLevel(InfoLevel)
			// ASSERT
			wanted := noop.levellogger
			if wanted != got {
				t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
			}
		})

		t.Run("when level enabled", func(t *testing.T) {
			// ARRANGE
			logger.enabled = func(ctx context.Context, l Level) bool { return true }

			// ACT
			got := sut.AtLevel(InfoLevel)

			// ASSERT
			wanted := &levelLogger{
				logcontext: sut,
				Level:      InfoLevel,
			}
			if !reflect.DeepEqual(wanted, got) {
				t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
			}
		})
	})

	t.Run("Error/Fatal", func(t *testing.T) {
		// ARRANGE
		testcases := []struct {
			name    string
			fn      func(any)
			enabled bool
			arg     any
			entry
		}{
			{name: "disabled Error()", fn: sut.Error, enabled: false, entry: noop.entry},
			{name: "Error(error)", fn: sut.Error, enabled: true, arg: errors.New("error"), entry: entry{logcontext: sut, Time: et, Level: ErrorLevel, Message: "error"}},
			{name: "Error(string)", fn: sut.Error, enabled: true, arg: "error", entry: entry{logcontext: sut, Time: et, Level: ErrorLevel, Message: "error"}},
			{name: "Error(int)", fn: sut.Error, enabled: true, arg: 1, entry: entry{logcontext: sut, Time: et, Level: ErrorLevel, Message: "1"}},
			{name: "disabled Fatal()", fn: sut.Fatal, enabled: false, entry: noop.entry},
			{name: "Fatal(error)", fn: sut.Fatal, enabled: true, arg: errors.New("error"), entry: entry{logcontext: sut, Time: et, Level: FatalLevel, Message: "error"}},
			{name: "Fatal(string)", fn: sut.Fatal, enabled: true, arg: "error", entry: entry{logcontext: sut, Time: et, Level: FatalLevel, Message: "error"}},
			{name: "Fatal(int)", fn: sut.Fatal, enabled: true, arg: 1, entry: entry{logcontext: sut, Time: et, Level: FatalLevel, Message: "1"}},
		}
		for _, tc := range testcases {
			t.Run(tc.name, func(t *testing.T) {
				// ARRANGE
				mock.Reset()
				logger.enabled = func(context.Context, Level) bool { return tc.enabled }

				// ACT
				tc.fn(tc.arg)

				// ASSERT
				wanted := tc.entry
				got := mock.entry
				if !reflect.DeepEqual(wanted, got) {
					t.Errorf("\nwanted %v\ngot    %v", wanted, got)
				}
			})
		}
	})

	t.Run("FatalError", func(t *testing.T) {
		// ARRANGE
		testcases := []struct {
			name    string
			enabled bool
			arg     any
			entry
		}{
			{name: "disabled", entry: noop.entry},
			{name: "FatalError(error)", arg: errors.New("error"), enabled: true, entry: entry{logcontext: sut, Time: et, Level: FatalLevel, Message: "error"}},
			{name: "FatalError(string)", arg: "error", enabled: true, entry: entry{logcontext: sut, Time: et, Level: FatalLevel, Message: "error"}},
			{name: "FatalError(int)", arg: 1, enabled: true, entry: entry{logcontext: sut, Time: et, Level: FatalLevel, Message: "1"}},
		}
		for _, tc := range testcases {
			t.Run(tc.name, func(t *testing.T) {
				// ARRANGE
				mock.Reset()
				logger.enabled = func(context.Context, Level) bool { return tc.enabled }

				// ACT
				sut.Fatal(tc.arg)

				// ASSERT
				wanted := tc.entry
				got := mock.entry
				if !reflect.DeepEqual(wanted, got) {
					t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
				}
			})
		}
	})

	t.Run("LogTo", func(t *testing.T) {
		// ARRANGE
		mux := &mux{
			targets: []*target{{id: "id"}},
		}
		og := sut.dispatcher
		defer func() { sut.dispatcher = og }()
		sut.dispatcher = mux

		t.Run("with valid target", func(t *testing.T) {
			// ACT
			got, ok := sut.LogTo("id")

			// ASSERT
			t.Run("finds target", func(t *testing.T) {
				wanted := true
				got := ok
				if wanted != got {
					t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
				}
			})

			t.Run("returns targetted logger", func(t *testing.T) {
				wanted := &logcontext{
					ctx:        sut.ctx,
					logger:     sut.logger,
					dispatcher: mux.targets[0],
				}
				if !reflect.DeepEqual(wanted, got) {
					t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
				}
			})
		})

		t.Run("with invalid target", func(t *testing.T) {
			// ACT
			got, ok := sut.LogTo("xid")

			// ASSERT
			t.Run("finds no target", func(t *testing.T) {
				wanted := false
				got := ok
				if wanted != got {
					t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
				}
			})

			t.Run("returns noop logger", func(t *testing.T) {
				wanted := true
				_, got := got.(*nooplogger)
				if wanted != got {
					t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
				}
			})
		})
	})

	t.Run("WithContext", func(t *testing.T) {
		// ARRANGE
		type key string
		ctx := context.WithValue(context.Background(), key("key"), "value")

		// ACT
		got := sut.WithContext(ctx)

		// ASSERT
		wanted := &logcontext{
			ctx:        ctx,
			logger:     sut.logger,
			dispatcher: sut.dispatcher,
		}
		if !reflect.DeepEqual(wanted, got) {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})

	t.Run("WithField", func(t *testing.T) {
		// ACT
		got := sut.WithField("key", "value")

		// ASSERT
		wanted := &logcontext{
			ctx:        sut.ctx,
			logger:     sut.logger,
			dispatcher: sut.dispatcher,
			xfields:    map[string]any{"key": "value"},
		}
		if !reflect.DeepEqual(wanted, got) {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})

	t.Run("WithFields", func(t *testing.T) {
		// ACT
		got := sut.WithFields(map[string]any{"key": "value"})

		// ASSERT
		wanted := &logcontext{
			ctx:        sut.ctx,
			logger:     sut.logger,
			dispatcher: sut.dispatcher,
			xfields:    map[string]any{"key": "value"},
		}
		if !reflect.DeepEqual(wanted, got) {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})

	t.Run("new(.. nil) with callsite reporting", func(t *testing.T) {
		// ARRANGE
		og := logger.getCallsite
		defer func() { logger.getCallsite = og }()
		logger.getCallsite = caller

		// ACT
		got := sut.new(sut.ctx, sut.dispatcher, nil)

		// ASSERT
		wanted := &logcontext{
			ctx:        sut.ctx,
			logger:     sut.logger,
			dispatcher: sut.dispatcher,
		}
		if !reflect.DeepEqual(wanted, got) {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})
}
