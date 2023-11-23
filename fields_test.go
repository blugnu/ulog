package ulog

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/blugnu/test"
)

func Test_newFields(t *testing.T) {
	t.Run("with capacity == 0", func(t *testing.T) {
		// ACT
		got := newFields(0)

		// ASSERT
		wanted := (*fields)(nil)
		if wanted != got {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})

	t.Run("with capacity > 0", func(t *testing.T) {
		// ACT
		got := newFields(1)

		// ASSERT
		wanted := true
		t.Run("initialises", func(t *testing.T) {
			t.Run("mutex", func(t *testing.T) {
				got := got.mutex != nil
				if wanted != got {
					t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
				}
			})

			t.Run("m", func(t *testing.T) {
				got := got.m != nil
				if wanted != got {
					t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
				}
			})

			t.Run("b", func(t *testing.T) {
				got := got.b != nil
				if wanted != got {
					t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
				}
			})
		})
	})
}

func Test_fields_merge(t *testing.T) {
	// ARRANGE
	mx := &mockmutex{}
	sut := &fields{
		mutex: mx,
		m: map[string]any{
			"k1": "v1",
			"k2": "v2",
		},
	}

	// ARRANGE
	type result struct {
		isCopy bool
		fields map[string]any
	}
	testcases := []struct {
		name     string
		fn       func(sut *fields) *fields
		syncsafe bool
		want     result
	}{
		{name: "merge no new key", fn: func(sut *fields) *fields { return sut.merge(nil) }, syncsafe: true, want: result{isCopy: false, fields: map[string]any{"k1": "v1", "k2": "v2"}}},
		{name: "merge new key", fn: func(sut *fields) *fields { return sut.merge(map[string]any{"k3": "v3"}) }, want: result{isCopy: true, fields: map[string]any{"k1": "v1", "k2": "v2", "k3": "v3"}}},
		{name: "merge existing key", fn: func(sut *fields) *fields { return sut.merge(map[string]any{"k3": "modified"}) }, want: result{isCopy: true, fields: map[string]any{"k1": "v1", "k2": "v2", "k3": "modified"}}},
		{name: "merge new and existing keys", fn: func(sut *fields) *fields { return sut.merge(map[string]any{"k1": "modified", "k3": "v3"}) }, want: result{isCopy: true, fields: map[string]any{"k1": "modified", "k2": "v2", "k3": "v3"}}},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// ARRANGE
			defer mx.Reset()

			// ACT
			cpy := tc.fn(sut)

			// ASSERT
			IsSyncSafe(t, tc.syncsafe, mx)

			t.Run("returns a copy", func(t *testing.T) {
				wanted := tc.want.isCopy
				got := reflect.ValueOf(sut).Pointer() != reflect.ValueOf(cpy).Pointer()
				if wanted != got {
					t.Errorf("\nwanted %v\ngot    %v", wanted, got)
				}

				t.Run("with fields", func(t *testing.T) {
					wanted := tc.want.fields
					got := cpy.m
					test.Maps(t, wanted, got)
				})
			})
		})
	}
}

func Test_fields_getFormattedBytes(t *testing.T) {
	// ARRANGE
	mx := &mockmutex{}

	testcases := []struct {
		name     string
		sut      *fields
		syncsafe bool
		result   *bytes.Buffer
	}{
		{name: "nil fields", sut: nil, syncsafe: true, result: nil},
		{name: "new format", sut: &fields{mutex: mx, b: map[int][]byte{}}, result: bytes.NewBuffer(nil)},
		{name: "cached format", sut: &fields{mutex: mx, b: map[int][]byte{1: []byte("content")}}, result: bytes.NewBufferString("content")},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// ARRANGE
			defer mx.Reset()

			// ACT
			got := tc.sut.getFormattedBytes(1)

			// ASSERT
			IsSyncSafe(t, tc.syncsafe, mx)

			wanted := tc.result
			switch {
			case wanted == nil:
				if wanted != got {
					t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
				}
			case wanted != nil:
				if !bytes.Equal(wanted.Bytes(), got.Bytes()) {
					t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
				}
			}
		})
	}
}

func Test_fields_setFormattedBytes(t *testing.T) {
	// ARRANGE
	mx := &mockmutex{}
	sut := &fields{mutex: mx, b: map[int][]byte{}}

	// ACT
	sut.setFormattedBytes(1, []byte("content"))

	// ASSERT
	IsSyncSafe(t, false, mx)

	wanted := map[int][]byte{1: []byte("content")}
	got := sut.b
	if !reflect.DeepEqual(wanted, got) {
		t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
	}
}
