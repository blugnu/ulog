package ulog

import (
	"errors"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/blugnu/test"
)

func TestLoggerConfiguration(t *testing.T) {
	t.Run("LogCallsite", func(t *testing.T) {
		// ARRANGE
		lg := &logger{}

		// ACT
		_ = LogCallsite(true)(lg)

		// ASSERT
		t.Run("sets callsite flag", func(t *testing.T) {
			wanted := true
			got := lg.getCallsite != nil
			if wanted != got {
				t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
			}
		})
	})

	t.Run("LoggerBackend", func(t *testing.T) {
		// ARRANGE
		be := &mockbackend{}
		lg := &logger{}

		// ACT
		_ = LoggerBackend(be)(lg)

		// ASSERT
		t.Run("sets backend", func(t *testing.T) {
			wanted := be
			got := lg.backend
			if wanted != got {
				t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
			}
		})
	})

	t.Run("LoggerLevel", func(t *testing.T) {
		// ARRANGE
		lg := &logger{}

		// ACT
		_ = LoggerLevel(DebugLevel)(lg)

		// ASSERT
		t.Run("sets level", func(t *testing.T) {
			wanted := DebugLevel
			got := lg.Level
			if wanted != got {
				t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
			}
		})
	})

	t.Run("LoggerFormat", func(t *testing.T) {
		t.Run("when factory fails to create format", func(t *testing.T) {
			// ARRANGE
			facterr := errors.New("factory error")
			lg := &logger{}

			// ACT
			err := LoggerFormat(func() (Formatter, error) { return nil, facterr })(lg)

			// ASSERT
			t.Run("returns error", func(t *testing.T) {
				wanted := facterr
				got := err
				if !errors.Is(got, wanted) {
					t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
				}
			})
		})

		t.Run("when no backend is configured", func(t *testing.T) {
			// ARRANGE
			mf := &mockformatter{}
			lg := &logger{}

			// ACT
			err := LoggerFormat(func() (Formatter, error) { return mf, nil })(lg)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// ASSERT
			t.Run("installs stdio backend", func(t *testing.T) {
				wanted := true
				_, got := lg.backend.(*stdio)
				if wanted != got {
					t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
				}
			})

			t.Run("sets format", func(t *testing.T) {
				wanted := mf
				got := lg.backend.(*stdio).Formatter
				if wanted != got {
					t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
				}
			})
		})

		t.Run("with backend that fails to apply the format", func(t *testing.T) {
			// ARRANGE
			setformatWasCalled := false
			fmterr := errors.New("formatter error")
			mf := &mockformatter{}
			lg := &logger{backend: &mockbackend{
				setformatfn: func(f Formatter) error { setformatWasCalled = true; return fmterr },
			}}

			// ACT
			err := LoggerFormat(func() (Formatter, error) { return mf, nil })(lg)

			// ASSERT
			t.Run("calls setFormat on backend", func(t *testing.T) {
				wanted := true
				got := setformatWasCalled
				if wanted != got {
					t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
				}
			})

			t.Run("returns setFormat error", func(t *testing.T) {
				wanted := fmterr
				got := err
				if !errors.Is(got, wanted) {
					t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
				}
			})
		})

		t.Run("with mux backend", func(t *testing.T) {
			// ARRANGE
			mf := &mockformatter{}
			lg := &logger{backend: &mux{}}

			// ACT
			err := LoggerFormat(func() (Formatter, error) { return mf, nil })(lg)

			// ASSERT
			test.ErrorIs(t, ErrInvalidConfiguration, err)
		})
	})

	t.Run("LoggerOutput", func(t *testing.T) {
		t.Run("when no backend is configured", func(t *testing.T) {
			// ARRANGE
			lg := &logger{}

			// ACT
			err := LoggerOutput(io.Discard)(lg)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// ASSERT
			t.Run("installs stdio backend", func(t *testing.T) {
				wanted := true
				_, got := lg.backend.(*stdio)
				if wanted != got {
					t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
				}
			})

			t.Run("sets output", func(t *testing.T) {
				wanted := io.Discard
				got := lg.backend.(*stdio).Writer
				if wanted != got {
					t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
				}
			})
		})

		t.Run("with backend that fails to apply the output", func(t *testing.T) {
			// ARRANGE
			setoutputWasCalled := false
			outerr := errors.New("output error")
			lg := &logger{backend: &mockbackend{
				setoutputfn: func(w io.Writer) error { setoutputWasCalled = true; return outerr },
			}}

			// ACT
			err := LoggerOutput(io.Discard)(lg)

			// ASSERT
			t.Run("calls setOutput on backend", func(t *testing.T) {
				wanted := true
				got := setoutputWasCalled
				if wanted != got {
					t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
				}
			})

			t.Run("returns setOutput error", func(t *testing.T) {
				wanted := outerr
				got := err
				if !errors.Is(got, wanted) {
					t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
				}
			})
		})

		t.Run("with mux backend", func(t *testing.T) {
			// ARRANGE
			lg := &logger{backend: &mux{}}

			// ACT
			err := LoggerOutput(io.Discard)(lg)

			// ASSERT
			test.ErrorIs(t, ErrInvalidConfiguration, err)
		})
	})
}

func TestStdioConfiguration(t *testing.T) {
	t.Run("StdioTransport", func(t *testing.T) {
		t.Run("with no options", func(t *testing.T) {
			// ARRANGE

			// ACT
			result, err := StdioTransport()(nil, nil)
			test.UnexpectedError(t, err)

			sut := result.(*stdio)

			// ASSERT
			t.Run("configures a default formatter", func(t *testing.T) {
				wanted := true
				got := sut.Formatter != nil
				if wanted != got {
					t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
				}
			})

			t.Run("configures default Writer", func(t *testing.T) {
				test.Equal(t, io.Writer(os.Stdout), sut.Writer)
			})
		})

		t.Run("with valid options", func(t *testing.T) {
			// ARRANGE
			cfgApplied := false
			cfg := func(*stdio) error { cfgApplied = true; return nil }

			// ACT
			result, err := StdioTransport(cfg)(nil, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			sut := result.(*stdio)

			// ASSERT
			t.Run("applies configuration functions", func(t *testing.T) {
				wanted := true
				got := cfgApplied
				if wanted != got {
					t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
				}
			})

			t.Run("configures default formatter", func(t *testing.T) {
				wanted := true
				got := sut.Formatter != nil
				if wanted != got {
					t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
				}
			})

			t.Run("configures default Writer", func(t *testing.T) {
				wanted := true
				got := sut.Writer != nil
				if wanted != got {
					t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
				}
			})
		})

		t.Run("with option error", func(t *testing.T) {
			// ARRANGE
			opterr := errors.New("option error")
			cfg := func(*stdio) error { return opterr }

			// ACT
			result, err := StdioTransport(cfg)(nil, nil)

			// ASSERT
			t.Run("returns nil transport", func(t *testing.T) {
				wanted := (transport)(nil)
				if wanted != result {
					t.Errorf("\nwanted %#v\ngot    %#v", wanted, result)
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
	})

	t.Run("StdioOutput", func(t *testing.T) {
		// ARRANGE
		sut := &stdio{}

		// ACT
		err := StdioOutput(os.Stdout)(sut)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// ASSERT
		wanted := &stdio{
			Writer: os.Stdout,
		}
		got := sut
		if !reflect.DeepEqual(wanted, got) {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})
}
