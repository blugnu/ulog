package ulog

import "testing"

func IsSyncSafe(t *testing.T, syncsafe bool, obj any) {
	t.Helper()
	t.Run("sync safe", func(t *testing.T) {
		t.Helper()

		wanted := true
		switch ss := obj.(type) {
		case *mockmutex:
			switch {
			case syncsafe:
				t.Run("sync safe code did not acquire/release mutex unnecessarily", func(t *testing.T) {
					got := !ss.lockWasCalled && !ss.unlockWasCalled
					if wanted != got {
						t.Errorf("\nwanted %v\ngot    %v", wanted, got)
					}
				})
			case !syncsafe:
				t.Run("mutex lock acquired", func(t *testing.T) {
					got := ss.lockWasCalled
					if wanted != got {
						t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
					}
				})
				t.Run("mutex lock released", func(t *testing.T) {
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
					got := !ss.GetWasCalled() && !ss.PutWasCalled()
					if wanted != got {
						t.Errorf("\nwanted %v\ngot    %v", wanted, got)
					}
				})
			case !syncsafe:
				t.Run("got from pool", func(t *testing.T) {
					got := ss.GetWasCalled()
					if wanted != got {
						t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
					}
				})
				t.Run("put to pool", func(t *testing.T) {
					got := ss.PutWasCalled()
					if wanted != got {
						t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
					}
				})
			}
		}
	})

}
