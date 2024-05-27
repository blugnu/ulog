package ulog

import "context"

type key int
type ContextLoggerOption int

const loggerKey = key(1)
const Required = ContextLoggerOption(1)

// returns any logger present in the supplied Context.
//
// The result if no logger is present in the context depends on whether
// a ContextLoggerOption is supplied:
//
//   - if called with ulog.Required and no logger is present, a panic
//     will occur (ErrNoLoggerInContext)
//
//   - if not called with ulog.Required, a no-op logger will be returned
//
// Example:
//
//	log := ulog.FromContext(ctx, ulog.Required)
func FromContext(ctx context.Context, opt ...ContextLoggerOption) Logger {
	if lg := ctx.Value(loggerKey); lg != nil {
		return lg.(Logger).WithContext(ctx)
	}
	if len(opt) > 0 && opt[0] == Required {
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
