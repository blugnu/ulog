package ulog

import (
	"bytes"
	"os"
	"reflect"
	"testing"
)

func TestStdio_initStdio(t *testing.T) {
	// ACT
	got := initStdioBackend(&mockformatter{}, os.Stdout)

	// ASSERT
	t.Run("sets formatter", func(t *testing.T) {
		wanted := true
		got := got.Formatter != nil
		if wanted != got {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})

	t.Run("sets writer", func(t *testing.T) {
		wanted := true
		got := got.Writer != nil
		if wanted != got {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})

	t.Run("initialises bytes.Buffer pool", func(t *testing.T) {
		wanted := true
		got := got.bufs != nil
		if wanted != got {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})

	t.Run("initialise buffer pool", func(t *testing.T) {
		wanted := bytes.NewBuffer([]byte{})
		got := got.bufs.Get()
		if !reflect.DeepEqual(wanted, got) {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})
}

func TestStdioBackend(t *testing.T) {
	// ARRANGE
	buf := &bytes.Buffer{}
	sut := initStdioBackend(&mockformatter{}, buf)
	sut.bufs = &mockpool[bytes.Buffer]{}

	// ACT
	sut.dispatch(entry{Message: "test"})

	// ASSERT
	wanted := "test\n"
	got := buf.String()
	if wanted != got {
		t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
	}

	IsSyncSafe(t, false, sut.bufs)
}

func TestStdio_setFormat(t *testing.T) {
	// ARRANGE
	sut := initStdioBackend(&mockformatter{}, os.Stdout)

	// ACT
	_ = sut.setFormatter(&mockformatter{})

	// ASSERT
	wanted := true
	got := sut.Formatter != nil
	if wanted != got {
		t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
	}
}

func TestStdio_setOutput(t *testing.T) {
	// ARRANGE
	sut := initStdioBackend(&mockformatter{}, os.Stdout)

	// ACT
	_ = sut.setOutput(os.Stdout)

	// ASSERT
	wanted := true
	got := sut.Writer != nil
	if wanted != got {
		t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
	}
}

func TestStdio_Log(t *testing.T) {
	// ARRANGE
	buf := &bytes.Buffer{}
	sut := initStdioTransport(buf)

	// ACT
	sut.log([]byte("test"))

	// ASSERT
	wanted := "test\n"
	got := buf.String()
	if wanted != got {
		t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
	}
}
