package ulog

import (
	"os"
	"time"
)

// LogtailEndpoint configures the endpoint of the BetterStack Logs service.
func LogtailEndpoint(s string) LogtailOption {
	return func(t *logtail) error {
		return t.batch.configure(logtailEndpoint, s)
	}
}

// configures the maximum number of log entries to send to the BetterStack
// Logs service in a single request.  The default value is 16.
//
// A batch is sent to the BetterStack Logs service when the number of entries
// in the batch reaches this number, even if the max latency time has not
// been reached.
func LogtailMaxBatch(m int) LogtailOption {
	return func(t *logtail) error {
		t.batch.init(t.batch.batchHandler, m)
		return nil
	}
}

// configures the maximum time to wait before sending a batch of log entries
// to the BetterStack Logs service.  If the batch contains at least one
// entry it will be sent after this time has elapsed.
func LogtailMaxLatency(d time.Duration) LogtailOption {
	return func(t *logtail) error {
		t.maxLatency = d
		return nil
	}
}

// LogtailSourceToken configures the source token of the log entries sent to the
// BetterStack Logs service.
//
// The parameter to this function may be:
//
//   - the name of an environment variable holding the source token value
//   - the name of a file containing the source token value (and ONLY the source
//     value)
//   - a source token value (not recommended, to avoid leaking secrets in source)
func LogtailSourceToken(s string) LogtailOption {
	return func(t *logtail) error {
		// if s identifies an environment variable, the source token is
		// the value of that variable
		if v, ok := os.LookupEnv(s); ok {
			return t.batch.configure(logtailSourceToken, v)
		}

		// if s identifies a file re can read from, the source token is
		// the contents of that file
		if b, err := os.ReadFile(s); err == nil {
			return t.batch.configure(logtailSourceToken, string(b))
		}

		// otherwise, s is the source token
		return t.batch.configure(logtailSourceToken, s)
	}
}
