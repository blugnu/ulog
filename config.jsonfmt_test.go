package ulog

import (
	"errors"
	"reflect"
	"testing"
)

func TestNewJsonFormatter(t *testing.T) {
	t.Run("applies options", func(t *testing.T) {
		// ARRANGE
		optWasApplied := false
		opt := func(*jsonfmt) error { optWasApplied = true; return nil }

		// ACT
		_, err := NewJSONFormatter(opt)()
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
		got, err := NewJSONFormatter()()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// ASSERT
		t.Run("configures default key labels and level values", func(t *testing.T) {
			wanted := &jsonfmt{
				keys:   defaultJsonKeys,
				levels: defaultJsonLevels,
			}
			if !reflect.DeepEqual(wanted, got) {
				t.Errorf("\nwanted %s\ngot    %s", wanted, got)
			}
		})
	})

	t.Run("with option errors", func(t *testing.T) {
		// ARRANGE
		opterr := errors.New("option error")
		opt := func(*jsonfmt) error { return opterr }

		// ACT
		got, err := NewJSONFormatter(opt)()

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
func TestJsonKeys(t *testing.T) {
	// ARRANGE
	sut := &jsonfmt{
		keys: [numFields]string{},
	}

	// ACT
	_ = JsonLabels(map[FieldId]string{
		TimeField:    "tm",
		LevelField:   "lv",
		MessageField: "msg",
	})(sut)

	// ASSERT
	wanted := [numFields]string{"", "", ""}
	wanted[TimeField] = "tm"
	wanted[LevelField] = "lv"
	wanted[MessageField] = "msg"

	got := sut.keys
	if !reflect.DeepEqual(wanted, got) {
		t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
	}
}

func TestJsonLevels(t *testing.T) {
	// ARRANGE
	f, _ := NewJSONFormatter()()
	sut := f.(*jsonfmt)

	// ACT
	_ = JsonLevels(map[Level]string{
		TraceLevel: "TRACE",
		DebugLevel: "DEBUG",
		ErrorLevel: "ERROR",
		FatalLevel: "FATAL",
	})(sut)

	// ASSERT
	wanted := [numLevels]string{
		"",
		TraceLevel: "TRACE",
		DebugLevel: "DEBUG",
		InfoLevel:  defaultJsonLevels[InfoLevel],
		WarnLevel:  defaultJsonLevels[WarnLevel],
		ErrorLevel: "ERROR",
		FatalLevel: "FATAL",
	}
	got := sut.levels
	if !reflect.DeepEqual(wanted, got) {
		t.Errorf("\nwanted %s\ngot    %s", wanted, got)
	}
}
