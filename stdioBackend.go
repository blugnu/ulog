package ulog

import (
	"bytes"
	"io"
	"os"
	"sync"
)

// stdioBackend implements the backend interface to provide
// simplex logging to an io.Writer.
type stdioBackend struct {
	Formatter
	io.Writer
	bufs pool
}

// init initialises a stdio backend with a specified Formatter
// and Writer.
func newStdioBackend(f Formatter, w io.Writer) *stdioBackend {
	be := &stdioBackend{
		Formatter: f,
		Writer:    w,
	}
	be.bufs = &sync.Pool{New: func() any { return bytes.NewBuffer(make([]byte, 0, 1024)) }}

	if be.Formatter == nil {
		be.Formatter, _ = LogfmtFormatter()()
	}
	if be.Writer == nil {
		be.Writer = os.Stdout
	}

	return be
}

// dispatch satisfies the backend interface, formatting each log entry
// and writing it to the configured io.Writer, appending char.newline.
func (stdio *stdioBackend) dispatch(e entry) {
	buf := stdio.bufs.Get().(*bytes.Buffer)
	defer stdio.bufs.Put(buf)

	buf.Reset()
	stdio.Format(0, e, buf)
	_ = buf.WriteByte(char.newline)

	_, _ = buf.WriteTo(stdio.Writer)
}

// SetFormatter sets the formatter of a stdio backend
func (stdio *stdioBackend) SetFormatter(f Formatter) error {
	stdio.Formatter = f
	return nil
}

// SetOutput sets the io.Writer of a stdio backend
func (stdio *stdioBackend) SetOutput(out io.Writer) error {
	stdio.Writer = out
	return nil
}
