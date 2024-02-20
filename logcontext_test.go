package ulog

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/blugnu/test"
)

func TestLogContext(t *testing.T) {
	// ARRANGE
	type key int

	var (
		ctx = context.Background()
		sut = &logcontext{}
	)

	testcases := []struct {
		scenario string
		exec     func(t *testing.T)
	}{
		// fromContext tests
		{scenario: "fromContext/same context",
			exec: func(t *testing.T) {
				// ACT
				result := sut.fromContext(ctx)

				// ASSERT
				test.Value(t, result).Equals(sut)
			},
		},
		{scenario: "fromContext/new context",
			exec: func(t *testing.T) {
				// ARRANGE
				ctx = context.WithValue(ctx, key(1), 0)

				// ACT
				result := sut.fromContext(ctx)

				// ASSERT
				test.Value(t, result).DoesNotEqual(sut)
			},
		},

		// makeEntry tests
		{scenario: "makeEntry/level disabled",
			exec: func(t *testing.T) {
				// ARRANGE
				sut.logger.Level = InfoLevel

				// ACT
				result := sut.makeEntry(DebugLevel, "debug")

				// ASSERT
				test.Value(t, result).Equals(noop.entry)
			},
		},
		{scenario: "makeEntry/no fields",
			exec: func(t *testing.T) {
				// ACT
				result := sut.makeEntry(InfoLevel, "info")

				// ASSERT
				test.IsTrue(t, result.logcontext == sut, "same logcontext")
				test.That(t, result.logcontext.fields, "entry fields").IsNil()
				test.That(t, result.Level, "level").Equals(InfoLevel)
				test.That(t, result.Message, "message").Equals("info")
			},
		},
		{scenario: "makeEntry/merges unmerged fields",
			exec: func(t *testing.T) {
				// ARRANGE
				sut.xfields = map[string]any{"key": "value"}

				// ACT
				result := sut.makeEntry(InfoLevel, "info")

				// ASSERT
				test.IsTrue(t, result.logcontext == sut, "same logcontext")
				test.That(t, result.logcontext.fields.m, "entry fields").Equals(map[string]any{"key": "value"})
				test.That(t, result.Level, "level").Equals(InfoLevel)
				test.That(t, result.Message, "message").Equals("info")
			},
		},

		// new test
		{scenario: "new",
			exec: func(t *testing.T) {
				// ARRANGE
				mux := &mux{targets: []*target{}}
				sut.dispatcher = mux

				// ACT
				result := sut.new(ctx, sut.dispatcher, map[string]any{"key": "value"}, 42)

				// ASSERT
				test.That(t, result).Equals(&logcontext{
					ctx:        ctx,
					logger:     sut.logger,
					dispatcher: sut.dispatcher,
					xfields:    map[string]any{"key": "value"},
					exitCode:   42,
				})
			},
		},

		// AtLevel tests
		{scenario: "AtLevel/level disabled",
			exec: func(t *testing.T) {
				// ARRANGE
				sut.logger.Level = InfoLevel

				// ACT
				result := sut.AtLevel(DebugLevel)

				// ASSERT
				test.Value(t, result).Equals(noop.levellogger)
			},
		},
		{scenario: "AtLevel/level enabled",
			exec: func(t *testing.T) {
				// ACT
				result := sut.AtLevel(InfoLevel)

				// ASSERT
				test.That(t, result).Equals(&levelLogger{
					logcontext: sut,
					Level:      InfoLevel,
				})
			},
		},

		// LogTo tests
		{scenario: "LogTo/with valid target",
			exec: func(t *testing.T) {
				// ARRANGE
				mux := &mux{
					targets: []*target{{id: "id"}},
				}
				og := sut.dispatcher
				defer func() { sut.dispatcher = og }()
				sut.dispatcher = mux

				// ACT
				result, ok := sut.LogTo("id")

				// ASSERT
				test.IsTrue(t, ok, "found target")
				test.That(t, result.(*logcontext).dispatcher, "dispatcher").Equals(mux.targets[0])
			},
		},
		{scenario: "LogTo/with invalid target",
			exec: func(t *testing.T) {
				// ARRANGE
				mux := &mux{targets: []*target{}}
				og := sut.dispatcher
				defer func() { sut.dispatcher = og }()
				sut.dispatcher = mux

				// ACT
				result, ok := sut.LogTo("xid")

				// ASSERT
				test.IsFalse(t, ok, "found target")
				test.Value(t, result, "logger").Equals(noop.logger)
			},
		},

		// WithContext tests
		{scenario: "WithContext",
			exec: func(t *testing.T) {
				// ARRANGE
				ctx := context.WithValue(ctx, key(1), 0)

				// ACT
				result := sut.WithContext(ctx)

				// ASSERT
				test.That(t, result).Equals(&logcontext{
					ctx:        ctx,
					logger:     sut.logger,
					dispatcher: sut.dispatcher,
				})
			},
		},

		// WithField tests
		{scenario: "WithField",
			exec: func(t *testing.T) {
				// ACT
				result := sut.WithField("key", "value")

				// ASSERT
				test.That(t, result).Equals(&logcontext{
					ctx:        sut.ctx,
					logger:     sut.logger,
					dispatcher: sut.dispatcher,
					xfields:    map[string]any{"key": "value"},
				})
			},
		},

		// WithFields tests
		{scenario: "WithFields",
			exec: func(t *testing.T) {
				// ACT
				result := sut.WithFields(map[string]any{
					"key-1": "value",
					"key-2": "value",
				})

				// ASSERT
				test.That(t, result).Equals(&logcontext{
					ctx:        sut.ctx,
					logger:     sut.logger,
					dispatcher: sut.dispatcher,
					xfields: map[string]any{
						"key-1": "value",
						"key-2": "value",
					},
				})
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			// ARRANGE
			sut = &logcontext{ctx: ctx,
				logger: &logger{
					Level:       InfoLevel,
					enabled:     sut.isLevelEnabled,
					getCallsite: noCaller,
				},
			}
			sut.logger.enrich = sut.logger.noEnrichment // TODO: this is horrible - can this be refactored to avoid it?  a nil receiver method maybe?

			// ACT
			tc.exec(t)
		})
	}
}

