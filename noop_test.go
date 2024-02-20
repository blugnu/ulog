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
	defer test.ExpectPanic(nil).Assert(t)

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
			fn        func()
			callsExit bool
		}{
			{name: "ExitAlways (Fatal)", fn: func() { noop.logger.Fatal("test") }, callsExit: true},
			{name: "ExitAlways (Fatalf)", fn: func() { noop.logger.Fatalf("test") }, callsExit: true},
			{name: "Log(FatalLevel)", fn: func() { noop.logger.Log(FatalLevel, "test") }, callsExit: false},
		}
		for _, tc := range testcases {
			t.Run(tc.name, func(t *testing.T) {
				// ARRANGE
				og := ExitFn
				defer func() { ExitFn = og }()
				exitWasCalled := false
				ExitFn = func(int) { exitWasCalled = true }

				// ACT
				tc.fn()

				// ASSERT
				test.That(t, exitWasCalled).Equals(tc.callsExit)
			})
		}
	})
}

func TestNoopWithExitCode(t *testing.T) {
	// ARRANGE
	og := ExitFn
	defer func() { ExitFn = og }()
	exitCode := 0
	ExitFn = func(n int) { exitCode = n }

	// ACT
	noop := &nooplogger{}
	result := noop.WithExitCode(42)
	result.Fatal("test")

	// ASSERT
	test.That(t, result).Equals(&nooplogger{exitCode: 42})
	test.That(t, exitCode).Equals(42)
}
