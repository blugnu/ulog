package ulog

import (
	"context"
	"testing"

	"github.com/blugnu/test"
)

func TestNoop(t *testing.T) {
	// with the exception of log.Fatal() and log.Fatalf() there is nothing
	// really to test here; the noop types are just empty structs that
	// implement interfaces with NO-OP stubs
	//
	// this test merely exercises those stubs to ensure that they do not
	// panic

	// ARRANGE/ASSERT
	test.ExpectPanic(nil).IsRecovered(t)

	// ACT
	noop.Close()

	noop.levellogger.Log("test")
	noop.levellogger.Logf("test %v", "test")
	noop.levellogger.WithField("test", "test").Log("test")
	noop.levellogger.WithField("test", "test").Logf("test %v", "test")
	noop.levellogger.WithFields(map[string]any{"test": "test"}).Log("test")
	noop.levellogger.WithFields(map[string]any{"test": "test"}).Logf("test %v", "test")

	TestLogger := func(logger Logger) {
		logger.Trace("test")
		logger.Tracef("test %v", "test")
		logger.Debug("test")
		logger.Debugf("test %v", "test")
		logger.Info("test")
		logger.Infof("test %v", "test")
		logger.Warn("test")
		logger.Warnf("test %v", "test")
		logger.Error("test")
		logger.Errorf("test %v", "test")
		logger.AtLevel(TraceLevel).Log("test")
		logger.AtLevel(TraceLevel).Logf("test %v", "test")
		logger.Log(TraceLevel, "test")
		logger.Logf(TraceLevel, "test %v", "test")
	}
	TestLogger(noop.logger)
	TestLogger(noop.logger.WithContext(context.Background()))
	TestLogger(noop.logger.WithField("test", "test"))
	TestLogger(noop.logger.WithFields(map[string]any{"test": "test"}))
	TestLogger(noop.logger.WithLevel(TraceLevel))
	TestLogger(noop.logger.LogTo("test"))

	t.Run("fatal logs", func(t *testing.T) {
		// ARRANGE
		testcases := []struct {
			name      string
			cfg       FatalExitOption
			fn        func()
			callsExit bool
		}{
			{name: "ExitAlways (Fatal)", cfg: ExitAlways, fn: func() { noop.logger.Fatal("test") }, callsExit: true},
			{name: "ExitAlways (Fatalf)", cfg: ExitAlways, fn: func() { noop.logger.Fatalf("test") }, callsExit: true},
			{name: "ExitNever (Fatal)", cfg: ExitNever, fn: func() { noop.logger.Fatal("test") }, callsExit: false},
			{name: "ExitNever (Fatalf)", cfg: ExitNever, fn: func() { noop.logger.Fatalf("test") }, callsExit: false},
			{name: "ExitWhenLogged (Fatal)", cfg: ExitWhenLogged, fn: func() { noop.logger.Fatal("test") }, callsExit: false},
			{name: "ExitWhenLogged (Fatalf)", cfg: ExitWhenLogged, fn: func() { noop.logger.Fatalf("test") }, callsExit: false},
		}
		for _, tc := range testcases {
			t.Run(tc.name, func(t *testing.T) {
				// ARRANGE
				{
					og := ExitOnFatalLog
					defer func() { ExitOnFatalLog = og }()
				}
				{
					og := ExitFn
					defer func() { ExitFn = og }()
				}
				ExitOnFatalLog = tc.cfg
				exitWasCalled := false
				ExitFn = func(int) { exitWasCalled = true }

				// ACT
				tc.fn()

				// ASSERT
				t.Run("calls ExitFn", func(t *testing.T) {
					test.Equal(t, tc.callsExit, exitWasCalled)
				})
			})
		}
	})
}
