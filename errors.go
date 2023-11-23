package ulog

import "errors"

var (
	ErrBackendNotConfigured    = errors.New("a backend must be configured first")
	ErrExpectationsNotMet      = errors.New("expectations were not met")
	ErrFormatAlreadyRegistered = errors.New("a format with this id is already registered")
	ErrInvalidFormatReference  = errors.New("invalid type for format; must be a Formatter or the (string) id of a Formatter previously added to the mux")
	ErrLogtailConfiguration    = errors.New("logtail transport configuration")
	ErrNoLoggerInContext       = errors.New("no logger in context")
	ErrKeyNotSupported         = errors.New("key not supported")
	ErrUnknownFormat           = errors.New("unknown format")
	ErrInvalidConfiguration    = errors.New("invalid configuration")
)
