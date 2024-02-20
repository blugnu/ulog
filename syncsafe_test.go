package ulog

import "testing"

// a test helper that checks whether code behaves in a sync safe manner.
//
// If syncsafe is true, then the code is expected to be sync safe and
// should not perform any unnecessary sync operations.
//
// If syncsafe is false, then the code is expected to be sync unsafe and
// should perform sync operations.
//
// The obj parameter is the object that is being tested.  It must be a
// *mockmutex or any type that implements the following interface:
//
//	interface {
//		GetWasCalled() bool
//		PutWasCalled() bool
//	}
func IsSyncSafe(t *testing.T, syncsafe bool, obj any) {
	t.Helper()
	t.Run("is sync safe", func(t *testing.T) {
		t.Helper()

		wanted := true
		switch ss := obj.(type) {
		case *mockmutex:
			switch {
			case syncsafe:
				t.Run("mutex not acquired/released unnecessarily", func(t *testing.T) {
					t.Helper()
					got := !ss.lockWasCalled && !ss.unlockWasCalled
					if wanted != got {
						t.Errorf("\nwanted %v\ngot    %v", wanted, got)
					}
				})
			case !syncsafe:
				t.Run("mutex lock is acquired", func(t *testing.T) {
					t.Helper()
					got := ss.lockWasCalled
					if wanted != got {
						t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
					}
				})
				t.Run("mutex lock is released", func(t *testing.T) {
					t.Helper()
					got := ss.unlockWasCalled
					if wanted != got {
						t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
					}
				})
			}
		case interface {
			GetWasCalled() bool
			PutWasCalled() bool
		}:
			switch {
			case syncsafe:
				t.Run("sync safe code did not call Get/Put unnecessarily", func(t *testing.T) {
					t.Helper()
					got := !ss.GetWasCalled() && !ss.PutWasCalled()
					if wanted != got {
						t.Errorf("\nwanted %v\ngot    %v", wanted, got)
					}
				})
			case !syncsafe:
				t.Run("got from pool", func(t *testing.T) {
					t.Helper()
					got := ss.GetWasCalled()
					if wanted != got {
						t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
					}
				})
				t.Run("put to pool", func(t *testing.T) {
					t.Helper()
					got := ss.PutWasCalled()
					if wanted != got {
						t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
					}
				})
			}
		}
	})

}
