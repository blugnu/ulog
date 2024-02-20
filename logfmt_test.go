package ulog

import (
	"bytes"
	"testing"
	"time"

	"github.com/blugnu/test"
)

func TestLogfmtFormatter(t *testing.T) {
	// ARRANGE
	mx := &mockmutex{}
	tm := time.Date(2010, 9, 8, 7, 6, 5, 432100000, time.UTC)
	buf := &bytes.Buffer{}
	f, _ := LogfmtFormatter()()
	sut := f.(*logfmt)

	testcases := []struct {
		scenario string
		exec     func(t *testing.T)
	}{
		{scenario: "no fields, no callsite",
			exec: func(t *testing.T) {
				// ACT
				sut.Format(0, entry{
					logcontext: &logcontext{},
					Time:       tm,
					Level:      InfoLevel,
					Message:    "message",
				}, buf)

				// ASSERT
				IsSyncSafe(t, true, mx)
				test.That(t, buf.String()).Equals("time=2010-09-08T07:06:05.432100Z level=INFO  message=\"message\"")
			},
		},

		// callsite formatting
		{scenario: "no fields/with callsite/line 0-9",
			exec: func(t *testing.T) {
				// ACT
				sut.Format(0, entry{
					logcontext: &logcontext{},
					callsite:   &callsite{function: "function", file: "/path/to/file.go", line: 1},
					Time:       tm,
					Level:      InfoLevel,
					Message:    "message",
				}, buf)

				// ASSERT
				IsSyncSafe(t, true, mx)
				test.That(t, buf.String()).Equals(`time=2010-09-08T07:06:05.432100Z level=INFO  message="message" function="function" file="/path/to/file.go:1"`)
			},
		},
		{scenario: "no fields/with callsite/line 10-99",
			exec: func(t *testing.T) {
				// ACT
				sut.Format(0, entry{
					logcontext: &logcontext{},
					callsite:   &callsite{function: "function", file: "/path/to/file.go", line: 12},
					Time:       tm,
					Level:      InfoLevel,
					Message:    "message",
				}, buf)

				// ASSERT
				IsSyncSafe(t, true, mx)
				test.That(t, buf.String()).Equals(`time=2010-09-08T07:06:05.432100Z level=INFO  message="message" function="function" file="/path/to/file.go:12"`)
			},
		},
		{scenario: "no fields/with callsite/line 100-999",
			exec: func(t *testing.T) {
				// ACT
				sut.Format(0, entry{
					logcontext: &logcontext{},
					callsite:   &callsite{function: "function", file: "/path/to/file.go", line: 123},
					Time:       tm,
					Level:      InfoLevel,
					Message:    "message",
				}, buf)

				// ASSERT
				IsSyncSafe(t, true, mx)
				test.That(t, buf.String()).Equals(`time=2010-09-08T07:06:05.432100Z level=INFO  message="message" function="function" file="/path/to/file.go:123"`)
			},
		},
		{scenario: "no fields/with callsite/line 1000-9999",
			exec: func(t *testing.T) {
				// ACT
				sut.Format(0, entry{
					logcontext: &logcontext{},
					callsite:   &callsite{function: "function", file: "/path/to/file.go", line: 1234},
					Time:       tm,
					Level:      InfoLevel,
					Message:    "message",
				}, buf)

				// ASSERT
				IsSyncSafe(t, true, mx)
				test.That(t, buf.String()).Equals(`time=2010-09-08T07:06:05.432100Z level=INFO  message="message" function="function" file="/path/to/file.go:1234"`)
			},
		},
		{scenario: "no fields/with callsite/line 10000+",
			exec: func(t *testing.T) {
				// ACT
				sut.Format(0, entry{
					logcontext: &logcontext{},
					callsite:   &callsite{function: "function", file: "/path/to/file.go", line: 12345},
					Time:       tm,
					Level:      InfoLevel,
					Message:    "message",
				}, buf)

				// ASSERT
				IsSyncSafe(t, true, mx)
				test.That(t, buf.String()).Equals(`time=2010-09-08T07:06:05.432100Z level=INFO  message="message" function="function" file="/path/to/file.go:12345"`)
			},
		},

		// int field formatting
		{scenario: "unformatted int field/0-9/0",
			exec: func(t *testing.T) {
				// ACT
				sut.Format(0, entry{
					logcontext: &logcontext{
						fields: &fields{
							mutex: mx,
							m:     map[string]any{"ikey": 0},
							b:     map[int][]byte{},
						},
					},
					Time:    tm,
					Level:   InfoLevel,
					Message: "message",
				}, buf)

				// ASSERT
				test.That(t, buf.String()).Equals("time=2010-09-08T07:06:05.432100Z level=INFO  message=\"message\" ikey=0")
				IsSyncSafe(t, false, mx)
			},
		},
		{scenario: "unformatted int field/0-9/9",
			exec: func(t *testing.T) {
				// ACT
				sut.Format(0, entry{
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
				}, buf)

				// ASSERT
				test.That(t, buf.String()).Equals("time=2010-09-08T07:06:05.432100Z level=INFO  message=\"message\" ikey=9")
				IsSyncSafe(t, false, mx)
			},
		},
		{scenario: "unformatted int field/10-99/10",
			exec: func(t *testing.T) {
				// ACT
				sut.Format(0, entry{
					logcontext: &logcontext{
						fields: &fields{
							mutex: mx,
							m:     map[string]any{"ikey": 10},
							b:     map[int][]byte{},
						},
					},
					Time:    tm,
					Level:   InfoLevel,
					Message: "message",
				}, buf)

				// ASSERT
				test.That(t, buf.String()).Equals("time=2010-09-08T07:06:05.432100Z level=INFO  message=\"message\" ikey=10")
				IsSyncSafe(t, false, mx)
			},
		},
		{scenario: "unformatted int field/10-99/99",
			exec: func(t *testing.T) {
				// ACT
				sut.Format(0, entry{
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
				}, buf)

				// ASSERT
				test.That(t, buf.String()).Equals("time=2010-09-08T07:06:05.432100Z level=INFO  message=\"message\" ikey=99")
				IsSyncSafe(t, false, mx)
			},
		},
		{scenario: "unformatted int field/100-999/100",
			exec: func(t *testing.T) {
				// ACT
				sut.Format(0, entry{
					logcontext: &logcontext{
						fields: &fields{
							mutex: mx,
							m:     map[string]any{"ikey": 100},
							b:     map[int][]byte{},
						},
					},
					Time:    tm,
					Level:   InfoLevel,
					Message: "message",
				}, buf)

				// ASSERT
				test.That(t, buf.String()).Equals("time=2010-09-08T07:06:05.432100Z level=INFO  message=\"message\" ikey=100")
				IsSyncSafe(t, false, mx)
			},
		},
		{scenario: "unformatted int field/100-999/999",
			exec: func(t *testing.T) {
				// ACT
				sut.Format(0, entry{
					logcontext: &logcontext{
						fields: &fields{
							mutex: mx,
							m:     map[string]any{"ikey": 999},
							b:     map[int][]byte{},
						},
					},
					Time:    tm,
					Level:   InfoLevel,
					Message: "message",
				}, buf)

				// ASSERT
				test.That(t, buf.String()).Equals("time=2010-09-08T07:06:05.432100Z level=INFO  message=\"message\" ikey=999")
				IsSyncSafe(t, false, mx)
			},
		},
		{scenario: "unformatted int field/1000-9999/1000",
			exec: func(t *testing.T) {
				// ACT
				sut.Format(0, entry{
					logcontext: &logcontext{
						fields: &fields{
							mutex: mx,
							m:     map[string]any{"ikey": 1000},
							b:     map[int][]byte{},
						},
					},
					Time:    tm,
					Level:   InfoLevel,
					Message: "message",
				}, buf)

				// ASSERT
				test.That(t, buf.String()).Equals("time=2010-09-08T07:06:05.432100Z level=INFO  message=\"message\" ikey=1000")
				IsSyncSafe(t, false, mx)
			},
		},
		{scenario: "unformatted int field/1000-9999/9999",
			exec: func(t *testing.T) {
				// ACT
				sut.Format(0, entry{
					logcontext: &logcontext{
						fields: &fields{
							mutex: mx,
							m:     map[string]any{"ikey": 9999},
							b:     map[int][]byte{},
						},
					},
					Time:    tm,
					Level:   InfoLevel,
					Message: "message",
				}, buf)

				// ASSERT
				test.That(t, buf.String()).Equals("time=2010-09-08T07:06:05.432100Z level=INFO  message=\"message\" ikey=9999")
				IsSyncSafe(t, false, mx)
			},
		},
		{scenario: "unformatted int field/10000+/12345",
			exec: func(t *testing.T) {
				// ACT
				sut.Format(0, entry{
					logcontext: &logcontext{
						fields: &fields{
							mutex: mx,
							m:     map[string]any{"ikey": 12345},
							b:     map[int][]byte{},
						},
					},
					Time:    tm,
					Level:   InfoLevel,
					Message: "message",
				}, buf)

				// ASSERT
				test.That(t, buf.String()).Equals("time=2010-09-08T07:06:05.432100Z level=INFO  message=\"message\" ikey=12345")
				IsSyncSafe(t, false, mx)
			},
		},
		{scenario: "unformatted int field/<0/-99",
			exec: func(t *testing.T) {
				// ACT
				sut.Format(0, entry{
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
				}, buf)

				// ASSERT
				test.That(t, buf.String()).Equals("time=2010-09-08T07:06:05.432100Z level=INFO  message=\"message\" ikey=-99")
				IsSyncSafe(t, false, mx)
			},
		},
		{scenario: "unformatted string field",
			exec: func(t *testing.T) {
				// ACT
				sut.Format(0, entry{
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
				}, buf)

				// ASSERT
				test.That(t, buf.String()).Equals("time=2010-09-08T07:06:05.432100Z level=INFO  message=\"message\" key=\"value\"")
				IsSyncSafe(t, false, mx)
			},
		},
		{scenario: "unformatted bool field",
			exec: func(t *testing.T) {
				// ACT
				sut.Format(0, entry{
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
				}, buf)

				// ASSERT
				test.That(t, buf.String()).Equals("time=2010-09-08T07:06:05.432100Z level=INFO  message=\"message\" key=true")
				IsSyncSafe(t, false, mx)
			},
		},
		{scenario: "cached formatted fields",
			exec: func(t *testing.T) {
				// ARRANGE
				buf.Reset()
				mx.Reset()

				// ACT
				sut.Format(0, entry{
					logcontext: &logcontext{
						fields: &fields{
							mutex: mx,
							b: map[int][]byte{
								0: []byte(" key=\"cached value\""),
							},
						},
					},
					Time:    tm,
					Level:   InfoLevel,
					Message: "message",
				}, buf)

				// ASSERT
				test.That(t, buf.String()).Equals("time=2010-09-08T07:06:05.432100Z level=INFO  message=\"message\" key=\"cached value\"")
				IsSyncSafe(t, false, mx)
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			// ARRANGE
			buf.Reset()
			mx.Reset()

			// ACT & ASSERT
			tc.exec(t)
		})
	}
}
