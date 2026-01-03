package summarizer

import (
	"context"
	"time"

	"github.com/drywaters/learnd/internal/model"
)

// Input contains the data to be summarized
type Input struct {
	Title       string
	Description string
	SourceType  model.SourceType
	URL         string
	Tags        []string
}

// Result contains the generated summary and metadata
type Result struct {
	Text        string
	Provider    string
	Model       string
	Version     string
	GeneratedAt time.Time
}

// Summarizer defines the interface for AI-powered text summarization
type Summarizer interface {
	// Summarize generates a concise summary from the provided input
	// Returns a 1-2 sentence summary suitable for learning logs
	Summarize(ctx context.Context, input Input) (*Result, error)

	// Provider returns the provider identifier (e.g., "gemini", "openai")
	Provider() string

	// Model returns the specific model being used
	Model() string

	// Version returns the implementation version for tracking changes
	Version() string
}
