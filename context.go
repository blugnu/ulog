package ulog

import "context"

// determines the behaviour of ulog.FromContext() when no logger is
// present in the supplied context and the caller did not explicitly
// request a logger:
//
//   - if false, then a no-op logger will be returned
//
//   - if true, then a panic(ErrNoContextInLogger) will occur
var ContextLoggerRequired bool

type key int
type ContextLoggerOption int

const loggerKey = key(1)
const Required = ContextLoggerOption(1)

// returns any logger present in the supplied Context.
//
// The result if no logger is present in the context depends on any
// ContextLoggerOption supplied and the value of ContextLoggerRequired:
//
//   - if ContextLoggerRequired is false and no ContextLoggerOption is
//     specified, a no-op logger is returned
//
//   - if ContextLoggerRequired is true or the ContextLoggerOption is
//     Required, a panic(ErrNoLoggerInContext) will occur
//
// Example:
//
//	log := ulog.FromContext(ctx, ulog.Required)
func FromContext(ctx context.Context, opt ...ContextLoggerOption) Logger {
	if lg := ctx.Value(loggerKey); lg != nil {
		return lg.(Logger).WithContext(ctx)
	}
	if ContextLoggerRequired || len(opt) > 0 && opt[0] == Required {
		panic(ErrNoLoggerInContext)
	}
	return noop.logger
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
