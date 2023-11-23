package ulog

import (
	"errors"
	"fmt"
	"sync"
)

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
func initMux() *mux {
	return &mux{
		ch:           make(chan entry, 100),
		formats:      map[string]*formatref{},
		targets:      []*target{},
		levelTargets: [numLevels][]*target{},
	}
}

// MuxOption is a configuration function for a mux backend.
type MuxOption = func(*mux) error

// Mux configures a backend to dispatch log messages to multiple targets.
//
// The Mux factory accepts a slice of configuration functions.
func Mux(cfg ...MuxOption) LoggerOption {
	return func(l *logger) error {
		mx := initMux()

		errs := []error{}
		for _, cfg := range cfg {
			errs = append(errs, cfg(mx))
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

// close closes the mux channel.
func (mx *mux) close() {
	close(mx.ch)
}

// dispatch dispatches a log entry to the mux channel.
func (mx *mux) dispatch(e entry) {
	mx.ch <- e
}

// run reads log entries from the mux channel and dispatches them to each
// target.  the run loop terminates when the mux channel is closed, after
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

// Format registers a Formatter with the mux, with a specified id.  The id
// must be unique within the mux.
//
// A Formatter added to the mux may be shared by multiple targets by
// specifying a TargetFormat(id) option with the same id as the Formatter.
// The Formatter must already have been added to the mux before it can be
// referenced by a target.
//
// A Formatter that is not shared by multiple targets does not need to be
// added to the mux separately; a target-specific Formatter may be configured
// directly using the TargetFormat(Formatter) option for the relevant target.
func Format(id string, f FormatterFactory) func(*mux) error {
	return func(mx *mux) error {
		f, err := f()
		if err != nil {
			return err
		}

		if _, ok := mx.formats[id]; ok {
			return fmt.Errorf("format id %q: %w", id, ErrFormatAlreadyRegistered)
		}
		mx.formats[id] = &formatref{
			idx:       len(mx.formats),
			Formatter: f,
		}
		return nil
	}
}
