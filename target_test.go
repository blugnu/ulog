package ulog

import (
	"bytes"
	"errors"
	"reflect"
	"testing"
)

func TestTarget_close(t *testing.T) {
	// ARRANGE
	sut := &target{
		transport: &mocktransport{},
	}

	// ACT
	sut.close()

	// ASSERT
	t.Run("calls Transport.stop()", func(t *testing.T) {
		wanted := true
		got := sut.transport.(*mocktransport).stopWasCalled
		if wanted != got {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})
}

func TestTarget_dispatch(t *testing.T) {
	// ARRANGE
	sut := &target{
		buf:       &bytes.Buffer{},
		Formatter: &mockformatter{},
		transport: &mocktransport{},
	}
	e := entry{}

	// ACT
	sut.dispatch(e)

	// ASSERT
	t.Run("formats the log with Formatter", func(t *testing.T) {
		wanted := true
		got := sut.Formatter.(*mockformatter).formatWasCalled
		if wanted != got {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})

	t.Run("sends log to Transport", func(t *testing.T) {
		wanted := true
		got := sut.transport.(*mocktransport).logWasCalled
		if wanted != got {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})
}

func TestTarget(t *testing.T) {
	// ARRANGE
	cfgWasApplied := false

	mx := &mux{}
	cfg := func() TargetOption {
		return func(mx *mux, t *target) error {
			cfgWasApplied = true
			return nil
		}
	}

	// ACT
	err := Target(cfg())(mx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// ASSERT
	t.Run("applies options", func(t *testing.T) {
		wanted := true
		got := cfgWasApplied
		if wanted != got {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})

	t.Run("adds target to mux", func(t *testing.T) {
		wanted := true
		got := len(mx.targets) == 1
		if !reflect.DeepEqual(wanted, got) {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})
}

func TestTarget_whenConfigurationFails(t *testing.T) {
	// ARRANGE
	cfgerr := errors.New("target error")

	mx := &mux{}
	cfg := func() TargetOption {
		return func(mx *mux, t *target) error {
			return cfgerr
		}
	}

	// ACT
	err := Target(cfg())(mx)

	// ASSERT
	t.Run("option error is returned", func(t *testing.T) {
		wanted := cfgerr
		got := err
		if !errors.Is(got, wanted) {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})
}

func TestTargetFormat_sharedFormat(t *testing.T) {
	// ARRANGE
	mux := &mux{
		formats: map[string]*formatref{
			"id": {
				idx:       0,
				Formatter: &mockformatter{},
			},
		},
	}
	tg := &target{}

	// ACT
	err := TargetFormat("id")(mux, tg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// ASSERT
	t.Run("sets Formatter", func(t *testing.T) {
		wanted := mux.formats["id"].Formatter
		got := tg.Formatter
		if wanted != got {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})
}

func TestTargetFormat_sharedFormatNotKnown(t *testing.T) {
	// ARRANGE
	mux := &mux{formats: map[string]*formatref{}}
	tg := &target{}

	// ACT
	err := TargetFormat("id")(mux, tg)

	// ASSERT
	t.Run("error", func(t *testing.T) {
		wanted := ErrUnknownFormat
		got := err
		if !errors.Is(got, wanted) {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})
}

func TestTargetFormat_targetSpecificFormat(t *testing.T) {
	// ARRANGE
	mux := &mux{formats: map[string]*formatref{}}
	tg := &target{}

	t.Run("specifying a Formatter", func(t *testing.T) {
		// ARRANGE
		f := &mockformatter{}
		defer func() { mux.formats = map[string]*formatref{} }()

		// ACT
		err := TargetFormat(f)(mux, tg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// ASSERT
		t.Run("sets Formatter", func(t *testing.T) {
			wanted := f
			got := tg.Formatter
			if wanted != got {
				t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
			}
		})

		t.Run("adds format to mux", func(t *testing.T) {
			wanted := map[string]*formatref{
				"unregistered format: 0": {
					idx:       0,
					Formatter: f,
				},
			}
			got := mux.formats
			if !reflect.DeepEqual(wanted, got) {
				t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
			}
		})
	})

	t.Run("specifying a FormatterFactory", func(t *testing.T) {
		t.Run("when factory is valid", func(t *testing.T) {
			// ARRANGE
			mf := &mockformatter{}
			defer func() { mux.formats = map[string]*formatref{} }()

			// ACT
			err := TargetFormat(func() (Formatter, error) { return mf, nil })(mux, tg)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// ASSERT
			t.Run("sets Formatter", func(t *testing.T) {
				wanted := mf
				got := tg.Formatter
				if wanted != got {
					t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
				}
			})

			t.Run("adds format to mux", func(t *testing.T) {
				wanted := map[string]*formatref{
					"unregistered format: 0": {
						idx:       0,
						Formatter: mf,
					},
				}
				got := mux.formats
				if !reflect.DeepEqual(wanted, got) {
					t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
				}
			})
		})

		t.Run("when factory returns error", func(t *testing.T) {
			// ARRANGE
			fcterr := errors.New("factory error")

			// ACT
			err := TargetFormat(func() (Formatter, error) { return nil, fcterr })(mux, tg)

			// ASSERT
			t.Run("returns error", func(t *testing.T) {
				wanted := fcterr
				got := err
				if !errors.Is(got, wanted) {
					t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
				}
			})
		})
	})
}

func TestTargetFormat_invalidFormatRefType(t *testing.T) {
	// ACT
	err := TargetFormat(42)(nil, nil)

	// ASSERT
	t.Run("error", func(t *testing.T) {
		wanted := ErrInvalidFormatReference
		got := err
		if !errors.Is(got, wanted) {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})
}

func TestTargetLevel(t *testing.T) {
	// ARRANGE
	tg := &target{}

	// ACT
	err := TargetLevel(InfoLevel)(nil, tg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// ASSERT
	t.Run("sets Level", func(t *testing.T) {
		wanted := InfoLevel
		got := tg.Level
		if wanted != got {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})
}

func TestTargetTransport(t *testing.T) {
	t.Run("with valid options", func(t *testing.T) {
		// ARRANGE
		tg := &target{}
		tp := &mocktransport{}
		cfg := func(*mux, *target) (transport, error) {
			return tp, nil
		}

		// ACT
		err := TargetTransport(cfg)(nil, tg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// ASSERT
		t.Run("sets Tranport", func(t *testing.T) {
			wanted := tp
			got := tg.transport
			if wanted != got {
				t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
			}
		})
	})

	t.Run("with option error", func(t *testing.T) {
		// ARRANGE
		tg := &target{}
		opterr := errors.New("option error")
		cfg := func(*mux, *target) (transport, error) {
			return nil, opterr
		}

		// ACT
		err := TargetTransport(cfg)(nil, tg)

		// ASSERT
		t.Run("returns error", func(t *testing.T) {
			wanted := opterr
			got := err
			if !errors.Is(got, wanted) {
				t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
			}
		})
	})
}
