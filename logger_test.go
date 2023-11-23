package ulog

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"
)

func Test_New_whenConfigurationFails(t *testing.T) {
	// ARRANGE
	ctx := context.Background()
	cfgerr := errors.New("configuration error")
	cfg := func(*logger) error { return cfgerr }

	// ACT
	lg, cfn, err := NewLogger(ctx, cfg)

	// ASSERT
	t.Run("returns error", func(t *testing.T) {
		wanted := cfgerr
		got := err
		if !errors.Is(got, wanted) {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})

	t.Run("nil logger", func(t *testing.T) {
		wanted := (Logger)(nil)
		got := lg
		if wanted != got {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})

	t.Run("nil close function", func(t *testing.T) {
		wanted := true
		got := cfn == nil
		if wanted != got {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})
}

func Test_New_withNoBackendConfigured(t *testing.T) {
	// ARRANGE
	ctx := context.Background()

	// ACT
	lg, cfn, err := NewLogger(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// ASSERT
	t.Run("non-nil logger", func(t *testing.T) {
		wanted := true
		got := lg != nil
		if wanted != got {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})

	t.Run("non-nil close function", func(t *testing.T) {
		wanted := true
		got := cfn != nil
		if wanted != got {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})

	t.Run("configures default backend", func(t *testing.T) {
		wanted := &stdio{}
		got, ok := lg.(*logcontext).logger.backend.(*stdio)
		if !ok {
			t.Errorf("\nwanted %T\ngot    %T", wanted, got)
		}

		t.Run("with default Formatter", func(t *testing.T) {
			wanted, _ := Logfmt()()
			got := lg.(*logcontext).logger.backend.(*stdio).Formatter
			if !reflect.DeepEqual(wanted, got) {
				t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
			}
		})

		t.Run("using stdout", func(t *testing.T) {
			wanted := os.Stdout
			got := lg.(*logcontext).logger.backend.(*stdio).Writer
			if wanted != got {
				t.Errorf("\nwanted %v\ngot    %v", wanted, got)
			}
		})
	})

	t.Run("close function", func(t *testing.T) {
		// ARRANGE/ASSERT
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("should not panic")
			}
		}()

		// ACT
		cfn()
	})
}

func Test_New_withStartableBackendThatFailsToStart(t *testing.T) {
	// ARRANGE
	ctx := context.Background()

	beerr := errors.New("backend error")
	be := &mockbackend{
		startfn: func() (func(), error) { return nil, beerr },
	}

	// ACT
	lg, cfn, err := NewLogger(ctx, LoggerBackend(be))

	// ASSERT
	t.Run("returns error", func(t *testing.T) {
		wanted := beerr
		got := err
		if !errors.Is(got, wanted) {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})

	t.Run("nil logger", func(t *testing.T) {
		wanted := (Logger)(nil)
		got := lg
		if wanted != got {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})

	t.Run("nil close function", func(t *testing.T) {
		wanted := true
		got := cfn == nil
		if wanted != got {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})
}

func Test_initLogger(t *testing.T) {
	// ARRANGE
	ctx := context.Background()
	cfgApplied := false

	// register an enrichment func to ensure coverage (we can't test function refs)
	og := enrichment
	defer func() { enrichment = og }()
	enrichment = []EnrichmentFunc{func(context.Context) map[string]any { return nil }}

	cfg := func(*logger) error { cfgApplied = true; return nil }

	// ACT
	lg, ic, err := initLogger(ctx, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// ASSERT
	t.Run("applies configuration functions", func(t *testing.T) {
		wanted := true
		got := cfgApplied
		if wanted != got {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})

	t.Run("sets default level", func(t *testing.T) {
		wanted := InfoLevel
		got := lg.Level
		if wanted != got {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})

	t.Run("returns initial context (no callsite info)", func(t *testing.T) {
		wanted := &logcontext{
			ctx:    ctx,
			logger: lg,
		}
		got := ic
		if !reflect.DeepEqual(wanted, got) {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})

	t.Run("returns initial context (with callsite info)", func(t *testing.T) {
		// ACT
		lg, got, err := initLogger(ctx, func(lg *logger) error { lg.getCallsite = caller; return nil })
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// ASSERT
		wanted := &logcontext{
			ctx:    ctx,
			logger: lg,
		}
		if !reflect.DeepEqual(wanted, got) {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})
}

func Test_initLogger_whenConfigurationFails(t *testing.T) {
	// ARRANGE
	ctx := context.Background()
	cfgerr := errors.New("configuration error")
	cfg := func(*logger) error { return cfgerr }

	// ACT
	lg, ic, err := initLogger(ctx, cfg)

	// ASSERT
	t.Run("returns error", func(t *testing.T) {
		wanted := cfgerr
		got := err
		if !errors.Is(got, wanted) {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})

	t.Run("nil logger", func(t *testing.T) {
		wanted := (*logger)(nil)
		got := lg
		if wanted != got {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})

	t.Run("nil initial context", func(t *testing.T) {
		wanted := (*logcontext)(nil)
		got := ic
		if wanted != got {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})
}

func TestLogger_log(t *testing.T) {
	// ARRANGE
	wasDispatched := false

	sut := &logger{
		backend: &mockbackend{dispatchfn: func(e entry) { wasDispatched = true }},
	}

	t.Run("nil entry", func(t *testing.T) {
		// ACT
		sut.log(noop.entry)

		// ASSERT
		t.Run("is not dispatched", func(t *testing.T) {
			wanted := true
			got := !wasDispatched
			if wanted != got {
				t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
			}
		})
	})

	t.Run("non-nil entry", func(t *testing.T) {
		// ACT
		sut.log(entry{
			logcontext: &logcontext{
				logger: sut,
			},
		})

		// ASSERT
		t.Run("is dispatched", func(t *testing.T) {
			wanted := true
			got := wasDispatched
			if wanted != got {
				t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
			}
		})
	})
}

func TestLogger_levelEnabled(t *testing.T) {
	// ARRANGE
	ctx := context.Background()
	sut := &logger{}

	// ACT
	for _, ll := range Levels {
		sut.Level = ll
		for _, el := range Levels {
			t.Run(fmt.Sprintf("logger: %s, entry: %s", ll, el), func(t *testing.T) {
				// ARRANGE
				sut.Level = ll

				// ACT
				got := sut.levelEnabled(ctx, el)

				// ASSERT
				wanted := el <= ll
				if wanted != got {
					t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
				}
			})
		}
	}
}

func TestLogger_noEnrichment(t *testing.T) {
	// ARRANGE
	type key int

	ctx := context.Background()
	newctx := context.WithValue(ctx, key(1), 0)
	logctx := &logcontext{
		dispatcher: &mockdispatcher{},
	}
	sut := &logger{getCallsite: noCaller}
	logctx.logger = sut

	// ACT
	got := sut.noEnrichment(logctx, newctx)

	// ASSERT
	t.Run("assigns context", func(t *testing.T) {
		wanted := newctx
		got := got.ctx
		if wanted != got {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})

	t.Run("assigns dispatcher", func(t *testing.T) {
		wanted := logctx.dispatcher
		got := got.dispatcher
		if wanted != got {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})

	t.Run("context fields", func(t *testing.T) {
		wanted := 0
		got := len(got.xfields)
		if wanted != got {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})
}

func TestLogger_withEnrichment(t *testing.T) {
	// ARRANGE
	type key int

	ctx := context.Background()
	newctx := context.WithValue(ctx, key(1), 0)
	logctx := &logcontext{
		dispatcher: &mockdispatcher{},
	}
	sut := &logger{getCallsite: noCaller}
	logctx.logger = sut

	og := enrichment
	defer func() { enrichment = og }()
	RegisterEnrichment(func(context.Context) map[string]any { return map[string]any{"key": "value"} })

	// ACT
	got := sut.withEnrichment(logctx, newctx)

	// ASSERT
	t.Run("assigns context", func(t *testing.T) {
		wanted := newctx
		got := got.ctx
		if wanted != got {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})

	t.Run("assigns dispatcher", func(t *testing.T) {
		wanted := logctx.dispatcher
		got := got.dispatcher
		if wanted != got {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})

	t.Run("context fields", func(t *testing.T) {
		wanted := map[string]any{"key": "value"}
		got := got.xfields
		if !reflect.DeepEqual(wanted, got) {
			t.Errorf("\nwanted %#v\ngot    %#v", wanted, got)
		}
	})
}
