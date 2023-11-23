package ulog

import (
	"bytes"
	"testing"
	"time"
)

func TestLogfmtFormatter(t *testing.T) {
	// ARRANGE
	mx := &mockmutex{}
	tm := time.Date(2010, 9, 8, 7, 6, 5, 432100000, time.UTC)
	buf := &bytes.Buffer{}
	f, _ := Logfmt()()
	sut := f.(*logfmt)

	testcases := []struct {
		name     string
		syncsafe bool
		entry    entry
		result   []byte
	}{
		{name: "no fields, no callsite", syncsafe: true,
			entry: entry{
				logcontext: &logcontext{},
				Time:       tm,
				Level:      InfoLevel,
				Message:    "message",
			},
			result: []byte("time=2010-09-08T07:06:05.432100Z level=INFO  message=\"message\"")},
		{name: "no fields, with callsite, line 0-9", syncsafe: true,
			entry: entry{
				logcontext: &logcontext{},
				callsite:   &callsite{function: "function", file: "/path/to/file.go", line: 1},
				Time:       tm,
				Level:      InfoLevel,
				Message:    "message",
			},
			result: []byte(`time=2010-09-08T07:06:05.432100Z level=INFO  message="message" function="function" file="/path/to/file.go:1"`)},
		{name: "no fields, with callsite, line 10-99", syncsafe: true,
			entry: entry{
				logcontext: &logcontext{},
				callsite:   &callsite{function: "function", file: "/path/to/file.go", line: 12},
				Time:       tm,
				Level:      InfoLevel,
				Message:    "message",
			},
			result: []byte(`time=2010-09-08T07:06:05.432100Z level=INFO  message="message" function="function" file="/path/to/file.go:12"`)},
		{name: "no fields, with callsite, line 100-999", syncsafe: true,
			entry: entry{
				logcontext: &logcontext{},
				callsite:   &callsite{function: "function", file: "/path/to/file.go", line: 123},
				Time:       tm,
				Level:      InfoLevel,
				Message:    "message",
			},
			result: []byte(`time=2010-09-08T07:06:05.432100Z level=INFO  message="message" function="function" file="/path/to/file.go:123"`)},
		{name: "no fields, with callsite, line 1000-9999", syncsafe: true,
			entry: entry{
				logcontext: &logcontext{},
				callsite:   &callsite{function: "function", file: "/path/to/file.go", line: 1234},
				Time:       tm,
				Level:      InfoLevel,
				Message:    "message",
			},
			result: []byte(`time=2010-09-08T07:06:05.432100Z level=INFO  message="message" function="function" file="/path/to/file.go:1234"`)},
		{name: "no fields, with callsite, line 10000+", syncsafe: true,
			entry: entry{
				logcontext: &logcontext{},
				callsite:   &callsite{function: "function", file: "/path/to/file.go", line: 12345},
				Time:       tm,
				Level:      InfoLevel,
				Message:    "message",
			},
			result: []byte(`time=2010-09-08T07:06:05.432100Z level=INFO  message="message" function="function" file="/path/to/file.go:12345"`)},
		{name: "unformatted int field (0 <= 99 < 100)",
			entry: entry{
				logcontext: &logcontext{
					fields: &fields{
						mutex: mx,
						m:     map[string]any{"ikey": 99},
						b:     map[int][]byte{},
					},
				},
				Time:    tm,
				Level:   InfoLevel,
				Message: "message",
			},
			result: []byte("time=2010-09-08T07:06:05.432100Z level=INFO  message=\"message\" ikey=99")},
		{name: "unformatted int field (0 <= 19 < 100)",
			entry: entry{
				logcontext: &logcontext{
					fields: &fields{
						mutex: mx,
						m:     map[string]any{"ikey": 19},
						b:     map[int][]byte{},
					},
				},
				Time:    tm,
				Level:   InfoLevel,
				Message: "message",
			},
			result: []byte("time=2010-09-08T07:06:05.432100Z level=INFO  message=\"message\" ikey=19")},
		{name: "unformatted int field (0 <= 9 < 100)",
			entry: entry{
				logcontext: &logcontext{
					fields: &fields{
						mutex: mx,
						m:     map[string]any{"ikey": 9},
						b:     map[int][]byte{},
					},
				},
				Time:    tm,
				Level:   InfoLevel,
				Message: "message",
			},
			result: []byte("time=2010-09-08T07:06:05.432100Z level=INFO  message=\"message\" ikey=9")},
		{name: "unformatted int field (-99 <= -99 < 0)",
			entry: entry{
				logcontext: &logcontext{
					fields: &fields{
						mutex: mx,
						m:     map[string]any{"ikey": -99},
						b:     map[int][]byte{},
					},
				},
				Time:    tm,
				Level:   InfoLevel,
				Message: "message",
			},
			result: []byte("time=2010-09-08T07:06:05.432100Z level=INFO  message=\"message\" ikey=-99")},
		{name: "unformatted int field (-99 <= -11 < 0)",
			entry: entry{
				logcontext: &logcontext{
					fields: &fields{
						mutex: mx,
						m:     map[string]any{"ikey": -11},
						b:     map[int][]byte{},
					},
				},
				Time:    tm,
				Level:   InfoLevel,
				Message: "message",
			},
			result: []byte("time=2010-09-08T07:06:05.432100Z level=INFO  message=\"message\" ikey=-11")},
		{name: "unformatted int field (-99 <= -1 < 0)",
			entry: entry{
				logcontext: &logcontext{
					fields: &fields{
						mutex: mx,
						m:     map[string]any{"ikey": int(-1)},
						b:     map[int][]byte{},
					},
				},
				Time:    tm,
				Level:   InfoLevel,
				Message: "message",
			},
			result: []byte("time=2010-09-08T07:06:05.432100Z level=INFO  message=\"message\" ikey=-1")},
		{name: "unformatted int field (>99)",
			entry: entry{
				logcontext: &logcontext{
					fields: &fields{
						mutex: mx,
						m:     map[string]any{"ikey": 123},
						b:     map[int][]byte{},
					},
				},
				Time:    tm,
				Level:   InfoLevel,
				Message: "message",
			},
			result: []byte("time=2010-09-08T07:06:05.432100Z level=INFO  message=\"message\" ikey=123")},
		{name: "unformatted string field",
			entry: entry{
				logcontext: &logcontext{
					fields: &fields{
						mutex: mx,
						m:     map[string]any{"key": "value"},
						b:     map[int][]byte{},
					},
				},
				Time:    tm,
				Level:   InfoLevel,
				Message: "message",
			},
			result: []byte("time=2010-09-08T07:06:05.432100Z level=INFO  message=\"message\" key=\"value\"")},
		{name: "unformatted bool field",
			entry: entry{
				logcontext: &logcontext{
					fields: &fields{
						mutex: mx,
						m:     map[string]any{"key": true},
						b:     map[int][]byte{},
					},
				},
				Time:    tm,
				Level:   InfoLevel,
				Message: "message",
			},
			result: []byte("time=2010-09-08T07:06:05.432100Z level=INFO  message=\"message\" key=true")},
		{name: "cached formatted fields",
			entry: entry{
				logcontext: &logcontext{
					fields: &fields{
						mutex: mx,
						b: map[int][]byte{
							0: []byte(" key=\"value\""),
						},
					},
				},
				Time:    tm,
				Level:   InfoLevel,
				Message: "message",
			},
			result: []byte("time=2010-09-08T07:06:05.432100Z level=INFO  message=\"message\" key=\"value\"")},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// ARRANGE
			defer mx.Reset()
			defer buf.Reset()

			// ACT
			sut.Format(0, tc.entry, buf)

			// ASSERT
			IsSyncSafe(t, tc.syncsafe, mx)

			wanted := tc.result
			got := buf.Bytes()
			if !bytes.Equal(wanted, got) {
				t.Errorf("\nwanted %s\ngot    %s", wanted, got)
			}
		})
	}
}
