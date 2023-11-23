package ulog

import (
	"context"
	"testing"
)

func TestRegisterEnrichment(t *testing.T) {
	// ARRANGE
	oef := enrichment
	defer func() { enrichment = oef }()

	f := func(ctx context.Context) map[string]any { return nil }

	// ACT
	if len(oef) != 0 {
		t.Fatal("`decorators` is not empty")
	}
	RegisterEnrichment(f)

	// ASSERT
	wanted := 1
	got := len(enrichment)
	if wanted != got {
		t.Errorf("wanted %v, got %v", wanted, got)
	}
}