func TestLogContext_Logging(t *testing.T) {
	// ARRANGE
	var (
		ctx    = context.Background()
		mock   = &mockdispatcher{}
		logger = &logger{}
		et     = time.Date(2010, 9, 8, 7, 6, 5, 432100, time.UTC)

		closeWasCalled = false
		exitWasCalled  = false
	)
	sut, _ := logger.init(ctx, LoggerBackend(mock))
	logger.closeFn = func() { closeWasCalled = true }

	ognow := now
	defer func() { now = ognow }()
	now = func() time.Time { return et }

	// logs emitted at FatalLevel are expected to close the logger and call exit
	ogexit := ExitFn
	defer func() { ExitFn = ogexit }()
	ExitFn = func(int) { exitWasCalled = true }

	// run testcases covering Log and Logf at each log level, when the level is
	// both enabled and disabled
	{
		for _, level := range Levels {
			testcases := []struct {
				scenario string
				logf     bool // set true for a Logf() test case, otherwise test Log()
				entry
			}{
				{scenario: "Log/%s/level disabled", logf: false, entry: noop.entry},
				{scenario: "Logf/%s/level disabled", logf: true, entry: noop.entry},
				{scenario: "Log/%s/level enabled", logf: false, entry: entry{logcontext: sut, Time: et, Level: level, Message: "log"}},
				{scenario: "Logf/%s/level enabled", logf: true, entry: entry{logcontext: sut, Time: et, Level: level, Message: "log with arg"}},
			}
			for _, tc := range testcases {
				t.Run(fmt.Sprintf(tc.scenario, level), func(t *testing.T) {
					// ARRANGE
					closeWasCalled = false
					exitWasCalled = false
					mock.Reset()
					logger.enabled = func(context.Context, Level) bool { return strings.Contains(tc.scenario, "level enabled") }

					// ACT
					switch tc.logf {
					case false:
						sut.Log(level, "log")
					case true:
						sut.Logf(level, "log with %s", "arg")
					}

					// ASSERT
					test.That(t, mock.entry, "emitted").Equals(tc.entry)

					// using Log() or Logf() even at FatalLevel should NOT call exit
					test.IsFalse(t, exitWasCalled, "calls exit")
					test.IsFalse(t, closeWasCalled, "closes the logger")
				})
			}
		}
	}

	// the remaining tests cover the different logging methods, and the different
	// behaviours when the level is enabled and disabled
	{
		// ARRANGE
		testcases := []struct {
			scenario string
			act      func()
			entry
		}{
			// logging with simple messages or errors
			// level disabled
			{scenario: "Trace/level disabled", act: func() { sut.Trace("log") }, entry: noop.entry},
			{scenario: "Debug/level disabled", act: func() { sut.Debug("log") }, entry: noop.entry},
			{scenario: "Info/level disabled", act: func() { sut.Info("log") }, entry: noop.entry},
			{scenario: "Warn/level disabled", act: func() { sut.Warn("log") }, entry: noop.entry},
			{scenario: "Error/level disabled", act: func() { sut.Error("msg") }, entry: noop.entry},
			{scenario: "Fatal/level disabled", act: func() { sut.Fatal("msg") }, entry: noop.entry},
			// level enabled
			{scenario: "Trace/level enabled", act: func() { sut.Trace("log") }, entry: entry{logcontext: sut, Time: et, Level: TraceLevel, Message: "log"}},
			{scenario: "Debug/level enabled", act: func() { sut.Debug("log") }, entry: entry{logcontext: sut, Time: et, Level: DebugLevel, Message: "log"}},
			{scenario: "Info/level enabled", act: func() { sut.Info("log") }, entry: entry{logcontext: sut, Time: et, Level: InfoLevel, Message: "log"}},
			{scenario: "Warn/level enabled", act: func() { sut.Warn("log") }, entry: entry{logcontext: sut, Time: et, Level: WarnLevel, Message: "log"}},
			{scenario: "Error(error)/level enabled", act: func() { sut.Error(errors.New("error")) }, entry: entry{logcontext: sut, Time: et, Level: ErrorLevel, Message: "error"}},
			{scenario: "Error(string)/level enabled", act: func() { sut.Error("error") }, entry: entry{logcontext: sut, Time: et, Level: ErrorLevel, Message: "error"}},
			{scenario: "Error(int)/level enabled", act: func() { sut.Error(1) }, entry: entry{logcontext: sut, Time: et, Level: ErrorLevel, Message: "1"}},
			{scenario: "Fatal(error)/level enabled", act: func() { sut.Fatal(errors.New("error")) }, entry: entry{logcontext: sut, Time: et, Level: FatalLevel, Message: "error"}},
			{scenario: "Fatal(string)/level enabled", act: func() { sut.Fatal("error") }, entry: entry{logcontext: sut, Time: et, Level: FatalLevel, Message: "error"}},
			{scenario: "Fatal(int)/level enabled", act: func() { sut.Fatal(1) }, entry: entry{logcontext: sut, Time: et, Level: FatalLevel, Message: "1"}},

			// logging with format verbs and args
			// level disabled
			{scenario: "Tracef/level disabled", act: func() { sut.Tracef("log with %s", "arg") }, entry: noop.entry},
			{scenario: "Debugf/level disabled", act: func() { sut.Debugf("log with %s", "arg") }, entry: noop.entry},
			{scenario: "Infof/level disabled", act: func() { sut.Infof("log with %s", "arg") }, entry: noop.entry},
			{scenario: "Warnf/level disabled", act: func() { sut.Warnf("log with %s", "arg") }, entry: noop.entry},
			{scenario: "Errorf/level disabled", act: func() { sut.Errorf("msg with %s", "arg") }, entry: noop.entry},
			{scenario: "Fatalf/level disabled", act: func() { sut.Fatalf("msg with %s", "arg") }, entry: noop.entry},
			// level enabled
			{scenario: "Tracef/level enabled", act: func() { sut.Tracef("log with %s", "arg") }, entry: entry{logcontext: sut, Time: et, Level: TraceLevel, Message: "log with arg"}},
			{scenario: "Debugf/level enabled", act: func() { sut.Debugf("log with %s", "arg") }, entry: entry{logcontext: sut, Time: et, Level: DebugLevel, Message: "log with arg"}},
			{scenario: "Infof/level enabled", act: func() { sut.Infof("log with %s", "arg") }, entry: entry{logcontext: sut, Time: et, Level: InfoLevel, Message: "log with arg"}},
			{scenario: "Warnf/level enabled", act: func() { sut.Warnf("log with %s", "arg") }, entry: entry{logcontext: sut, Time: et, Level: WarnLevel, Message: "log with arg"}},
			{scenario: "Errorf/level enabled", act: func() { sut.Errorf("msg with %s", "arg") }, entry: entry{logcontext: sut, Time: et, Level: ErrorLevel, Message: "msg with arg"}},
			{scenario: "Fatalf/level enabled", act: func() { sut.Fatalf("msg with %s", "arg") }, entry: entry{logcontext: sut, Time: et, Level: FatalLevel, Message: "msg with arg"}},
		}
		for _, tc := range testcases {
			t.Run(tc.scenario, func(t *testing.T) {
				// ARRANGE
				closeWasCalled = false
				exitWasCalled = false
				logger.enabled = func(context.Context, Level) bool { return strings.Contains(tc.scenario, "level enabled") }
				mock.Reset()
				expectedToExit := strings.HasPrefix(tc.scenario, "Fatal")

				// ACT
				tc.act()

				// ASSERT
				test.That(t, mock.entry).Equals(tc.entry)
				test.That(t, exitWasCalled, "calls exit").Equals(expectedToExit)
				test.That(t, closeWasCalled, "closes the logger").Equals(expectedToExit)
			})
		}
	}
}

func TestLogContext_ExitCode(t *testing.T) {
	// ARRANGE
	lc := &logcontext{}

	// ACT
	result := lc.WithExitCode(42).(*logcontext)

	// ASSERT
	test.That(t, result).Equals(&logcontext{exitCode: 42})
}
