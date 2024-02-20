package ulog

import (
	"errors"
	"sync"
)

// MuxOption is a configuration function for a mux backend.
type MuxOption = func(*mux) error

// Mux configures a backend to dispatch log messages to multiple targets.
//
// The Mux factory accepts a slice of configuration functions.
func Mux(opts ...MuxOption) LoggerOption {
	return func(l *logger) error {
		mx := &mux{}
		mx.init()

		errs := []error{}
		for _, opt := range opts {
			errs = append(errs, opt(mx))
		}
		if err := errors.Join(errs...); err != nil {
			return err
		}

		// initialise the list of targets for each level
		for _, t := range mx.targets {
			for _, lv := range Levels {
				if lv <= t.Level {
					mx.levelTargets[lv] = append(mx.levelTargets[lv], t)
				}
			}
		}

		l.backend = mx
		return nil
	}
}

// formatref is a reference to a Formatter.  A formatref is created when a
// Format is introduced into a mux, either by using the Format() mux option
// or with a TargetFormat() target option.
//
// The formatref captures an index (idx) for a format; the idx is used to
// uniquely identify the specific format when caching formatter results.
type formatref struct {
	idx int
	Formatter
}

// mux is a backend implementation that dispatches log entries to multiple
// targets.
//
// The mux is initialised with a slice of targets, each of which is
// configured to receive log entries at a specific level.
//
// The mux is responsible for dispatching log entries to the targets
// according to their configured level.
//
// The mux is also responsible for initialising and running the targets.
// Initialisation of a target may involve opening a file or network
// connection, and running a target may involve starting a goroutine to
// write to a file or network connection.
//
// The mux is also responsible for closing the targets when the logger is
// closed.
//
// The mux establishes a channel to which log entries are sent as they are
// received and a goroutine to read entries from that channel, to be sent
// to each configured target.
type mux struct {
	formats      map[string]*formatref
	targets      []*target
	levelTargets [numLevels][]*target
	ch           chan entry
}

// initMux initialises a mux.
func (m *mux) init() {
	m.ch = make(chan entry, 100)
	m.formats = map[string]*formatref{}
	m.targets = []*target{}
	m.levelTargets = [numLevels][]*target{}
}

// close closes the mux channel.
func (mx *mux) close() {
	close(mx.ch)
}

// dispatch dispatches a log entry to the mux channel.
func (mx *mux) dispatch(e entry) {
	mx.ch <- e
}

// provides a run loop reading log entries from the mux channel and
// dispatching them to each target enabled for the log level of the
// entry.
//
// This function runs in a goroutine initiated by the start() function.
//
// The run loop terminates when the mux channel is closed, after
// which the stop() function is called of any target transport that
// implements such a function.
func (mx *mux) run() {
	for entry := range mx.ch {
		func() {
			for _, t := range mx.levelTargets[entry.Level] {
				t.dispatch(entry)
			}
		}()
	}

	for _, t := range mx.targets {
		if t, ok := t.transport.(interface{ stop() }); ok {
			t.stop()
		}
	}
}

// start initialises the mux, initialising and running the Transport on
// any target that implements an init() or run() function before
// starting the goroutine for the mux itself.
//
// a close function is returned which will close the mux then wait for
// the goroutines of any transports to terminate.  This allows the
// transports to complete the processing of log entries that may still
// be waiting in the mux or target channels at the point that the logger
// is closed, which otherwise might be lost.
//
// start() is called by the logger after completing backend configuration.
func (mx *mux) start() (func(), error) {
	type runnable interface{ run() }

	wg := &sync.WaitGroup{}
	for _, t := range mx.targets {
		if transport, ok := t.transport.(runnable); ok {
			wg.Add(1)
			go func() {
				defer wg.Done()
				transport.run()
			}()
		}
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		mx.run()
	}()

	cfn := func() {
		mx.close()
		wg.Wait()
	}

	return cfn, nil
}
