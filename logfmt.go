package ulog

import (
	"fmt"
	"strconv"
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
	writeint := func(b ByteWriter, i int) {
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
		_, _ = b.Write(w.keys[FunctionField])
		_, _ = b.Write([]byte(e.callsite.function))
		_ = b.WriteByte(char.quote)
		_, _ = b.Write(w.keys[FileField])
		_, _ = b.Write([]byte(e.callsite.file))
		_ = b.WriteByte(char.colon)
		writeint(b, e.callsite.line)
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

	for k, v := range e.fields.m {
		_ = fbb.WriteByte(char.space)
		_, _ = fbb.Write([]byte(k))
		_ = fbb.WriteByte(char.equal)

		switch v := v.(type) {
		case int:
			writeint(fbb, v)
		case string:
			_ = fbb.WriteByte(char.quote)
			_, _ = fbb.Write([]byte(v))
			_ = fbb.WriteByte(char.quote)
		default:
			_, _ = fbb.Write([]byte(fmt.Sprintf("%v", v)))
		}
	}
	fb := make([]byte, fbb.Len())
	copy(fb, fbb.Bytes())

	e.fields.setFormattedBytes(id, fb)

	_, _ = b.Write(fb)
}
