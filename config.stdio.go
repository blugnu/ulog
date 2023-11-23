package ulog

import (
	"errors"
	"io"
	"os"
)

type StdioOption = func(*stdio) error // StdioOption is a function for configuring a stdio transport

// StdioTransport returns a factory that configures a transport to log messages
// to an io.Writer.
func StdioTransport(opt ...StdioOption) TransportFactory {
	return func(*mux, *target) (transport, error) {
		so := initStdioTransport(os.Stdout)

		errs := []error{}
		for _, cfg := range opt {
			errs = append(errs, cfg(so))
		}
		if err := errors.Join(errs...); err != nil {
			return nil, err
		}

		if so.Formatter == nil {
			so.Formatter, _ = Logfmt()()
		}

		return so, nil
	}
}

func StdioOutput(out io.Writer) StdioOption {
	return func(so *stdio) error {
		return so.setOutput(out)
	}
}
