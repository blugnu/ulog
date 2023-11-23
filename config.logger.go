package ulog

import (
	"context"
	"fmt"
	"io"
	"os"
)

type LoggerOption = func(*logger) error           // LoggerOption is a function for configuring a logger
type FormatterFactory = func() (Formatter, error) // FormatterFactory is a function that returns a new Formatter

// NewLogger returns a new logger with the given configuration options applied and a
// close function that must be called to cleanly shutdown the logger and ensure
// that log entries are not lost.
//
// If there are any errors in the configuration, a nil logger is returned along
// with the error.  In this situation, the close function returned is a no-op;
// it is safe to call it, immediately or deferred.
//
// If the logger is configured without a backend, a stdio backend will be
// used with a logfmt formatter and os.Stdout as the output.
//
// If no Level is configured, the default level is Info.
func NewLogger(ctx context.Context, cfg ...LoggerOption) (Logger, func(), error) {
	logger, ic, err := initLogger(ctx, cfg...)
	if err != nil {
		return nil, nil, err
	}

	if logger.backend == nil {
		f, _ := Logfmt()()
		logger.backend = initStdioBackend(f, os.Stdout)
	}
	ic.dispatcher = logger.backend

	// a close function is always returned, even if the logger has no backend
	// or the backend does not require a close function
	//
	// the close function is a no-op by default
	//
	// if the backend implements a start function, the close function will
	// be set to the return value of the start function
	//
	// the close function is returned to the caller so that the caller can
	// ensure that the logger is cleanly shutdown in a normal exit process
	// so that log entries are not lost
	//
	// the close function is also called if the logger.exit() function is
	// called, prior to terminating the process (in which case the caller
	// will not have an opportunity to call the close function, even if
	// deferred)

	logger.closeFn = func() { /* NO-OP */ }
	if backend, ok := logger.backend.(interface{ start() (func(), error) }); ok {
		logger.closeFn, err = backend.start()
		if err != nil {
			return nil, nil, err
		}
	}
	return ic, logger.closeFn, nil
}

// LogCallsite returns a function that sets whether or not call site
// information is included in logs produced by a logger
func LogCallsite(e bool) LoggerOption {
	return func(l *logger) error {
		l.getCallsite = noCaller
		if e {
			l.getCallsite = caller
			_ = caller() // call callsite to perform the required run-once initialisation
		}
		return nil
	}
}

// LoggerBackend returns a function that sets the backend of a logger
func LoggerBackend(t dispatcher) LoggerOption {
	return func(l *logger) error {
		l.backend = t
		return nil
	}
}

// LoggerLevel returns a function that sets the log level of a logger
func LoggerLevel(level Level) LoggerOption {
	return func(l *logger) error {
		l.Level = level
		return nil
	}
}

// LoggerFormat sets the formatter of a logger.  This configuration option
// only makes sense for a non-muxing logger with a backend that supports
// configurable formatting.
//
// If configured without/before a backend being configured, a stdio
// backend will be installed using the supplied formatter.
//
// Returns ErrInvalidConfiguration error if configured on a mux logger.
func LoggerFormat(f FormatterFactory) LoggerOption {
	return func(l *logger) error {
		f, err := f()
		if err != nil {
			return err
		}

		switch backend := l.backend.(type) {
		case nil:
			l.backend = initStdioBackend(f, os.Stdout)
			return nil
		case interface{ setFormatter(Formatter) error }:
			return backend.setFormatter(f)
		}
		return fmt.Errorf("%w: backend (%T) does not support LoggerFormat", ErrInvalidConfiguration, l.backend)
	}
}

// LoggerOutput sets the io.Writer of a logger.  This configuration option
// only makes sense for a non-muxing logger.
//
// If configured without a backend, a stdio backend will be configured
// using a default logfmt formatter and the supplied io.Writer.
//
// Returns ErrInvalidConfiguration error if configured on a mux logger.
func LoggerOutput(out io.Writer) LoggerOption {
	return func(l *logger) error {
		switch backend := l.backend.(type) {
		case nil:
			f, _ := Logfmt()() // we can discard the error from the default Logfmt formatter factory as the default configuration cannot return an error
			l.backend = initStdioBackend(f, out)
			return nil
		case interface{ setOutput(io.Writer) error }:
			return backend.setOutput(out)
		}
		return fmt.Errorf("%w: backend (%T) does not support LoggerOutput", ErrInvalidConfiguration, l.backend)
	}
}
