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
						LevelField:    msgpack.EncodeString("level"),
						MessageField:  msgpack.EncodeString("message"),
						TimeField:     msgpack.EncodeString("timestamp"),
						FileField:     msgpack.EncodeString("file"),
						FunctionField: msgpack.EncodeString("function"),
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

// func Test_msgpackfmt_Format(t *testing.T) {
// 	// ARRANGE

// 	mx := &mockmutex{}
// 	tm := time.Date(2010, 9, 8, 7, 6, 5, 432100000, time.UTC)
// 	buf := &bytes.Buffer{}
// 	sut := newMsgpackFormatter()

// 	// we need a msgpack encoded timestampe to include in the expected
// 	// format output; we use a msgpack encoder to do this, using the
// 	// existing []byte buffer
// 	//
// 	enc, _ := msgpack.NewEncoder(buf)
// 	_ = enc.EncodeTimestamp(tm)
// 	ts := append([]byte{}, buf.Bytes()...)
// 	buf.Reset()

// 	testcases := []struct {
// 		name     string
// 		syncsafe bool
// 		entry    entry
// 		result   []byte
// 	}{
// 		{name: "no fields", syncsafe: true,
// 			entry: entry{
// 				logcontext: &logcontext{},
// 				Time:       tm,
// 				Level:      InfoLevel,
// 				Message:    "message",
// 			},
// 			result: packedBytes(0x83, 0xa9, "timestamp", ts, 0xa5, "level", 0xa4, "info", 0xa7, "message", 0xa7, "message")},

// 		{name: "unformatted fields",
// 			entry: entry{
// 				logcontext: &logcontext{
// 					fields: &fields{
// 						mutex: mx,
// 						m:     map[string]any{"ikey": 123},
// 						b:     map[int][]byte{},
// 					},
// 				},
// 				Time:    tm,
// 				Level:   InfoLevel,
// 				Message: "message",
// 			},
// 			result: packedBytes(0x84, 0xa9, "timestamp", ts, 0xa5, "level", 0xa4, "info", 0xa7, "message", 0xa7, "message",
// 				0xa4, "ikey", 123)},
// 		{name: "cached formatted fields",
// 			entry: entry{
// 				logcontext: &logcontext{
// 					fields: &fields{
// 						mutex: mx,
// 						m:     map[string]any{"key": "value"},
// 						b: map[int][]byte{
// 							0: packedBytes(0xa3, "key", 0xa5, "value"),
// 						},
// 					},
// 				},
// 				Time:    tm,
// 				Level:   InfoLevel,
// 				Message: "message",
// 			},
// 			result: packedBytes(0x84, 0xa9, "timestamp", ts, 0xa5, "level", 0xa4, "info", 0xa7, "message", 0xa7, "message",
// 				0xa3, "key", 0xa5, "value"),
// 		},
// 	}
// 	for _, tc := range testcases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			// ARRANGE
// 			defer mx.Reset()
// 			defer buf.Reset()

// 			// ACT
// 			sut.Format(0, tc.entry, buf)

// 			// ASSERT
// 			IsSyncSafe(t, tc.syncsafe, mx)

// 			wanted := tc.result
// 			got := buf.Bytes()
// 			if !bytes.Equal(wanted, got) {
// 				t.Errorf("\nwanted %s\ngot    %s", wanted, got)
// 			}
// 		})
// 	}
// }
