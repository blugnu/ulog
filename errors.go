package ulog

import "errors"

var (
	ErrBackendNotConfigured    = errors.New("a backend must be configured first")
	ErrFormatAlreadyRegistered = errors.New("a format with this id is already registered")
	ErrInvalidConfiguration    = errors.New("invalid configuration")
	ErrInvalidFormatReference  = errors.New("invalid type for format; must be a Formatter or the (string) id of a Formatter previously added to the mux")
	ErrKeyNotSupported         = errors.New("key not supported")
	ErrLogtailConfiguration    = errors.New("logtail transport configuration")
	ErrNoLoggerInContext       = errors.New("no logger in context")
	ErrNotImplemented          = errors.New("not implemented")
	ErrUnknownFormat           = errors.New("unknown format")

	// errors returns by the mock listener when expectations are not met
	ErrExpectationsNotMet      = errors.New("expectations were not met")
	ErrMalformedLogEntry       = errors.New("log entry did not meet expectations")
	ErrMissingExpectedLogEntry = errors.New("missing an expected log entry")
	ErrUnexpectedLogEntry      = errors.New("unexpected log entry")
)
