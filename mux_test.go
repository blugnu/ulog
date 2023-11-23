package ulog

import (
	"bytes"
	"errors"
	"reflect"
	"sync"
	"testing"
)

func Test_initMux(t *testing.T) {
	// ACT
	got := initMux()

	// ASSERT
	t.Run("initialises channel", func(t *testing.T) {
		wanted := true
		got := got.ch != nil
		if wanted != got {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})

	t.Run("initialises slice of targets", func(t *testing.T) {
		wanted := true
		got := got.targets != nil
		if wanted != got {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})
}

func TestMux(t *testing.T) {
	// ARRANGE
	logger := &logger{}
	tgt := &target{Level: InfoLevel}
	cfg := func(mx *mux) error { mx.targets = append(mx.targets, tgt); return nil }
	sut := Mux(cfg)

	// ACT
	err := sut(logger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// ASSERT
	t.Run("applies configuration functions", func(t *testing.T) {
		wanted := true
		got := logger.backend != nil && len(logger.backend.(*mux).targets) == 1
		if wanted != got {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})

	t.Run("configures level targets", func(t *testing.T) {
	})
}

func TestMuxWhenConfigurationFails(t *testing.T) {
	// ARRANGE
	logger := &logger{}
	cfgerr := errors.New("configuration error")

	cfg := func(*mux) error { return cfgerr }
	sut := Mux(cfg)

	// ACT
	err := sut(logger)

	// ASSERT
	t.Run("returns configuration errors", func(t *testing.T) {
		wanted := cfgerr
		got := err
		if !errors.Is(got, wanted) {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})

	t.Run("does not configure backend", func(t *testing.T) {
		wanted := true
		got := logger.backend == nil
		if wanted != got {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})
}

func TestMux_close(t *testing.T) {
	// ARRANGE
	mt := &mocktransport{}
	sut := &mux{
		ch: make(chan entry),
		targets: []*target{
			{transport: mt},
		},
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-sut.ch // will not receive anything; will yield when the channel is closed
	}()

	// ACT / ASSERT
	// (nothing to assert; the test will timeout if the channel is not closed)
	sut.close()
	wg.Wait()
}

func TestMux_dispatch(t *testing.T) {
	// ARRANGE
	sut := &mux{
		ch: make(chan entry, 1),
	}
	e := entry{Level: InfoLevel, Message: "dispatched"}

	var got entry
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		got = <-sut.ch
		close(sut.ch)
	}()

	// ACT
	sut.dispatch(e)
	wg.Wait()

	// ASSERT
	wanted := entry{
		Level:   InfoLevel,
		Message: "dispatched",
	}
	if !reflect.DeepEqual(wanted, got) {
		t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
	}
}

func TestMux_start(t *testing.T) {
	// ARRANGE
	mt := &mocktransport{}
	sut := &mux{
		ch: make(chan entry),
		targets: []*target{
			{transport: mt},
		},
	}

	// ACT
	cfn, err := sut.start()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// we must explicitly call the close function (rather than defer)
	// to ensure that our assertions are performed only after the mux
	// has completed initialisation
	cfn()

	// ASSERT
	t.Run("calls run on transports", func(t *testing.T) {
		wanted := true
		got := mt.runWasCalled
		if wanted != got {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})
}

func TestMux_run(t *testing.T) {
	// ARRANGE
	dt := &mocktransport{}
	et := &mocktransport{}
	sut := &mux{
		ch: make(chan entry),
		targets: []*target{
			{ // an 'enabled' transport (level >= Info)
				Level:     InfoLevel,
				transport: et,
				Formatter: &mockformatter{},
				buf:       &bytes.Buffer{},
			},
			{ // a 'disabled' transport (level < Info)
				Level:     ErrorLevel,
				transport: dt,
				Formatter: &mockformatter{},
				buf:       &bytes.Buffer{},
			},
		},
		levelTargets: [numLevels][]*target{},
	}
	sut.levelTargets[InfoLevel] = append(sut.levelTargets[InfoLevel], sut.targets[0])
	sut.levelTargets[WarnLevel] = append(sut.levelTargets[WarnLevel], sut.targets[0])
	sut.levelTargets[ErrorLevel] = append(sut.levelTargets[ErrorLevel], sut.targets[0])
	sut.levelTargets[FatalLevel] = append(sut.levelTargets[FatalLevel], sut.targets[0])
	sut.levelTargets[ErrorLevel] = append(sut.levelTargets[ErrorLevel], sut.targets[1])
	sut.levelTargets[FatalLevel] = append(sut.levelTargets[FatalLevel], sut.targets[1])
	e := entry{
		logcontext: &logcontext{
			logger: &logger{},
		},
		Level:   InfoLevel,
		Message: "dispatched",
	}

	// ACT
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		sut.run()
	}()
	sut.ch <- e
	close(sut.ch)
	wg.Wait()

	// ASSERT
	t.Run("dispatches to enabled transports", func(t *testing.T) {
		wanted := true
		got := et.logWasCalled
		if wanted != got {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})

	t.Run("does NOT dispatch to disabled transports", func(t *testing.T) {
		wanted := true
		got := !dt.logWasCalled
		if wanted != got {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})

	t.Run("calls stop on transports", func(t *testing.T) {
		wanted := true
		got := et.stopWasCalled
		if wanted != got {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})
}

func TestMux_Format(t *testing.T) {
	// ARRANGE
	mx := &mux{formats: map[string]*formatref{}}

	t.Run("first format", func(t *testing.T) {
		// ACT
		err := Format("id", func() (Formatter, error) { return &mockformatter{}, nil })(mx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// ASSERT
		t.Run("added with index 0", func(t *testing.T) {
			wanted := map[string]*formatref{"id": {0, &mockformatter{}}}
			got := mx.formats
			if !reflect.DeepEqual(wanted, got) {
				t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
			}
		})
	})

	t.Run("second format", func(t *testing.T) {
		// ACT
		err := Format("id2", func() (Formatter, error) { return &mockformatter{}, nil })(mx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// ASSERT
		t.Run("added with index 1", func(t *testing.T) {
			wanted := map[string]*formatref{
				"id":  {0, &mockformatter{}},
				"id2": {1, &mockformatter{}},
			}
			got := mx.formats
			if !reflect.DeepEqual(wanted, got) {
				t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
			}
		})
	})

	t.Run("duplicate id", func(t *testing.T) {
		// ACT
		err := Format("id", func() (Formatter, error) { return &mockformatter{}, nil })(mx)

		// ASSERT
		t.Run("returns error", func(t *testing.T) {
			wanted := ErrFormatAlreadyRegistered
			got := err
			if !errors.Is(got, wanted) {
				t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
			}
		})
	})

	t.Run("factory error", func(t *testing.T) {
		// ARRANGE
		facterr := errors.New("factory error")

		// ACT
		err := Format("id", func() (Formatter, error) { return nil, facterr })(mx)

		// ASSERT
		t.Run("returns error", func(t *testing.T) {
			wanted := facterr
			got := err
			if !errors.Is(got, wanted) {
				t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
			}
		})
	})
}
