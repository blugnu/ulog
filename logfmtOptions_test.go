package ulog

import (
	"errors"
	"reflect"
	"testing"
)

func TestLogfmt(t *testing.T) {
	t.Run("applies options", func(t *testing.T) {
		// ARRANGE
		optWasApplied := false
		opt := func(*logfmt) error { optWasApplied = true; return nil }

		// ACT
		_, err := LogfmtFormatter(opt)()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// ASSERT
		wanted := true
		got := optWasApplied
		if wanted != got {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})

	t.Run("with no options", func(t *testing.T) {
		// ACT
		got, err := LogfmtFormatter()()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// ASSERT
		t.Run("configures default key labels and level values", func(t *testing.T) {
			wanted := &logfmt{
				keys:   [numFields][]byte{[]byte(`time=`), []byte(` level=`), []byte(` message="`), []byte(` file="`), []byte(` function="`)},
				levels: [numLevels][]byte{{}, []byte("FATAL"), []byte("ERROR"), []byte("WARN "), []byte("INFO "), []byte("DEBUG"), []byte("TRACE")},
			}
			if !reflect.DeepEqual(wanted, got) {
				t.Errorf("\nwanted %s\ngot    %s", wanted, got)
			}
		})
	})

	t.Run("with option errors", func(t *testing.T) {
		// ARRANGE
		opterr := errors.New("option error")
		opt := func(*logfmt) error { return opterr }

		// ACT
		got, err := LogfmtFormatter(opt)()

		// ASSERT
		t.Run("returns nil formatter", func(t *testing.T) {
			wanted := (Formatter)(nil)
			if wanted != got {
				t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
			}
		})

		t.Run("returns error", func(t *testing.T) {
			wanted := opterr
			got := err
			if !errors.Is(got, wanted) {
				t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
			}
		})
	})
}
func TestLogfmtKeys(t *testing.T) {
	// ARRANGE
	sut := &logfmt{
		keys: [numFields][]byte{},
	}

	// ACT
	_ = LogfmtFieldNames(map[FieldId]string{
		TimeField:    "tm",
		LevelField:   "lv",
		MessageField: "msg",
	})(sut)

	// ASSERT
	wanted := [numFields][]byte{{}, {}, {}}
	wanted[TimeField] = []byte("tm=")
	wanted[LevelField] = []byte(" lv=")
	wanted[MessageField] = []byte(" msg=\"")

	got := sut.keys
	if !reflect.DeepEqual(wanted, got) {
		t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
	}
}

func TestLogfmtLevels(t *testing.T) {
	// ARRANGE
	f, _ := LogfmtFormatter()()
	sut := f.(*logfmt)

	// ACT
	_ = LogfmtLevelLabels(map[Level]string{
		TraceLevel: "diag",
		DebugLevel: "debug",
		ErrorLevel: "error",
		FatalLevel: "fatal",
	})(sut)

	// ASSERT
	wanted := [numLevels][]byte{
		{},
		TraceLevel: []byte("diag "),
		DebugLevel: []byte("debug"),
		InfoLevel:  []byte("INFO "),
		WarnLevel:  []byte("WARN "),
		ErrorLevel: []byte("error"),
		FatalLevel: []byte("fatal"),
	}
	got := sut.levels
	if !reflect.DeepEqual(wanted, got) {
		t.Errorf("\nwanted %s\ngot    %s", wanted, got)
	}
}
