package ulog

import (
	"bytes"
	"io"
	"sync"
)

// stdio implements both backend and Transport interfaces.
//
// It may be used as a Transport by a mux target to write log
// entries to an io.Writer or may be used as a backend for a
// non-muxing logger to write directly to an io.Writer.
type stdio struct {
	Formatter
	io.Writer
	bufs pool
}

// initStdio initialises a stdio backend with a specified Formatter
// and Writer.
//
// if Formatter is nil, a logfmt formatter will be used with default
// configuration.
func initStdioBackend(f Formatter, w io.Writer) *stdio {
	return &stdio{
		Formatter: f,
		Writer:    w,
		bufs:      &sync.Pool{New: func() any { return bytes.NewBuffer(make([]byte, 0, 1024)) }},
	}
}

// initStdio initialises a stdio backend with a specified Formatter
// and Writer.
//
// if Formatter is nil, a logfmt formatter will be used with default
// configuration.
func initStdioTransport(w io.Writer) *stdio {
	return &stdio{
		Writer: w,
	}
}

// dispatch satisfies the backend interface, formatting each log entry
// and writing it to the configured io.Writer.
func (stdio *stdio) dispatch(e entry) {
	buf := stdio.bufs.Get().(*bytes.Buffer)
	defer stdio.bufs.Put(buf)

	buf.Reset()
	stdio.Format(0, e, buf)
	_ = buf.WriteByte(char.newline)

	_, _ = buf.WriteTo(stdio.Writer)
}

// setFormatter sets the formatter of a stdio backend
func (stdio *stdio) setFormatter(f Formatter) error {
	stdio.Formatter = f
	return nil
}

// setOutput sets the io.Writer of a stdio backend
func (stdio *stdio) setOutput(out io.Writer) error {
	stdio.Writer = out
	return nil
}

// log implements the log method to satsify the Transport
// interface. It writes the log entry to the configured io.Writer.
func (stdio *stdio) log(b []byte) {
	// there is no need to copy the slice contents in this transport
	// as the output to the writer is synchronous; the target will
	// not be able to re-use the slice for subsequent log entries
	// until we have returned from this call
	_, _ = stdio.Write(b)
	_, _ = stdio.Write(buf.newline)
}
