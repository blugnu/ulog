package ulog

import (
	"io"
)

// import (
// 	"bytes"
// 	"io"
// 	"sync"
// )

type StdioTransportOption = func(*stdioTransport) error // StdioOption is a function for configuring a stdio transport

// Stdio returns a factory that configures a transport to log messages
// to an io.Writer.
func StdioTransport(w io.Writer) TransportFactory {
	return func() (transport, error) {
		t := &stdioTransport{}
		t.init(w)
		return t, nil
	}
}

// stdioTransport implements the Transport interface.
//
// It is used as a Transport by a mux target to write log
// entries to an io.Writer.
type stdioTransport struct {
	io.Writer
}

// init initialises a stdio transport with a specified Writer.
func (t *stdioTransport) init(w io.Writer) {
	t.Writer = w
}

// log implements the log method to satisfy the Transport
// interface. It writes the log entry to the configured io.Writer
// followed by a newline.
func (t *stdioTransport) log(b []byte) {
	// there is no need to copy the slice contents in this transport
	// as the output to the writer is synchronous; the target will
	// not be able to re-use the slice for subsequent log entries
	// until we have returned from this call
	_, _ = t.Write(b)
	_, _ = t.Write(buf.newline)
}
