package ulog

import (
	"context"
	"io"
)

// ByteWriter is an interface that implements both io.Writer and
// io.ByteWriter
type ByteWriter interface {
	io.Writer
	io.ByteWriter
}

// CloseFn is a function that closes a resource
type CloseFn func()

// Formatter is an interface implemented by a log formatter
type Formatter interface {
	Format(int, entry, ByteWriter)
}

// LevelLogger is an interface for a logger that is limited to emitting
// log messages at a specified level.
//
// A LevelLogger is obtained by calling Logger.AtLevel(Level).  If the
// specified level is not enabled on the Logger the returned LevelLogger
// will be a no-op logger.
//
// Example:
//
//	logger := ulog.FromContext(ctx)
//	debug := logger.AtLevel(ulog.DebugLevel)
//
//	debug.Log("this is a debug message")
type LevelLogger interface {
	Log(string)                            // Log emits a log entry at the level of the LevelLogger
	Logf(string, ...any)                   // Logf emits a log entry at the level of the LevelLogger, using a specified format string and args
	WithField(string, any) LevelLogger     // WithField returns a new LevelLogger that will add a specified field to all log entries
	WithFields(map[string]any) LevelLogger // WithFields returns a new LevelLogger that will add a specified set of fields to all log entries
}

// Logger is the interface used by applications and modules to emit
// log entries and provide enriched context for logging.
type Logger interface {
	Log(Level, string)                            // Log emits a log message at a specified level
	Logf(level Level, format string, args ...any) // Logf emits a log message at a specified level using a specified format string and args
	Debug(s string)                               // Debug emits a Debug level log message
	Debugf(format string, args ...any)            // Debugf emits a Debug level log message using a specified format string and args
	Error(err any)                                // Error emits an Error level log message consisting of err
	Errorf(format string, args ...any)            // Errorf emits an Error level log message using a specified format string and args
	Fatal(err any)                                // Fatal emits a Fatal level log message then calls ExitFn(n) with the current exit code (or 1 if not set)
	Fatalf(format string, args ...any)            // Fatalf emits a Fatal level log message using a specified format string and args, then calls ExitFn(n) with the current exit code (or 1 if not set)
	Info(s string)                                // Info emits an Info level log message
	Infof(format string, args ...any)             // Infof emits an Info level log message using a specified format string and args
	Trace(s string)                               // Trace emits a Trace level log message
	Tracef(format string, args ...any)            // Tracef emits a Trace level log message using a specified format string and args
	Warn(s string)                                // Warn emits a Warn level log message
	Warnf(format string, args ...any)             // Warnf emits a Warn level log message using a specified format string and args

	AtLevel(Level) LevelLogger          // AtLevel returns a new LevelLogger with the same Context and Fields (if any) as those on the receiver Logger
	WithContext(context.Context) Logger // WithContext returns a new Logger encapsulating the specific Context
	WithExitCode(int) Logger            // WithExitCode returns a new Logger with a specified exit code set
	WithField(string, any) Logger       // WithField returns a new Logger that will add a specified field to all log entries
	WithFields(map[string]any) Logger   // WithFields returns a new Logger that will add a specified set of fields to all log entries
}

// MockLog is an interface implemented by a mock logger that can be used
// to verify that log entries are emitted as expected.
type MockLog interface {
	ExpectEntry(...EntryExpectation)
	ExpectTrace(...EntryExpectation)
	ExpectDebug(...EntryExpectation)
	ExpectInfo(...EntryExpectation)
	ExpectWarn(...EntryExpectation)
	ExpectFatal(...EntryExpectation)
	ExpectationsWereMet() error
	Reset()
}

// these additional interfaces are used internally by ulog but are not exported

// mutex is an interface that implements a lock/unlock pair
type mutex interface {
	Lock()
	Unlock()
}

// pool is an interface that implements a pool of resources
type pool interface {
	Get() any
	Put(any)
}

// dispatcher is an interface that implements a dispatch method for handling
// a entry.  The dispatcher is responsible for releasing the entry when
// dispatch is complete.
//
// logger backends and mux targets implement the dispatcher interface.
type dispatcher interface {
	dispatch(entry)
}

// transport is an interface implemented by a log transport that
// accepts pre-formatted log messages.
type transport interface {
	log([]byte)
}
