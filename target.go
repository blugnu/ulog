package ulog

import (
	"bytes"
	"errors"
	"os"
)

type TargetOption = func(*mux, *target) error // TargetOption is a function that configures a target

// MuxTarget is a MuxOption that configures and adds a target to a mux.
func MuxTarget(cfg ...TargetOption) MuxOption {
	return func(mx *mux) error {
		t := &target{
			buf: bytes.NewBuffer(make([]byte, 0, 1024)),
		}

		// apply configuration
		errs := []error{}
		for _, cfg := range cfg {
			errs = append(errs, cfg(mx, t))
		}
		if err := errors.Join(errs...); err != nil {
			return err
		}

		// if no formatter or transport has been configured, use the default
		if t.Formatter == nil {
			t.Formatter, _ = LogfmtFormatter()()
		}
		if t.transport == nil {
			t.transport, _ = StdioTransport(os.Stdout)()
		}

		mx.targets = append(mx.targets, t)
		return nil
	}
}

// target is a combination of a Level, a Formatter and a Transport.
// A target is used by a mux to format and send log entries to a
// specific destination.
//
// The Level determines which log entries are dispatched to the
// target.  The Formatter formats the log entries.  The Transport
// sends the formatted log entries to the required destination.
type target struct {
	id        string        // unique id for the target
	Level                   // minimum level of logs dispatched to the target
	formatIdx int           // index of the formatter in the mux
	Formatter               // formats log entries
	transport               // sends formatted log entries to some destination
	buf       *bytes.Buffer // buffer used for formatting log entries
}

// close closes the target by calling the stop function on the
// Transport, if implemented.
func (t *target) close() {
	if t, ok := t.transport.(interface{ stop() }); ok {
		t.stop()
	}
}

// dispatch dispatches a log entry to the target.  The entry is
// formatted and sent to the target Transport's Log function.
//
// dispatch is not thread-safe, using a single, shared buffer for
// all calls to the function; the buffer is managed by the target.
//
// if the transport Log() function is asynchronous then the
// Transport is responsible for making its own copy of the slice
// content BEFORE returning from the Log() function.
func (t *target) dispatch(e entry) {
	t.buf.Reset()
	t.Format(t.formatIdx, e, t.buf)

	// HERE BE DRAGONS!
	//
	// t.buf.Bytes(), which is sent to the transport.Log() function,
	// is the slice which backs the target bytes buffer; it will be
	// re-used for the next log entry once the Log() function has
	// returned.
	//
	// If the transport is asynchronous (e.g. logtail, which batches
	// logs over a channel), then the TRANSPORT must copy the slice
	// before returning from the Log() call.  If the transport is
	// synchronous (e.g. stdio), then the slice can be used directly.
	//
	// i.e. the TRANSPORT is responsible for copying the slice if
	// required because only the TRANSPORT knows if it IS required.
	//
	// This improves the efficiency of the target by avoiding copying
	// slices that do not need to be copied.

	t.log(t.buf.Bytes())
}
