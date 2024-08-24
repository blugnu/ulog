package ulog

import (
	"bytes"
	"encoding/json"
	"errors"
)

type JSONFormatterOption func(*jsonfmt) error // a function for configuring a json formatter

// NewJSONFormatter returns a function that configures a json formatter.
func JSONFormatter(opt ...JSONFormatterOption) FormatterFactory {
	return func() (Formatter, error) {
		mf := &jsonfmt{
			keys: [numFields]string{
				TimeField:             "time",
				LevelField:            "level",
				MessageField:          "message",
				CallsiteFileField:     "file",
				CallsiteFunctionField: "function",
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
			if err, ok := v.(error); ok {
				v = err.Error()
			}
			entry[k] = v
		}
	}

	// json encoder appends a trailing \n which we do not want
	jb := bytes.NewBuffer(nil)
	enc := json.NewEncoder(jb)
	_ = enc.Encode(entry)
	_, _ = b.Write(jb.Bytes()[0 : jb.Len()-1])
}
