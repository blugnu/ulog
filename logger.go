package ulog

import (
	"context"
	"errors"
)

// LoggerOption is a function for configuring a logger
type LoggerOption = func(*logger) error

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
func NewLogger(ctx context.Context, cfg ...LoggerOption) (Logger, CloseFn, error) {
	cfn := func() { /* NO-OP */ }

	logger := &logger{}
	ic, err := logger.init(ctx, cfg...)
	if err != nil {
		return nil, cfn, err
	}

	if logger.backend == nil {
		logger.backend = newStdioBackend(nil, nil)
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

	logger.closeFn = cfn
	if backend, ok := logger.backend.(interface{ start() (func(), error) }); ok {
		logger.closeFn, err = backend.start()
		if err != nil {
			return nil, cfn, err
		}
	}
	return ic, logger.closeFn, nil
}

// logger provides a configurable logger. The logger is responsible for
// dispatching log entries to a configured backend.
//
// Methods for emitting log entries are provided by a separate logcontext,
// referencing a logger.
//
// Output of log entries is handled by the logger.backend, which is
// responsible for formatting and writing log entries to an output (or
// outputs).
type logger struct {
	backend dispatcher
	Level
	closeFn     func()                                         // a function that closes the logger backend; set to noOp by default
	enrich      func(*logcontext, context.Context) *logcontext // a function that returns a new logcontext with the specified context using the same dispatcher as the receiver, with additional fields derived from the context; set to noEnrichment by default (no additional fields)
	enabled     func(context.Context, Level) bool              // a function that returns true if the specified level is enabled for the logger; set to levelEnabled by default
	getCallsite func() *callsite                               // a function that returns the first non-ulog call site in the caller stack; set to noCallSite by default (always returns nil)
}

// createLogger creates a new logger and applies the supplied LoggerOptions
func (lg *logger) init(ctx context.Context, cfg ...LoggerOption) (*logcontext, error) {
	lg.Level = InfoLevel      // default level is Info
	lg.getCallsite = noCaller // callsite logging is not enabled by default

	// the logger will apply no enrichment unless at least one enrichment
	// function has been registered
	lg.enrich = lg.noEnrichment
	if len(enrichment) > 0 {
		lg.enrich = lg.withEnrichment
	}

	// log level enablement is determined by the isLevelEnabled method
	// unless overridden in the options
	lg.enabled = lg.isLevelEnabled

	// apply options, collecting any errors
	errs := []error{}
	for _, cfg := range cfg {
		errs = append(errs, cfg(lg))
	}
	if err := errors.Join(errs...); err != nil {
		return nil, err
	}

	// create the initial logcontext for the logger
	ic := &logcontext{
		ctx:      ctx,
		logger:   lg,
		exitCode: 1,
	}

	return ic, nil
}

// exit calls the `ExitFn` with the specified exit code.  Code paths in `ulog`
// that require termination of the process (e.g. `log.FatalError()`) call this
// `exit` function which in turn calls the `ExitFn` func var.
//
// Any logger close function is called prior to calling `ExitFn`, to ensure
// that log entries are not lost.
//
// To prevent `ulog` causing a process to terminate, replace `ExitFn`.
func (l *logger) exit(code int) {
	l.closeFn()
	exit(code)
}

// log sends a log entry to the logger's backend
func (l *logger) log(e entry) {
	if e.noop {
		return
	}
	l.backend.dispatch(e)
}

// isLevelEnabled returns true if the given level is enabled for the logger
//
// This is the default implementation of the enabled function.  It may
// replaced by the SetEnablement configuration option (future enhancement).
func (l *logger) isLevelEnabled(ctx context.Context, level Level) bool {
	return level <= l.Level
}

// noEnrichment returns a new logcontext with the specified context using the
// same dispatcher as the receiver, with no additional fields.
//
// This is the implementation of the enrich function used when no enrichment
// functions are registered (since in that case enrichment cannot result in
// additional fields being derived from the new context).
func (l *logger) noEnrichment(og *logcontext, ctx context.Context) *logcontext {
	return og.new(ctx, og.dispatcher, nil, og.exitCode)
}

// withEnrichment returns a new logcontext with the specified context using the
// same dispatcher as the receiver, with additional fields derived from the
// context using the registered enrichment functions.
//
// This is the implementation of the enrich function used when enrichment
// functions are registered.
func (l *logger) withEnrichment(og *logcontext, ctx context.Context) *logcontext {
	erm := map[string]any{}
	for _, fn := range enrichment {
		enrichment := fn(ctx)
		for k, v := range enrichment {
			erm[k] = v
		}
	}

	return og.new(ctx, og.dispatcher, erm, og.exitCode)
}
