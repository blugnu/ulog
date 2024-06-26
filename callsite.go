package ulog

import (
	"runtime"
	"strings"
)

var ulogframes = 3

type callsite struct {
	function string
	file     string
	line     int
}

// caller returns the callsite of the first non-ulog caller in the call stack.
func caller() *callsite {
	const (
		maxCallerDepth = 25 // restrict the lookback frames to avoid runaway lookups
		ulogpkg        = "github.com/blugnu/ulog"
	)

	pcs := make([]uintptr, maxCallerDepth)
	depth := runtime.Callers(ulogframes, pcs)
	frames := runtime.CallersFrames(pcs[:depth])

	for f, again := frames.Next(); again; f, again = frames.Next() {
		// if the caller isn't part of this package, we're done
		if !strings.HasPrefix(f.Function, ulogpkg) {
			cs := &callsite{
				function: f.Function,
				file:     f.File,
				line:     f.Line,
			}
			return cs //nolint:scopelint
		}
	}

	// if we got here, we failed to find the caller's context
	return nil
}

// noCaller is used when call-site logging is disabled
func noCaller() *callsite { return nil }
