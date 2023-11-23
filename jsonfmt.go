package ulog

import (
	"encoding/json"
)

var (
	defaultJsonKeys   [numFields]string
	defaultJsonLevels [numLevels]string
)

var JSONFormatter = newJsonFormatter()

func init() {
	defaultJsonKeys[LevelField] = "level"
	defaultJsonKeys[MessageField] = "message"
	defaultJsonKeys[TimeField] = "time"

	defaultJsonLevels[TraceLevel] = "trace"
	defaultJsonLevels[DebugLevel] = "debug"
	defaultJsonLevels[InfoLevel] = "info"
	defaultJsonLevels[WarnLevel] = "warning"
	defaultJsonLevels[ErrorLevel] = "error"
	defaultJsonLevels[FatalLevel] = "fatal"
}

func newJsonFormatter() *jsonfmt {
	k := [numFields]string{}
	copy(k[:], defaultJsonKeys[:])
	l := [numLevels]string{}
	copy(l[:], defaultJsonLevels[:])

	return &jsonfmt{
		keys:   k,
		levels: l,
	}
}

type jsonfmt struct {
	keys   [numFields]string
	levels [numLevels]string
}

// Format implements a Formatter that writes log entries as JSON.
func (w *jsonfmt) Format(id int, e entry, b ByteWriter) {
	entry := map[string]any{}
	entry[w.keys[TimeField]] = e.Time
	entry[w.keys[LevelField]] = w.levels[e.Level]
	entry[w.keys[MessageField]] = e.Message

	if e.fields != nil {
		for k, v := range e.fields.m {
			entry[k] = v
		}
	}
	enc := json.NewEncoder(b)
	_ = enc.Encode(entry)
}
