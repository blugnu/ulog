package ulog

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

// function variables for testing
var (
	jsonMarshal = json.Marshal
)

type LogfmtOption func(*logfmt) error // LogfmtOption is a function that configures a logfmt formatter

// LogfmtFormatter configures a logfmt formatter to format entries in the logfmt format:
//
// The default formatter will produce output similar to:
//
//	time=2006-01-02T15:04:05.000000Z level=INFO  message="some logged message"
//
// Configuration options are provided to allow the default labels and values for the level
// field:
//
// LogfmtLabels() may be used to configure the labels used for each of the core fields
// in an entry: TimestampField, LevelField, MessageField.  The order of the fields
// is fixed and cannot be changed.
//
// LogfmtLevels() may be used to configure the values used for the LevelField value.
func LogfmtFormatter(opt ...LogfmtOption) func() (Formatter, error) {
	return func() (Formatter, error) {
		lf := &logfmt{
			keys: [numFields][]byte{
				[]byte("time="),
				[]byte(" level="),
				[]byte(" message=\""),
				[]byte(" file=\""),
				[]byte(" function=\""),
			},
			levels: [numLevels][]byte{
				{},
				[]byte("FATAL"),
				[]byte("ERROR"),
				[]byte("WARN "),
				[]byte("INFO "),
				[]byte("DEBUG"),
				[]byte("TRACE"),
			},
		}

		for _, o := range opt {
			if err := o(lf); err != nil {
				return nil, err
			}
		}
		return lf, nil
	}
}

// logfmt is a formatter that formats entries as logfmt.
type logfmt struct {
	keys   [numFields][]byte
	levels [numLevels][]byte
}

// Format implements the Formatter interface to Format log entries
// in the logfmt Format.
func (w *logfmt) Format(id int, e entry, b ByteWriter) {
	utc := e.Time
	y := utc.Year()
	ns := utc.Nanosecond()
	_, _ = b.Write(w.keys[TimeField])
	_, _ = b.Write(buf.digits2[y/100])
	_, _ = b.Write(buf.digits2[y%100])
	_ = b.WriteByte(char.hyphen)
	_, _ = b.Write(buf.digits2[int(utc.Month())])
	_ = b.WriteByte(char.hyphen)
	_, _ = b.Write(buf.digits2[utc.Day()])
	_ = b.WriteByte(char.T)
	_, _ = b.Write(buf.digits2[utc.Hour()])
	_ = b.WriteByte(char.colon)
	_, _ = b.Write(buf.digits2[utc.Minute()])
	_ = b.WriteByte(char.colon)
	_, _ = b.Write(buf.digits2[utc.Second()])
	_ = b.WriteByte(char.period)
	_, _ = b.Write(buf.digits2[ns/10000000])     // 123456789 / 10000000 = 12
	_, _ = b.Write(buf.digits2[(ns/100000)%100]) // 123456789 / 100000   = 1234 % 100 = 34
	_, _ = b.Write(buf.digits2[(ns/1000)%100])   // 123456789 / 1000     = 123456 % 100 = 56
	_ = b.WriteByte(char.Z)
	_, _ = b.Write(w.keys[LevelField])
	_, _ = b.Write(w.levels[e.Level])
	_, _ = b.Write(w.keys[MessageField])
	_, _ = b.Write([]byte(e.Message))
	_ = b.WriteByte(char.quote)

	if e.callsite != nil {
		_, _ = b.Write(w.keys[CallsiteFunctionField])
		_, _ = b.Write([]byte(e.callsite.function))
		_ = b.WriteByte(char.quote)
		_, _ = b.Write(w.keys[CallsiteFileField])
		_, _ = b.Write([]byte(e.callsite.file))
		_ = b.WriteByte(char.colon)
		w.writeInt(b, e.callsite.line)
		_ = b.WriteByte(char.quote)
	}

	fbb := e.fields.getFormattedBytes(id)
	if fbb == nil {
		return
	}

	if fbb.Len() > 0 {
		_, _ = b.Write(fbb.Bytes())
		return
	}

	w.writeFields(fbb, e.fields)

	fb := make([]byte, fbb.Len())
	copy(fb, fbb.Bytes())

	e.fields.setFormattedBytes(id, fb)

	_, _ = b.Write(fb)
}

