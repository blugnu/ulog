package ulog

import (
	"fmt"
)

// MuxFormat registers a Formatter with the mux, with a specified id.  The id
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
func MuxFormat(id string, f FormatterFactory) func(*mux) error {
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
