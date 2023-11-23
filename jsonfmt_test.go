package ulog

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/blugnu/test"
)

func TestJsonFormatter(t *testing.T) {
	// ARRANGE
	mx := &mockmutex{}
	tm := time.Date(2010, 9, 8, 7, 6, 5, 432100000, time.UTC)
	buf := &bytes.Buffer{}
	f, _ := NewJSONFormatter()()
	sut := f.(*jsonfmt)

	testcases := []struct {
		name   string
		entry  entry
		result []byte
	}{
		{name: "no fields",
			entry: entry{
				logcontext: &logcontext{},
				Time:       tm,
				Level:      InfoLevel,
				Message:    "message",
			},
			result: []byte("{\"time\":\"2010-09-08T07:06:05.4321Z\",\"level\":\"info\",\"message\":\"message\"}")},
		{name: "with int field",
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
			result: []byte("{\"time\":\"2010-09-08T07:06:05.4321Z\",\"level\":\"info\",\"message\":\"message\",\"ikey\":99}")},
		{name: "with string field",
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
			result: []byte("{\"time\":\"2010-09-08T07:06:05.4321Z\",\"level\":\"info\",\"message\":\"message\",\"key\":\"value\"}")},
		{name: "with bool field",
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
			result: []byte("{\"time\":\"2010-09-08T07:06:05.4321Z\",\"level\":\"info\",\"message\":\"message\",\"key\":true}")},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// ARRANGE
			defer mx.Reset()
			defer buf.Reset()

			// ACT
			sut.Format(0, tc.entry, buf)

			// ASSERT
			IsSyncSafe(t, true, mx) // the current implementation of the json formatter is synsafe, the mutex should not be acquired

			wanted := map[string]any{}
			got := map[string]any{}
			test.UnexpectedError(t, json.Unmarshal(tc.result, &wanted))
			test.UnexpectedError(t, json.Unmarshal(buf.Bytes(), &got))
			test.Maps(t, wanted, got)
		})
	}
}
