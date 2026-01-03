package enricher

import (
	"context"
	"sort"
	"time"

	"github.com/drywaters/learnd/internal/model"
)

// Result contains extracted metadata from a URL
type Result struct {
	CanonicalURL   string
	Domain         string
	SourceType     model.SourceType
	Title          string
	Description    string
	PublishedAt    *time.Time
	RuntimeSeconds *int
	Metadata       map[string]interface{}
}

// Enricher extracts metadata from URLs
type Enricher interface {
	// CanHandle returns true if this enricher can process the URL
	CanHandle(url string) bool

	// Enrich extracts metadata from the URL
	Enrich(ctx context.Context, url string) (*Result, error)

	// Name returns the enricher identifier
	Name() string

	// Priority returns the enricher priority (lower = higher priority)
	Priority() int
}

// Registry manages enrichers and routes URLs to appropriate handlers
type Registry struct {
	enrichers []Enricher
	fallback  Enricher
}

// NewRegistry creates a new enricher registry with a fallback enricher
func NewRegistry(fallback Enricher) *Registry {
	return &Registry{
		enrichers: make([]Enricher, 0),
		fallback:  fallback,
	}
}

// Register adds an enricher to the registry
func (r *Registry) Register(e Enricher) {
	r.enrichers = append(r.enrichers, e)
	// Sort by priority (lower = higher priority)
	sort.Slice(r.enrichers, func(i, j int) bool {
		return r.enrichers[i].Priority() < r.enrichers[j].Priority()
	})
}

// Enrich processes a URL using the appropriate enricher
func (r *Registry) Enrich(ctx context.Context, url string) (*Result, error) {
	for _, e := range r.enrichers {
		if e.CanHandle(url) {
			result, err := e.Enrich(ctx, url)
			if err == nil {
				return result, nil
			}
			// Log error but continue to fallback
		}
	}
	return r.fallback.Enrich(ctx, url)
}
