package ulog

import (
	"bytes"
	"errors"
	"sync"
	"testing"

	"github.com/blugnu/test"
)

func TestMux(t *testing.T) {
	// ARRANGE
	testcases := []struct {
		scenario string
		exec     func(t *testing.T)
	}{
		// Mux() tests
		{scenario: "Mux/option error",
			exec: func(t *testing.T) {
				// ARRANGE
				logger := &logger{}
				opterr := errors.New("option error")
				opt := func(mx *mux) error { return opterr }

				// ACT
				err := Mux(opt)(logger)

				// ASSERT
				test.Error(t, err).Is(opterr)
			},
		},
		{scenario: "Mux/no targets",
			exec: func(t *testing.T) {
				// ARRANGE
				logger := &logger{}
				optApplied := false
				opt := func(mx *mux) error { optApplied = true; return nil }

				// ACT
				err := Mux(opt)(logger)

				// ASSERT
				test.That(t, err).IsNil()
				test.IsTrue(t, optApplied)
				if mux, ok := test.IsType[*mux](t, logger.backend); ok {
					test.Slice(t, mux.targets).IsEmpty()
					test.That(t, mux.levelTargets).Equals([numLevels][]*target{})
				}
			},
		},
		{scenario: "Mux/info and error level targets",
			exec: func(t *testing.T) {
				// ARRANGE
				logger := &logger{}
				tgi := &target{Level: InfoLevel}
				tge := &target{Level: ErrorLevel}
				withTargets := func(mx *mux) error { mx.targets = append(mx.targets, []*target{tgi, tge}...); return nil }

				// ACT
				err := Mux(withTargets)(logger)

				// ASSERT
				test.That(t, err).IsNil()
				if mux, ok := test.IsType[*mux](t, logger.backend); ok {
					test.Slice(t, mux.levelTargets[levelNotSet]).Equals([]*target{})
					test.Slice(t, mux.levelTargets[TraceLevel]).Equals([]*target{})
					test.Slice(t, mux.levelTargets[DebugLevel]).Equals([]*target{})
					test.Slice(t, mux.levelTargets[InfoLevel]).Equals([]*target{tgi})
					test.Slice(t, mux.levelTargets[WarnLevel]).Equals([]*target{tgi})
					test.Slice(t, mux.levelTargets[ErrorLevel]).Equals([]*target{tgi, tge})
					test.Slice(t, mux.levelTargets[FatalLevel]).Equals([]*target{tgi, tge})
				}
			},
		},

		// init test
		{scenario: "init",
			exec: func(t *testing.T) {
				// ARRANGE
				mux := &mux{}

				// ACT
				mux.init()

				// ASSERT
				test.That(t, mux.ch).IsNotNil()
				test.That(t, mux.targets).IsNotNil()
			},
		},

		// close test
		{scenario: "close",
			exec: func(t *testing.T) {
				// ARRANGE
				mux := &mux{ch: make(chan entry)}
				channelIsClosed := false

				wg := &sync.WaitGroup{}
				wg.Add(1)
				go func() {
					<-mux.ch
					channelIsClosed = true
					wg.Done()
				}()

				// ACT
				mux.close()

				// CLEANUP
				wg.Wait()

				// ASSERT
				test.IsTrue(t, channelIsClosed)
			},
		},

		// dispatch test
		{scenario: "dispatch",
			exec: func(t *testing.T) {
				// ARRANGE
				mux := &mux{ch: make(chan entry)}
				e := entry{Level: InfoLevel, Message: "dispatched"}

				var sentToChannel entry
				wg := &sync.WaitGroup{}
				wg.Add(1)
				go func() {
					sentToChannel = <-mux.ch
					wg.Done()
				}()

				// ACT
				mux.dispatch(e)

				// CLEANUP
				close(mux.ch)
				wg.Wait()

				// ASSERT
				test.That(t, sentToChannel).Equals(e)
			},
		},

		// run test
		{scenario: "run",
			exec: func(t *testing.T) {
				// ARRANGE
				logger := &logger{}
				dtr := &mocktransport{}
				etr := &mocktransport{}
				etg := &target{ // an 'enabled' transport (level >= Info)
					Level:     InfoLevel,
					transport: etr,
					Formatter: &mockformatter{},
					buf:       &bytes.Buffer{},
				}
				dtg := &target{ // a 'disabled' transport (level < Info)
					Level:     ErrorLevel,
					transport: dtr,
					Formatter: &mockformatter{},
					buf:       &bytes.Buffer{},
				}
				_ = Mux(func(mx *mux) error { mx.targets = append(mx.targets, []*target{dtg, etg}...); return nil })(logger)
				mux := logger.backend.(*mux)

				e := entry{
					logcontext: &logcontext{
						logger: logger,
					},
					Level:   InfoLevel,
					Message: "dispatched",
				}

				// ACT
				wg := &sync.WaitGroup{}
				wg.Add(1)
				go func() {
					defer wg.Done()
					mux.run()
				}()
				mux.ch <- e

				// CLEANUP
				close(mux.ch)
				wg.Wait()

				// ASSERT
				test.IsTrue(t, etr.logWasCalled)
				test.IsFalse(t, dtr.logWasCalled)

				test.IsTrue(t, etr.stopWasCalled)
				test.IsTrue(t, dtr.stopWasCalled)
			},
		},

		// start test
		{scenario: "start",
			exec: func(t *testing.T) {
				// ARRANGE
				mt := &mocktransport{}
				mux := &mux{}
				mux.init()
				mux.targets = []*target{{transport: mt}}

				// ACT
				cfn, err := mux.start()

				// we must explicitly call the close function (rather than defer)
				// to ensure that our assertions are performed only after the mux
				// has completed initialisation
				cfn()

				// ASSERT
				test.Error(t, err).IsNil()
				test.IsTrue(t, mt.runWasCalled)
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.exec(t)
		})
	}
}
