package ulog

import (
	"errors"
	"reflect"
	"testing"

	"github.com/blugnu/test"
)

func TestBatch_add(t *testing.T) {
	// ARRANGE
	batchesEqual := func(got, wanted *Batch) bool {
		return got.max == wanted.max &&
			got.size == wanted.size &&
			got.len == wanted.len &&
			reflect.DeepEqual(got.entries, wanted.entries)
	}

	testcases := []struct {
		scenario string
		exec     func(*testing.T)
	}{
		// add tests
		{scenario: "add/empty batch",
			exec: func(t *testing.T) {
				// ARRANGE
				handler := &mockBatchHandler{}
				sut := &Batch{entries: [][]byte{}, max: 10, batchHandler: handler}
				want := &Batch{entries: [][]byte{[]byte("foo")}, max: 10, size: 3, len: 1}

				// ACT
				sut.add([]byte("foo"))

				// ASSERT
				test.That(t, sut).Equals(want, batchesEqual)
				test.That(t, handler.sendCalls, "batches sent").Equals(0)
			},
		},
		{scenario: "add/full batch/flush ok",
			exec: func(t *testing.T) {
				// ARRANGE
				handler := &mockBatchHandler{}
				sut := &Batch{entries: [][]byte{}, max: 1, batchHandler: handler}
				want := &Batch{entries: [][]byte{}, max: 1, size: 0, len: 0}

				// ACT
				sut.add([]byte("foo"))

				// ASSERT
				test.That(t, sut).Equals(want, batchesEqual)
				test.That(t, handler.sendCalls, "batches sent").Equals(1)
			},
		},
		{scenario: "add/full batch/flush unsuccessful",
			exec: func(t *testing.T) {
				// ARRANGE
				handler := &mockBatchHandler{
					sendfn: func(*Batch) error { return errors.New("flush error") },
				}
				sut := &Batch{entries: [][]byte{}, max: 1, batchHandler: handler}
				want := &Batch{entries: [][]byte{[]byte("foo")}, max: 1, size: 3, len: 1}

				// ACT
				sut.add([]byte("foo"))

				// ASSERT
				test.That(t, sut).Equals(want, batchesEqual)
				test.That(t, handler.sendCalls, "batches sent").Equals(1)
				test.That(t, handler.sentEntries, "entries sent").Equals(0)
			},
		},

		// clear tests
		{scenario: "clear",
			exec: func(t *testing.T) {
				// ARRANGE
				sut := &Batch{entries: [][]byte{[]byte("foo")}, max: 10, size: 3, len: 1}
				want := &Batch{entries: [][]byte{}, max: 10}

				// ACT
				sut.clear()

				// ASSERT
				test.That(t, sut).Equals(want, batchesEqual)
				test.That(t, cap(sut.entries)).Equals(sut.max)
			},
		},

		// flush tests
		{scenario: "flush/empty batch",
			exec: func(t *testing.T) {
				// ARRANGE
				handler := &mockBatchHandler{}
				sut := &Batch{entries: [][]byte{}, max: 10, batchHandler: handler}

				// ACT
				sut.flush()

				// ASSERT
				test.That(t, handler.sendCalls, "batches sent").Equals(0)
			},
		},
		{scenario: "flush/non-empty batch/ok",
			exec: func(t *testing.T) {
				// ARRANGE
				handler := &mockBatchHandler{}
				sut := &Batch{entries: [][]byte{}, max: 10, len: 1, batchHandler: handler}
				want := &Batch{entries: [][]byte{}, max: 10}

				// ACT
				sut.flush()

				// ASSERT
				test.That(t, handler.sendCalls, "batches sent").Equals(1)
				test.That(t, handler.sentEntries, "entries sent").Equals(1)
				test.That(t, sut).Equals(want, batchesEqual)
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.exec(t)
		})
	}
}
