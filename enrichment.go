package ulog

import "context"

// EnrichmentFunc provides the signature of a function that may be registered
// to extract log enrichment fields from a specified Context
//
// An EnrichmentFunc identifies any values in the supplied Context that should
// be added to the log entry and returns them in the form of a map[string]any
// of fields (keys) and values.
type EnrichmentFunc func(context.Context) map[string]any

// enrichment is a slice of registered enrichment functions.
var enrichment []EnrichmentFunc

// RegisterEnrichment registers a specified function.
func RegisterEnrichment(d EnrichmentFunc) {
	enrichment = append(enrichment, d)
}
