package ulog

import (
	"context"
	"testing"

	"github.com/blugnu/test"
)

func TestRegisterEnrichment(t *testing.T) {
	t.Run("registered enrichments", func(t *testing.T) {
		// ARRANGE
		test.That(t, len(enrichment), "initial").Equals(0)

		oef := enrichment
		defer func() { enrichment = oef }()

		// ACT
		RegisterEnrichment(func(ctx context.Context) map[string]any { return nil })

		// ASSERT
		test.That(t, len(enrichment), "after registration").Equals(1)
	})
}
