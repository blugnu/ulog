package ulog

import (
	"bytes"
	"sync"
	"testing"

	"github.com/blugnu/test"
)

func Test_newFields(t *testing.T) {
	// ARRANGE
	testcases := []struct {
		scenario string
		exec     func(t *testing.T)
	}{
		{scenario: "newFields(0)",
			exec: func(t *testing.T) {
				// ACT
				result := newFields(0)

				// ASSERT
				test.That(t, result).IsNil()
			},
		},
		{scenario: "newFields(1)",
			exec: func(t *testing.T) {
				// ACT
				result := newFields(1)

				// ASSERT
				test.That(t, result).Equals(&fields{
					mutex: &sync.Mutex{},
					m:     make(map[string]any),
					b:     map[int][]byte{},
				})
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.exec(t)
		})
	}
}

func Test_fields(t *testing.T) {
	// ARRANGE
	mx := &mockmutex{}
	sut := &fields{mutex: mx}

	testcases := []struct {
		scenario string
		exec     func(t *testing.T)
	}{
		// merge tests
		{scenario: "merge(nil)",
			exec: func(t *testing.T) {
				// ARRANGE
				sut.m = map[string]any{
					"k1": "v1",
					"k2": "v2",
				}

				// ACT
				result := sut.merge(nil)

				// ASSERT
				IsSyncSafe(t, true, mx)
				test.Value(t, result).Equals(sut)
				test.That(t, result.m).Equals(map[string]any{
					"k1": "v1",
					"k2": "v2",
				})
			},
		},
		{scenario: "merge(new key)",
			exec: func(t *testing.T) {
				// ARRANGE
				sut.m = map[string]any{
					"k1": "v1",
					"k2": "v2",
				}

				// ACT
				result := sut.merge(map[string]any{"k3": "v3"})

				// ASSERT
				IsSyncSafe(t, false, mx)
				test.Value(t, result).DoesNotEqual(sut)
				test.That(t, result.m).Equals(map[string]any{
					"k1": "v1",
					"k2": "v2",
					"k3": "v3",
				})
			},
		},
		{scenario: "merge(modify existing key)",
			exec: func(t *testing.T) {
				// ARRANGE
				sut.m = map[string]any{
					"k1": "v1",
					"k2": "v2",
				}

				// ACT
				result := sut.merge(map[string]any{"k2": "modified"})

				// ASSERT
				IsSyncSafe(t, false, mx)
				test.Value(t, result).DoesNotEqual(sut)
				test.That(t, result.m).Equals(map[string]any{
					"k1": "v1",
					"k2": "modified",
				})
			},
		},
		{scenario: "merge(add key and modify existing key)",
			exec: func(t *testing.T) {
				// ARRANGE
				sut.m = map[string]any{
					"k1": "v1",
					"k2": "v2",
				}

				// ACT
				result := sut.merge(map[string]any{"k1": "modified", "k3": "v3"})

				// ASSERT
				IsSyncSafe(t, false, mx)
				test.Value(t, result).DoesNotEqual(sut)
				test.That(t, result.m).Equals(map[string]any{
					"k1": "modified",
					"k2": "v2",
					"k3": "v3",
				})
			},
		},

		// getFormattedBytes tests
		{scenario: "getFormattedBytes(id)/nil receiver",
			exec: func(t *testing.T) {
				// ARRANGE
				id := 1

				// ACT
				result := ((*fields)(nil)).getFormattedBytes(id)

				// ASSERT
				IsSyncSafe(t, true, mx)
				test.That(t, result).IsNil()
			},
		},
		{scenario: "getFormattedBytes(id)/new id",
			exec: func(t *testing.T) {
				// ARRANGE
				id := 1
				sut.b = map[int][]byte{}

				// ACT
				result := sut.getFormattedBytes(id)

				// ASSERT
				IsSyncSafe(t, false, mx)
				test.That(t, result).Equals(bytes.NewBuffer(nil))
			},
		},
		{scenario: "getFormattedBytes(id)/cached id",
			exec: func(t *testing.T) {
				// ARRANGE
				id := 1
				sut.b = map[int][]byte{id: []byte("content")}

				// ACT
				result := sut.getFormattedBytes(id)

				// ASSERT
				IsSyncSafe(t, false, mx)
				test.That(t, result).Equals(bytes.NewBufferString("content"))
			},
		},

		// setFormattedBytes tests
		{scenario: "setFormattedBytes(id, content)",
			exec: func(t *testing.T) {
				// ARRANGE
				id := 1
				sut.b = map[int][]byte{}

				// ACT
				sut.setFormattedBytes(id, []byte("content"))

				// ASSERT
				IsSyncSafe(t, false, mx)
				test.That(t, sut.b).Equals(map[int][]byte{id: []byte("content")})
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			defer func() {
				mx.Reset()
				sut = &fields{mutex: mx}
			}()
			tc.exec(t)
		})
	}
}
