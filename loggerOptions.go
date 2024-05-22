package ulog

import (
	"fmt"
	"io"
	"os"
)

type FormatterFactory = func() (Formatter, error) // FormatterFactory is a function that returns a new Formatter

// LogCallsite returns a function that sets whether or not call site
// information is included in logs produced by a logger
func LogCallsite(e bool) LoggerOption {
	return func(l *logger) error {
		if e {
			l.getCallsite = caller
			_ = caller() // call callsite to perform the required run-once initialisation
		}
		return nil
	}
}

// LoggerFormat sets the formatter of a logger.  This configuration option
// only makes sense for a non-muxing logger with a backend that supports
// configurable formatting.
//
// If configured without/before a backend being configured, a stdio
// backend will be installed using the supplied formatter writing to
// os.Stdout.
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
			l.backend = newStdioBackend(f, os.Stdout)
			return nil

		case interface{ SetFormatter(Formatter) error }:
			return backend.SetFormatter(f)

		default:
			return fmt.Errorf("%w: backend (%T) does not support LoggerFormat", ErrInvalidConfiguration, l.backend)
		}
	}
}

// LoggerLevel returns a function that sets the log level of a logger
func LoggerLevel(level Level) LoggerOption {
	return func(l *logger) error {
		l.Level = level
		return nil
	}
}

// LoggerOutput sets the io.Writer of a logger.  This configuration option
// only makes sense for a non-muxing logger.
//
// Returns ErrInvalidConfiguration error if configured on a mux logger.
func LoggerOutput(out io.Writer) LoggerOption {
	return func(l *logger) error {
		switch backend := l.backend.(type) {
		case nil:
			l.backend = newStdioBackend(nil, out)
			return nil

		case interface{ SetOutput(io.Writer) error }:
			return backend.SetOutput(out)

		default:
			return fmt.Errorf("%w: backend (%T) does not support LoggerOutput", ErrInvalidConfiguration, l.backend)
		}
	}
}
