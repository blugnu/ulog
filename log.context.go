package ulog

import (
	"context"
	"fmt"

	"github.com/blugnu/errorcontext"
)

// logcontext captures the context in which log entries are created.
//
// An initial context is created when a Logger is initialized.
type logcontext struct {
	ctx        context.Context // ctx is the context from which the logcontext was created
	*fields                    // fields are the fields that are added to all log entries created from the context
	*logger                    // logger is the logger to which log entries are sent
	dispatcher                 // the backend to which log entries are sent; LogTo() creates a new logcontext with a backend assigned to the specified target, otherwise the dispatcher is the *logger.backend

	// xfields are fields that have been added to the context but not yet merged with the fields from the original context (which are still in the fields member)
	// if xfields is not nil, they are merged with original context fields the first time a new entry is created from the context, replacing
	// the existing fields member (xfields is then set to nil to prevent them being merged again when any subsequent entries are emitted)
	xfields map[string]any
}

// fromContext returns a new logcontext with fields enriched from a
// specified Context.
//
// If the specified Context is the same as the Context from which the
// logcontext was created or if there are no enrichment functions
// registered then the logcontext is returned unchanged.
func (lc *logcontext) fromContext(ctx context.Context) *logcontext {
	if ctx == lc.ctx {
		return lc
	}
	return lc.enrich(lc, ctx)
}

// makeEntry creates a new entry at a specified level with a specified string.
//
// If the logcontext is not enabled for the specified level then nil is returned.
//
// If the logcontext has any xfields then a new set of fields are created for
// the context by merging the xfields with the current fields.  The xfields
// are then set to nil to prevent them being merged again when any subsequent
// entries are emitted from the same context.
//
// The entry is created from the logcontext logger's entry pool.
func (lc *logcontext) makeEntry(level Level, s string) entry {
	if !lc.enabled(lc.ctx, level) {
		return noop.entry
	}

	if lc.xfields != nil {
		lc.fields = lc.fields.merge(lc.xfields)
		lc.xfields = nil
	}

	entry := entry{}
	entry.logcontext = lc
	entry.callsite = lc.getCallsite()
	entry.Time = now().UTC()
	entry.Level = level
	entry.Message = s

	return entry
}

// makeEntryf creates a new entry at a specified level with a specified format
// string and args.
//
// If the logcontext is not enabled for the specified level then nil is returned.
func (lc *logcontext) makeEntryf(level Level, s string, args ...any) entry {
	entry := lc.makeEntry(level, s)
	if entry != noop.entry {
		entry.Message = fmt.Sprintf(s, args...)
	}
	return entry
}

// new returns a new logcontext.
func (lc *logcontext) new(ctx context.Context, d dispatcher, xf map[string]any) *logcontext {
	return &logcontext{ctx, lc.fields, lc.logger, d, xf}
}

// Log emits a log entry of a specified level to the log.
func (lc *logcontext) Log(level Level, s string) {
	if !lc.enabled(lc.ctx, level) {
		return
	}
	lc.log(lc.makeEntry(level, s))
}

// Logf emits a log entry of a specified level to the log.  The message is
// formatted using the specified format string and args.
func (lc *logcontext) Logf(level Level, format string, args ...any) {
	if !lc.enabled(lc.ctx, level) {
		return
	}
	lc.log(lc.makeEntryf(level, format, args...))
}

// Trace emits a log entry at the Trace level to the log.
func (lc *logcontext) Trace(s string) {
	if !lc.enabled(lc.ctx, TraceLevel) {
		return
	}
	lc.log(lc.makeEntry(TraceLevel, s))
}

// Tracef emits a log entry at the Trace level to the log.  The message is
// formatted using the specified format string and args.
func (lc *logcontext) Tracef(format string, args ...any) {
	if !lc.enabled(lc.ctx, TraceLevel) {
		return
	}
	lc.log(lc.makeEntryf(TraceLevel, format, args...))
}

// Debug emits a log entry at the Debug level to the log.
func (lc *logcontext) Debug(s string) {
	if !lc.enabled(lc.ctx, DebugLevel) {
		return
	}
	lc.log(lc.makeEntry(DebugLevel, s))
}

// Debugf emits a log entry at the Debug level to the log.  The message is
// formatted using the specified format string and args.
func (lc *logcontext) Debugf(format string, args ...any) {
	if !lc.enabled(lc.ctx, DebugLevel) {
		return
	}
	lc.log(lc.makeEntryf(DebugLevel, format, args...))
}

// Info emits a log entry at the Info level to the log.
func (lc *logcontext) Info(s string) {
	if !lc.enabled(lc.ctx, InfoLevel) {
		return
	}
	lc.log(lc.makeEntry(InfoLevel, s))
}

