package ulog

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/blugnu/msgpack"
)

type MsgpackOption func(*msgpackfmt) error // MsgpackOption is a function for configuring a msgpack formatter

// MsgpackFormatter returns a function that configures a msgpack formatter.
func MsgpackFormatter(opts ...MsgpackOption) FormatterFactory {
	return func() (Formatter, error) {
		mf := &msgpackfmt{}
		mf.init()

		errs := []error{}
		for _, opt := range opts {
			errs = append(errs, opt(mf))
		}
		if err := errors.Join(errs...); err != nil {
			return nil, err
		}

		return mf, nil
	}
}

func (f *msgpackfmt) init() {
	f.keys = [numFields][]byte{
		LevelField:            msgpack.EncodeString("level"),
		MessageField:          msgpack.EncodeString("message"),
		TimeField:             msgpack.EncodeString("timestamp"),
		CallsiteFileField:     msgpack.EncodeString("file"),
		CallsiteFunctionField: msgpack.EncodeString("function"),
	}
	f.levels = [numLevels][]byte{
		TraceLevel: msgpack.EncodeString("trace"),
		DebugLevel: msgpack.EncodeString("debug"),
		InfoLevel:  msgpack.EncodeString("info"),
		WarnLevel:  msgpack.EncodeString("warning"),
		ErrorLevel: msgpack.EncodeString("error"),
		FatalLevel: msgpack.EncodeString("fatal"),
	}
	f.enc = &sync.Pool{New: func() any {
		enc, _ := msgpack.NewEncoder(nil)
		return enc
	}}

}

type msgpackfmt struct {
	keys   [numFields][]byte // slice indexed by ord(Key) to pre-packed byte slices
	levels [numLevels][]byte // slice indexed by level of pre-packed byte slices
	enc    *sync.Pool        // msgpack encoder
}

func (fmt *msgpackfmt) Format(id int, e entry, b ByteWriter) {
	// writes a log entry as a msgpack map
	//
	// the map has n + 3 elements, where n is the number of fields
	// in the entry, the 3 fixed elements being time, level, and message
	//
	// the time element is a string in ISO-18601 format
	enc := fmt.enc.Get().(*msgpack.Encoder)
	defer fmt.enc.Put(enc)

	fn := 3
	if e.fields != nil {
		fn += len(e.fields.m)
	}
	enc.SetWriter(b)
	_ = enc.WriteMapHeader(fn)

	_ = enc.Write(fmt.keys[TimeField])
	_ = enc.EncodeTimestamp(e.Time)

	_ = enc.Write(fmt.keys[LevelField])
	_ = enc.Write(fmt.levels[e.Level])

	_ = enc.Write(fmt.keys[MessageField])
	_ = enc.EncodeString(e.Message)

	fbb := e.fields.getFormattedBytes(id)
	if fbb == nil { // nil => no fields
		return
	}

	if fbb.Len() > 0 { // cached msgpack encoding retrieved
		_ = enc.Write(fbb.Bytes())
		return
	}

	// encode the fields
	_ = enc.Using(fbb, func() error {
		for k, v := range e.fields.m {
			_ = enc.EncodeString(k)

			if err, ok := v.(error); ok {
				_ = enc.EncodeString(err.Error())
				continue
			}

			if reflect.ValueOf(v).Kind() == reflect.Struct ||
				(reflect.ValueOf(v).Kind() == reflect.Ptr && reflect.ValueOf(v).Elem().Kind() == reflect.Struct) {
				fmt.writeStruct(enc, v)
				continue
			}
			_ = enc.Encode(v)
		}
		return enc.Err()
	})
	fb := fbb.Bytes()
	e.fields.setFormattedBytes(id, fb)

	_ = enc.Write(fb)
}

func (fmt *msgpackfmt) writeMap(enc *msgpack.Encoder, m map[string]any) {
	_ = enc.WriteMapHeader(len(m))
	for k, v := range m {
		_ = enc.EncodeString(strings.ToLower(k))
		if m, ok := v.(map[string]any); ok {
			fmt.writeMap(enc, m)
			continue
		}
		_ = enc.Encode(v)

	}
}

func (msg *msgpackfmt) writeStruct(enc *msgpack.Encoder, v any) {
	j, err := jsonMarshal(v)
	if err != nil {
		_ = enc.EncodeString(fmt.Sprintf("LOGFMT_ERROR: marshalling error: %v", err))
		return
	}

	// unmarshal the json into a map (no need to check for errors; we are
	// unmarshalling marshalled JSON, it cannot be invalid)
	var m map[string]any
	_ = json.Unmarshal(j, &m)

	msg.writeMap(enc, m)
}
