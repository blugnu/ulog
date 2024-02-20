package ulog

import (
	"encoding/json"
	"errors"
)

type JsonFormatterOption func(*jsonfmt) error // JsonFormatterOption is a function for configuring a json formatter

// NewJSONFormatter returns a function that configures a json formatter.
func NewJSONFormatter(opt ...JsonFormatterOption) FormatterFactory {
	return func() (Formatter, error) {
		mf := &jsonfmt{
			keys: [numFields]string{
				TimeField:     "time",
				LevelField:    "level",
				MessageField:  "message",
				FileField:     "file",
				FunctionField: "function",
			},
			levels: [numLevels]string{
				TraceLevel: "trace",
				DebugLevel: "debug",
				InfoLevel:  "info",
				WarnLevel:  "warning",
				ErrorLevel: "error",
				FatalLevel: "fatal",
			},
		}

		errs := []error{}
		for _, cfg := range opt {
			errs = append(errs, cfg(mf))
		}
		if err := errors.Join(errs...); err != nil {
			return nil, err
		}

		return mf, nil
	}
}

var JSONFormatter, _ = NewJSONFormatter()()

type jsonfmt struct {
	keys   [numFields]string
	levels [numLevels]string
}

// Format implements a Formatter that writes log entries as JSON.
func (w *jsonfmt) Format(id int, e entry, b ByteWriter) {
	entry := map[string]any{
		w.keys[TimeField]:    e.Time,
		w.keys[LevelField]:   w.levels[e.Level],
		w.keys[MessageField]: e.Message,
	}

	if e.fields != nil {
		for k, v := range e.fields.m {
			entry[k] = v
		}
	}
	enc := json.NewEncoder(b)
	_ = enc.Encode(entry)
}
