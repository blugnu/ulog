package ulog

import (
	"context"
)

// noopimpl is a no-op level logger.  All receiver methods are nil-safe no-ops
// (ore return references to nil-safe no-ops) so there is no need to actually
// allocate the structs.
type noopimpl struct {
	entry       entry
	logger      *nooplogger
	levellogger *nooplevellogger
}

// noop holds nil references to the nil-safe no-op structs.
var noop = &noopimpl{
	entry: entry{noop: true},
}

// Close implements a no-op Close method.
func (*noopimpl) Close() { /* NO-OP */ }

// nooplogger is a no-op logger.  All receiver methods are nil-safe no-ops (or
// return a reference to a nil-safe no-op)
type nooplogger struct{ exitCode int }

func (*nooplogger) Trace(string)          { /* NO-OP */ }
func (*nooplogger) Tracef(string, ...any) { /* NO-OP */ }
func (*nooplogger) Debug(string)          { /* NO-OP */ }
func (*nooplogger) Debugf(string, ...any) { /* NO-OP */ }
func (*nooplogger) Info(string)           { /* NO-OP */ }
func (*nooplogger) Infof(string, ...any)  { /* NO-OP */ }
func (*nooplogger) Warn(string)           { /* NO-OP */ }
func (*nooplogger) Warnf(string, ...any)  { /* NO-OP */ }
func (*nooplogger) Error(any)             { /* NO-OP */ }
func (*nooplogger) Errorf(string, ...any) { /* NO-OP */ }

// nooplogger is the Nul() logger
func (logger *nooplogger) Fatal(any) {
	switch {
	case logger != nil:
		exit(logger.exitCode)
	default:
		exit(1)
	}
}
func (logger *nooplogger) Fatalf(string, ...any) {
	logger.Fatal("")
}

func (*nooplogger) AtLevel(Level) LevelLogger          { return noop.levellogger }
func (*nooplogger) Log(Level, string)                  { /* NO-OP */ }
func (*nooplogger) Logf(Level, string, ...any)         { /* NO-OP */ }
func (*nooplogger) LogTo(string) Logger                { return noop.logger }
func (*nooplogger) WithContext(context.Context) Logger { return noop.logger }
func (*nooplogger) WithExitCode(n int) Logger          { return &nooplogger{n} }
func (*nooplogger) WithField(string, any) Logger       { return noop.logger }
func (*nooplogger) WithFields(map[string]any) Logger   { return noop.logger }
func (*nooplogger) WithLevel(Level) Logger             { return noop.logger }

// nooplevellogger is a no-op level logger.  All receiver methods are nil-safe
// no-ops (or return a reference to a nil-safe no-op)
type nooplevellogger struct{}

func (*nooplevellogger) Log(string)                            { /* NO-OP */ }
func (*nooplevellogger) Logf(string, ...any)                   { /* NO-OP */ }
func (*nooplevellogger) WithField(string, any) LevelLogger     { return noop.levellogger }
func (*nooplevellogger) WithFields(map[string]any) LevelLogger { return noop.levellogger }
