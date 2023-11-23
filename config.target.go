package ulog

import (
	"bytes"
	"errors"
	"fmt"
)

type TargetOption = func(*mux, *target) error // TargetOption is a function that configures a target

// Target is a MuxOption that configures and adds a target to a mux.
func Target(cfg ...TargetOption) MuxOption {
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
			t.Formatter, _ = Logfmt()()
		}
		if t.transport == nil {
			t.transport, _ = StdioTransport()(mx, t)
		}

		mx.targets = append(mx.targets, t)
		return nil
	}
}

// TargetFormat sets the Formatter for a target.  The parameter to
// this function must be a Formatter or a (string) id of a Formatter
// previously added to the mux.
func TargetFormat(f any) TargetOption {
	return func(mx *mux, t *target) error {
		switch f := f.(type) {
		case string:
			if f, ok := mx.formats[f]; ok {
				t.formatIdx = f.idx
				t.Formatter = f.Formatter
				return nil
			}
			return fmt.Errorf("TargetFormat: %w: %s", ErrUnknownFormat, f)
		case func() (Formatter, error):
			fmt, err := f()
			if err != nil {
				return err
			}
			return TargetFormat(fmt)(mx, t)
		case Formatter:
			// an unregistered formatter is added to the mux and assigned a
			// unique id, required as a unique key/id for the Formatter to
			// cache formatted bytes in logcontext fields
			//
			// context fields are cached for unregistered formatters because
			// the same context may be used to emit multiple log entries
			// for which the formatted field bytes can be re-used
			id := len(mx.formats)
			key := fmt.Sprintf("unregistered format: %d", id)
			mx.formats[key] = &formatref{
				idx:       id,
				Formatter: f,
			}
			t.formatIdx = id
			t.Formatter = f
		default:
			return fmt.Errorf("TargetFormat: %w", ErrInvalidFormatReference)
		}
		return nil
	}
}

// TargetLevel sets the minimum Level of entries that will be dispatched
// to a target.
func TargetLevel(level Level) TargetOption {
	return func(_ *mux, t *target) error {
		t.Level = level
		return nil
	}
}

type TransportFactory = func(*mux, *target) (transport, error)

// TargetTransport sets the Transport for a target.
func TargetTransport(cfg TransportFactory) TargetOption {
	return func(mx *mux, tg *target) error {
		tr, err := cfg(mx, tg)
		if err != nil {
			return err
		}
		tg.transport = tr
		return nil
	}
}