func (w *logfmt) writeFields(buf ByteWriter, f *fields) {
	for k, v := range f.m {
		if reflect.ValueOf(v).Kind() == reflect.Struct ||
			(reflect.ValueOf(v).Kind() == reflect.Ptr && reflect.ValueOf(v).Elem().Kind() == reflect.Struct) {
			w.writeStruct(buf, k, v)
			continue
		}

		_ = buf.WriteByte(char.space)
		_, _ = buf.Write([]byte(k))
		_ = buf.WriteByte(char.equal)
		switch v := v.(type) {
		case int:
			w.writeInt(buf, v)

		case string:
			_ = buf.WriteByte(char.quote)
			_, _ = buf.Write([]byte(v))
			_ = buf.WriteByte(char.quote)

		default:
			_, _ = buf.Write([]byte(fmt.Sprintf("%v", v)))
		}
	}
}

func (w *logfmt) writeInt(b ByteWriter, i int) {
	if i < 0 {
		_ = b.WriteByte(char.hyphen)
		i = i ^ -1 + 1
	}

	switch {
	case i >= 0 && i <= 9:
		_ = b.WriteByte(char.digit[i])
	case i >= 10 && i <= 99:
		_, _ = b.Write(buf.digits2[i])
	case i >= 100 && i <= 999:
		_ = b.WriteByte(char.digit[i/100])
		_, _ = b.Write(buf.digits2[i%100])
	case i >= 1000 && i <= 9999:
		_, _ = b.Write(buf.digits2[i/100])
		_, _ = b.Write(buf.digits2[i%100])
	default:
		_, _ = b.Write([]byte(strconv.Itoa(i)))
	}
}

func (w *logfmt) writeMap(buf ByteWriter, k string, m map[string]any) {
	keys := make([]string, 0, len(m))
	for mk := range m {
		keys = append(keys, mk)
	}
	slices.Sort(keys)

	for _, mk := range keys {
		mv := m[mk]
		mk = strings.ToLower(mk)
		if m, ok := mv.(map[string]any); ok {
			w.writeMap(buf, k+"."+mk, m)
			continue
		}

		_ = buf.WriteByte(char.space)
		_, _ = buf.Write([]byte(k))
		_ = buf.WriteByte(char.period)
		_, _ = buf.Write([]byte(mk))
		_ = buf.WriteByte(char.equal)

		switch v := mv.(type) {
		case bool:
			if v {
				_, _ = buf.Write([]byte("true"))
			} else {
				_, _ = buf.Write([]byte("false"))
			}

		case string:
			_ = buf.WriteByte(char.quote)
			_, _ = buf.Write([]byte(v))
			_ = buf.WriteByte(char.quote)

		default:
			_, _ = buf.Write([]byte(fmt.Sprintf("%v", v)))
		}
	}
}

func (w *logfmt) writeStruct(buf ByteWriter, k string, v any) bool {
	j, err := jsonMarshal(v)
	if err != nil {
		_ = buf.WriteByte(char.space)
		_, _ = buf.Write([]byte(k))
		_ = buf.WriteByte(char.equal)
		_, _ = buf.Write([]byte(fmt.Sprintf("%q", fmt.Sprintf("LOGFMT_ERROR: error marshalling struct field: %v", err))))
		return false
	}

	// unmarshal the json into a map (no need to check for errors; we are
	// unmarshalling marshalled JSON, it cannot be invalid)
	m := map[string]any{}
	_ = json.Unmarshal(j, &m)

	w.writeMap(buf, k, m)
	return true
}
