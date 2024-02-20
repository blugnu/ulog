package ulog

import (
	"errors"
	"time"
)

type LogtailOption = func(*logtail) error // LogtailOption is a function that configures a logtail transport

// LogtailTransport returns a transport factory function to create and
// configure a LogtailTransport transport with specified configuration options
// applied.
func LogtailTransport(opts ...LogtailOption) func() (transport, error) {
	return func() (transport, error) {
		bh := newLogtailBatchHandler()
		bh.endpoint = "https://in.logs.betterstack.com"

		t := &logtail{
			ch:         make(chan []byte, 100),
			batch:      &Batch{},
			maxLatency: 10 * time.Second,
		}
		t.batch.init(bh, 16)

		errs := []error{}
		for _, opt := range opts {
			errs = append(errs, opt(t))
		}
		if err := errors.Join(errs...); err != nil {
			return nil, err
		}
		return t, nil
	}
}

// logtail implements a transport that sends log entries to the
// BetterStack Logs service (formerly known as Logtail) using the
// BetterStack Logs REST Api.
type logtail struct {
	ch         chan []byte
	batch      *Batch
	maxLatency time.Duration
}

// sends a formatted log entry to the transport.
func (t *logtail) log(b []byte) {
	// we need to copy the contents of the slice before sending to
	// the transport channel (asynchronous) as the slice is owned by
	// the target; if we don't copy the contents, the target will
	// re-use the slice for subsequent log entries
	buf := make([]byte, len(b))
	copy(buf, b)

	t.ch <- buf
}

// stop closes the channel over which log entries are received.
func (t *logtail) stop() {
	trace("logtail transport requested to stop...")
	close(t.ch)
}

// run is the goroutine run loop for the transport.  The run loop
// terminates when the channel over which log entries are received is
// closed.
//
// logs are read from the channel and added to a batch.  when the batch
// reaches a certain size, it is sent to the BetterStack Logs service
// and the batch is reset to receive new entries.
//
// if the channel is idle for a certain period of time, or the channel
// is closed, the current batch is sent to the BetterStack Logs service
// if it contains > 0 entries.
func (t *logtail) run() {
	batch := t.batch
loop:
	for {
		select {
		case entry := <-t.ch:
			if len(entry) == 0 {
				trace("logtail transport stopping...")
				batch.flush()
				break loop
			}
			batch.add(entry)
		case <-time.After(t.maxLatency):
			batch.flush()
		}
	}
	trace("logtail transport stopped")
}
