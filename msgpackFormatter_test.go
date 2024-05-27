package ulog

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/blugnu/msgpack"
	"github.com/blugnu/test"
)

func packedBytes(args ...any) []byte {
	buf := &bytes.Buffer{}
	for _, arg := range args {
		switch arg := arg.(type) {
		case string:
			buf.Write([]byte(arg))
		case int:
			buf.Write([]byte{byte(arg)})
		case byte:
			buf.Write([]byte{arg})
		case []byte:
			buf.Write([]byte(arg))
		}
	}
	return buf.Bytes()
}

func TestMsgpackFormatter(t *testing.T) {
	// ARRANGE
	testcases := []struct {
		scenario string
		exec     func(t *testing.T)
	}{
		{scenario: "MsgpackFormatter/with no options",
			exec: func(t *testing.T) {
				// ACT
				result, err := MsgpackFormatter()()

				// ASSERT
				test.Error(t, err).IsNil()
				if result, ok := test.IsType[*msgpackfmt](t, result); ok {
					test.That(t, result.keys).Equals([numFields][]byte{
						LevelField:            msgpack.EncodeString("level"),
						MessageField:          msgpack.EncodeString("message"),
						TimeField:             msgpack.EncodeString("timestamp"),
						CallsiteFileField:     msgpack.EncodeString("file"),
						CallsiteFunctionField: msgpack.EncodeString("function"),
					})
					test.That(t, result.levels).Equals([numLevels][]byte{
						TraceLevel: msgpack.EncodeString("trace"),
						DebugLevel: msgpack.EncodeString("debug"),
						InfoLevel:  msgpack.EncodeString("info"),
						WarnLevel:  msgpack.EncodeString("warning"),
						ErrorLevel: msgpack.EncodeString("error"),
						FatalLevel: msgpack.EncodeString("fatal"),
					})
				}
			},
		},
		{scenario: "MsgpackFormatter/option error",
			exec: func(t *testing.T) {
				// ARRANGE
				err := errors.New("option error")
				opt := func(*msgpackfmt) error { return err }

				// ACT
				result, got := MsgpackFormatter(opt)()

				// ASSERT
				test.Error(t, got).Is(err)
				test.That(t, result).IsNil()
			},
		},

		// Format tests
		{scenario: "Format",
			exec: func(t *testing.T) {
				// ARRANGE
				mx := &mockmutex{}
				buf := &bytes.Buffer{}
				e := entry{
					Time:    time.Date(2010, 9, 8, 7, 6, 5, 432100000, time.UTC),
					Level:   InfoLevel,
					Message: "entry",
				}
				sut := &msgpackfmt{}

				// we need a msgpack encoded timestampe to include in the expected
				// format output; we use a msgpack encoder to do this, using the
				// existing []byte buffer

				enc, _ := msgpack.NewEncoder(buf)
				_ = enc.EncodeTimestamp(e.Time)
				tsb := append([]byte{}, buf.Bytes()...)
				buf.Reset()

				testcases := []struct {
					scenario string
					exec     func(t *testing.T)
				}{
					{scenario: "no fields",
						exec: func(t *testing.T) {
							// ACT
							sut.Format(0, e, buf)

							// ASSERT
							IsSyncSafe(t, true, mx)
							test.That(t, buf.Bytes()).Equals(packedBytes(0x83, 0xa9, "timestamp", tsb, 0xa5, "level", 0xa4, "info", 0xa7, "message", 0xa5, "entry"))
						},
					},
					{scenario: "unformatted fields",
						exec: func(t *testing.T) {
							// ARRANGE
							e.logcontext = &logcontext{
								fields: &fields{
									mutex: mx,
									m:     map[string]any{"ikey": 123},
									b:     map[int][]byte{},
								},
							}

							// ACT
							sut.Format(0, e, buf)

							// ASSERT
							IsSyncSafe(t, false, mx)
							test.That(t, buf.Bytes()).Equals(packedBytes(0x84, 0xa9, "timestamp", tsb, 0xa5, "level", 0xa4, "info", 0xa7, "message", 0xa5, "entry", 0xa4, "ikey", 123))
						},
					},
					{scenario: "struct field",
						exec: func(t *testing.T) {
							// ARRANGE
							e.logcontext = &logcontext{
								fields: &fields{
									mutex: mx,
									m:     map[string]any{"struct": struct{ Field string }{"value"}},
									b:     map[int][]byte{},
								},
							}

							// ACT
							sut.Format(0, e, buf)

							// ASSERT
							IsSyncSafe(t, false, mx)
							test.That(t, buf.String()).Equals(string(packedBytes(0x84, 0xa9, "timestamp", tsb, 0xa5, "level", 0xa4, "info", 0xa7, "message", 0xa5, "entry", 0xa6, "struct", 0x81, 0xa5, "field", 0xa5, "value")))
						},
					},
					{scenario: "*struct field",
						exec: func(t *testing.T) {
							// ARRANGE
							e.logcontext = &logcontext{
								fields: &fields{
									mutex: mx,
									m:     map[string]any{"struct": &struct{ Field string }{"value"}},
									b:     map[int][]byte{},
								},
							}

							// ACT
							sut.Format(0, e, buf)

							// ASSERT
							IsSyncSafe(t, false, mx)
							test.That(t, buf.String()).Equals(string(packedBytes(0x84, 0xa9, "timestamp", tsb, 0xa5, "level", 0xa4, "info", 0xa7, "message", 0xa5, "entry", 0xa6, "struct", 0x81, 0xa5, "field", 0xa5, "value")))
						},
					},
					{scenario: "nested structs",
						exec: func(t *testing.T) {
							// ARRANGE
							e.logcontext = &logcontext{
								fields: &fields{
									mutex: mx,
									m: map[string]any{
										"outer": struct {
											Inner struct {
												Field string
											}
										}{Inner: struct{ Field string }{"value"}},
									},
									b: map[int][]byte{},
								},
							}

							// ACT
							sut.Format(0, e, buf)

							// ASSERT
							IsSyncSafe(t, false, mx)
							test.That(t, buf.String()).Equals(string(packedBytes(0x84, 0xa9, "timestamp", tsb, 0xa5, "level", 0xa4, "info", 0xa7, "message", 0xa5, "entry", 0xa5, "outer", 0x81, 0xa5, "inner", 0x81, 0xa5, "field", 0xa5, "value")))
						},
					},
					{scenario: "struct field/marshalling error",
						exec: func(t *testing.T) {
							// ARRANGE
							e.logcontext = &logcontext{
								fields: &fields{
									mutex: mx,
									m:     map[string]any{"struct": struct{}{}},
									b:     map[int][]byte{},
								},
							}
							defer test.Using(&jsonMarshal, func(v any) ([]byte, error) { return nil, errors.New("\"marshalling\" error") })()

							// ACT
							sut.Format(0, e, buf)

							// ASSERT
							IsSyncSafe(t, false, mx)
							test.That(t, buf.String()).Equals(string(packedBytes(0x84, 0xa9, "timestamp", tsb, 0xa5, "level", 0xa4, "info", 0xa7, "message", 0xa5, "entry", 0xa6, "struct", 0xd9, 0x34 /* 52 chars */, "LOGFMT_ERROR: marshalling error: \"marshalling\" error")))
						},
					},
					{scenario: "cached formatted fields",
						exec: func(t *testing.T) {
							// ARRANGE
							e.logcontext = &logcontext{
								fields: &fields{
									mutex: mx,
									m:     map[string]any{"key": "value"},
									b: map[int][]byte{
										0: packedBytes(0xa3, "key", 0xa5, "value"),
									},
								},
							}

							// ACT
							sut.Format(0, e, buf)

							// ASSERT
							IsSyncSafe(t, false, mx)
							test.That(t, buf.Bytes()).Equals(packedBytes(0x84, 0xa9, "timestamp", tsb, 0xa5, "level", 0xa4, "info", 0xa7, "message", 0xa5, "entry", 0xa3, "key", 0xa5, "value"))
						},
					},
				}
				for _, tc := range testcases {
					t.Run(tc.scenario, func(t *testing.T) {
						// ARRANGE
						sut.init()
						mx.Reset()
						buf.Reset()
						e.logcontext = &logcontext{}

						// ACT
						tc.exec(t)
					})
				}
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.exec(t)
		})
	}
}