// Infof emits a log entry at the Info level to the log.  The message is
// formatted using the specified format string and args.
func (lc *logcontext) Infof(format string, args ...any) {
	if !lc.enabled(lc.ctx, InfoLevel) {
		return
	}
	lc.log(lc.makeEntryf(InfoLevel, format, args...))
}

// Warn emits a log entry at the Warn level to the log.
func (lc *logcontext) Warn(s string) {
	if !lc.enabled(lc.ctx, WarnLevel) {
		return
	}
	lc.log(lc.makeEntry(WarnLevel, s))
}

// Warnf emits a log entry at the Warn level to the log.  The message is
// formatted using the specified format string and args.
func (lc *logcontext) Warnf(format string, args ...any) {
	if !lc.enabled(lc.ctx, WarnLevel) {
		return
	}
	lc.log(lc.makeEntryf(WarnLevel, format, args...))
}

// Error emits an error or string as an Error level entry to the log.
//
// If logging an error wrapping a specific context then the error is
// logged using an entry  enriched with any information in the error context.
//
// If logging a string then the string is logged as-is.
//
// If logging any other type then the type is logged using `fmt.Sprintf` with
// the `%v` format.
func (lc *logcontext) Error(err any) {
	if !lc.enabled(lc.ctx, ErrorLevel) {
		return
	}
	switch err := err.(type) {
	case error:
		ctx := errorcontext.From(lc.ctx, err)
		lc := lc.fromContext(ctx)
		lc.log(lc.makeEntry(ErrorLevel, err.Error()))
	case string:
		lc.log(lc.makeEntry(ErrorLevel, err))
	default:
		lc.log(lc.makeEntryf(ErrorLevel, "%v", []any{err}...))
	}
}

// Errorf emits a log entry at the Error level to the log.  The message is
// formatted using the specified format string and args.
func (lc *logcontext) Errorf(format string, args ...any) {
	if !lc.enabled(lc.ctx, ErrorLevel) {
		return
	}
	lc.log(lc.makeEntryf(ErrorLevel, format, args...))
}

// Fatal emits a log entry at the Fatal level to the log then calls exit(1)
func (lc *logcontext) Fatal(err any) {
	switch err := err.(type) {
	case error:
		ctx := errorcontext.From(lc.ctx, err)
		lc := lc.fromContext(ctx)
		lc.log(lc.makeEntry(FatalLevel, err.Error()))
	case string:
		lc.log(lc.makeEntry(FatalLevel, err))
	default:
		lc.log(lc.makeEntryf(FatalLevel, "%v", []any{err}...))
	}
	lc.exit(1)
}

// Fatalf emits a log entry at the Fatal level to the log then calls exit(1).
// The message is formatted using the specified format string and args.
func (lc *logcontext) Fatalf(format string, args ...any) {
	lc.log(lc.makeEntryf(FatalLevel, format, args...))
	lc.exit(1)
}

// AtLevel returns a LevelLogger that is limited to emitting log messages at
// a specified level.  If the specified level is not enabled on the Logger
// the returned LevelLogger will be a no-op logger.
func (lc *logcontext) AtLevel(level Level) LevelLogger {
	if !lc.enabled(lc.ctx, level) {
		return noop.levellogger
	}
	return &levelLogger{lc.new(lc.ctx, lc.dispatcher, nil), level}
}

// LogTo returns a Logger that is limited to emitting log messages to a
// specified target.  If the specified target is not registered on the
// Logger the returned Logger will be a no-op logger and the additional
// bool value returned will be false (otherwise it is true).
func (lc *logcontext) LogTo(id string) (Logger, bool) {
	if mux, ok := lc.dispatcher.(*mux); ok {
		for _, t := range mux.targets {
			if t.id == id {
				return lc.new(lc.ctx, t, nil), true
			}
		}
	}
	return noop.logger, false
}

// WithContext returns a new Logger initialised with a specified context.
// If the specified context is the same as the context from which the Logger
// was created or if there are no enrichment functions registered then the
// receiver context is returned unchanged.
func (lc *logcontext) WithContext(ctx context.Context) Logger {
	return lc.fromContext(ctx)
}

// WithField returns a new Logger with an additional field.
func (lc *logcontext) WithField(key string, value any) Logger {
	return lc.new(lc.ctx, lc.dispatcher, map[string]any{key: value})
}

// WithFields returns a new Logger with additional fields.
func (lc *logcontext) WithFields(fields map[string]any) Logger {
	return lc.new(lc.ctx, lc.dispatcher, fields)
}
