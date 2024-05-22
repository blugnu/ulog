package ulog

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/blugnu/test"
)

func TestNewJsonFormatter(t *testing.T) {
	// ARRANGE
	def := &jsonfmt{
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

	testcases := []struct {
		scenario string
		exec     func(t *testing.T)
	}{
		{scenario: "with no options",
			exec: func(t *testing.T) {
				// ACT
				result, err := NewJSONFormatter()()

				// ASSERT
				test.That(t, result.(*jsonfmt)).Equals(def)
				test.Error(t, err).IsNil()
			},
		},
		{scenario: "with options",
			exec: func(t *testing.T) {
				// ARRANGE
				optWasApplied := false

				// ACT
				result, err := NewJSONFormatter(func(*jsonfmt) error { optWasApplied = true; return nil })()

				// ASSERT
				test.That(t, result.(*jsonfmt)).Equals(def)
				test.Error(t, err).IsNil()
				test.IsTrue(t, optWasApplied)
			},
		},
		{scenario: "with option errors",
			exec: func(t *testing.T) {
				// ARRANGE
				opterr := errors.New("option error")

				// ACT
				result, err := NewJSONFormatter(func(*jsonfmt) error { return opterr })()

				// ASSERT
				test.That(t, result).IsNil()
				test.Error(t, err).Is(opterr)
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.exec(t)
		})
	}
}

func TestJsonFormatter(t *testing.T) {
	// ARRANGE
	var (
		mx   = &mockmutex{}
		tm   = time.Date(2010, 9, 8, 7, 6, 5, 432100000, time.UTC)
		e    = entry{}
		sut  = JSONFormatter
		dest = bytes.NewBuffer([]byte{})
	)

	testcases := []struct {
		scenario string
		exec     func(t *testing.T)
	}{
		{scenario: "no fields",
			exec: func(t *testing.T) {
				// ARRANGE
				wanted := map[string]any{}
				got := map[string]any{}

				// ACT
				sut.Format(0, e, dest)

				// ASSERT
				_ = json.Unmarshal([]byte(`{"time":"2010-09-08T07:06:05.4321Z","level":"info","message":"message"}`), &wanted)
				_ = json.Unmarshal(dest.Bytes(), &got)
				test.Map(t, got).Equals(wanted)
			},
		},
		{scenario: "with int field",
			exec: func(t *testing.T) {
				// ARRANGE
				e.logcontext = &logcontext{
					fields: &fields{
						mutex: mx,
						m:     map[string]any{"ikey": 99},
						b:     map[int][]byte{},
					},
				}
				wanted := map[string]any{}
				got := map[string]any{}

				// ACT
				sut.Format(0, e, dest)

				// ASSERT
				_ = json.Unmarshal([]byte(`{"time":"2010-09-08T07:06:05.4321Z","level":"info","message":"message","ikey":99}`), &wanted)
				_ = json.Unmarshal(dest.Bytes(), &got)
				test.Map(t, got).Equals(wanted)
			},
		},
		{scenario: "with string field",
			exec: func(t *testing.T) {
				// ARRANGE
				e.logcontext = &logcontext{
					fields: &fields{
						mutex: mx,
						m:     map[string]any{"key": "value"},
						b:     map[int][]byte{},
					},
				}
				wanted := map[string]any{}
				got := map[string]any{}

				// ACT
				sut.Format(0, e, dest)

				// ASSERT
				_ = json.Unmarshal([]byte(`{"time":"2010-09-08T07:06:05.4321Z","level":"info","message":"message","key":"value"}`), &wanted)
				_ = json.Unmarshal(dest.Bytes(), &got)
				test.Map(t, got).Equals(wanted)
			},
		},
		{scenario: "with bool field",
			exec: func(t *testing.T) {
				// ARRANGE
				e.logcontext = &logcontext{
					fields: &fields{
						mutex: mx,
						m:     map[string]any{"key": true},
						b:     map[int][]byte{},
					},
				}
				wanted := map[string]any{}
				got := map[string]any{}

				// ACT
				sut.Format(0, e, dest)

				// ASSERT
				_ = json.Unmarshal([]byte(`{"time":"2010-09-08T07:06:05.4321Z","level":"info","message":"message","key":true}`), &wanted)
				_ = json.Unmarshal(dest.Bytes(), &got)
				test.Map(t, got).Equals(wanted)
			},
		},
		{scenario: "struct field",
			exec: func(t *testing.T) {
				// ARRANGE
				type str struct {
					A int
					B string
				}
				e.logcontext = &logcontext{
					fields: &fields{
						mutex: mx,
						m:     map[string]any{"key": str{A: 1, B: "two"}},
						b:     map[int][]byte{},
					},
				}
				wanted := map[string]any{}
				got := map[string]any{}

				// ACT
				sut.Format(0, e, dest)

				// ASSERT
				_ = json.Unmarshal([]byte(`{"time":"2010-09-08T07:06:05.4321Z","level":"info","message":"message","key":{"A":1,"B":"two"}}`), &wanted)
				_ = json.Unmarshal(dest.Bytes(), &got)
				test.Map(t, got).Equals(wanted)
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			// ARRANGE
			mx.Reset()
			e = entry{
				logcontext: &logcontext{},
				Time:       tm,
				Level:      InfoLevel,
				Message:    "message",
			}
			dest.Reset()

			// ACT
			tc.exec(t)

			// ASSERT
			IsSyncSafe(t, true, mx) // the current implementation of the json formatter is lockless, the mutex should never be acquired
		})
	}
}
