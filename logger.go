package ulog

import (
	"context"
	"errors"
)

// logger provides a configurable logger. The logger provides a sync.Pool
// of entries that are reused to reduce allocations.  The logger also
// provides a dispatcher that is responsible for dispatching log entries
// to a backend.
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

// initLogger initialises a logger with the given configuration options
func initLogger(ctx context.Context, cfg ...LoggerOption) (*logger, *logcontext, error) {
	lg := &logger{
		Level:       InfoLevel, // default level is Info
		getCallsite: noCaller,  // callsite logging is not enabled by default
	}

	// the logger will apply no enrichment unless at least one enrichment
	// function has been registered
	lg.enrich = lg.noEnrichment
	if len(enrichment) > 0 {
		lg.enrich = lg.withEnrichment
	}

	// the log level enablement is determined by levelEnabled unless
	// overridden in configuration
	lg.enabled = lg.levelEnabled

	// apply configuration, collecting any errors
	errs := []error{}
	for _, cfg := range cfg {
		errs = append(errs, cfg(lg))
	}
	if err := errors.Join(errs...); err != nil {
		return nil, nil, err
	}

	// create the initial context for the logger
	ic := &logcontext{
		ctx:    ctx,
		logger: lg,
	}

	return lg, ic, nil
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

// levelEnabled returns true if the given level is enabled for the logger
//
// This is the default implementation of the enabled function.  It may
// replaced by the SetEnablement configuration option (future enhancement).
func (l *logger) levelEnabled(ctx context.Context, level Level) bool {
	return level <= l.Level
}

// noEnrichment returns a new logcontext with the specified context using the
// same dispatcher as the receiver, with no additional fields.
//
// This is the implementation of the enrich function used when no enrichment
// functions are registered (since in that case enrichment cannot result in
// additional fields being derived from the new context).
func (l *logger) noEnrichment(og *logcontext, ctx context.Context) *logcontext {
	return og.new(ctx, og.dispatcher, nil)
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

	return og.new(ctx, og.dispatcher, erm)
}
