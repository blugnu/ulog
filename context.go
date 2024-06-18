package ulog

import "context"

type key int

const loggerKey = key(1)

// ContextLoggerOption values determine the behaviour of ulog.FromContext
// when no logger is present in the context.
//
// # Values
//
//	NoopIfNotPresent    // a no-op logger will be returned
//
//	PanicIfNotPresent   // a panic will occur (ErrNoLoggerInContext)
//
//	NilIfNotPresent     // nil will be returned
type ContextLoggerOption int

const (
	// NoopIfNotPresent is a ContextLoggerOption that specifies that if no logger is
	// present in the context, a no-op logger should be returned; this is also
	// the default behavior if no ContextLoggerOption is supplied.
	//
	// The caller can safely emit logs without checking for the presence
	// of a logger.
	NoopIfNotPresent = ContextLoggerOption(0)

	// PanicIfNotPresent is a ContextLoggerOption that specifies that if no logger is
	// present in the context, a panic should occur (ErrNoLoggerInContext).
	PanicIfNotPresent = ContextLoggerOption(1)

	// NilIfNotPresent is a ContextLoggerOption that specifies that if no logger is
	// present in the context, nil should be returned.  It is then the
	// responsibility of the caller to handle the absence of a logger.
	NilIfNotPresent = ContextLoggerOption(2)
)

// Required is equivalent to PanicIfNotPresent.
//
// Deprecated: new code should use PanicIfNotPresent.
const Required = PanicIfNotPresent

// FromContext returns any logger present in the supplied Context.
// The result if no logger is present in the context depends on the
// value of the (optional) ContextLoggerOption supplied:
//
//	NoopIfNotPresent    // a no-op logger will be returned
//
//	PanicIfNotPresent   // a panic will occur (ErrNoLoggerInContext)
//
//	NilIfNotPresent     // nil will be returned
//
// 0, 1 or many ContextLoggerOption values may be specified:
//
//	0       // NoopIfNotPresent is assumed
//	1       // the specified option is applied
//	>1      // the first specified option is applied
//
// Example:
//
//	log := ulog.FromContext(ctx, ulog.Required)
func FromContext(ctx context.Context, opt ...ContextLoggerOption) Logger {
	if lg := ctx.Value(loggerKey); lg != nil {
		return lg.(Logger).WithContext(ctx)
	}

	def := NoopIfNotPresent
	if len(opt) > 0 {
		def = opt[0]
	}

	switch def {
	case PanicIfNotPresent:
		panic(ErrNoLoggerInContext)
	case NilIfNotPresent:
		return nil
	default:
		return noop.logger
	}
}

// returns a new context with the supplied logger added to it.
//
// The supplied logger will be enriched with the context to avoid
// incurring this encrichment every time the logger is retrieved
// from the context (if the context is further enriched, the logger
//
//	then a new
//
// ).
//
// Example:
//
//	ctx := ulog.ContextWithLogger(ctx, logger)
func ContextWithLogger(ctx context.Context, lg Logger) context.Context {
	lg = lg.WithContext(ctx)
	return context.WithValue(ctx, loggerKey, lg)
}
