package model

import (
	"time"

	"github.com/google/uuid"
)

// SourceType represents the type of content source
type SourceType string

const (
	SourceTypeYouTube SourceType = "youtube"
	SourceTypePodcast SourceType = "podcast"
	SourceTypeArticle SourceType = "article"
	SourceTypeDoc     SourceType = "doc"
	SourceTypeOther   SourceType = "other"
)

// ProcessingStatus represents the status of async processing
type ProcessingStatus string

const (
	StatusPending    ProcessingStatus = "pending"
	StatusProcessing ProcessingStatus = "processing"
	StatusOK         ProcessingStatus = "ok"
	StatusFailed     ProcessingStatus = "failed"
	StatusSkipped    ProcessingStatus = "skipped"
)

// Entry represents a learning log entry
type Entry struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// User input
	SourceURL        string   `json:"source_url"`
	NormalizedURL    string   `json:"normalized_url"`
	Tags             []string `json:"tags"`
	TimeSpentSeconds *int     `json:"time_spent_seconds,omitempty"`
	Quantity         *int     `json:"quantity,omitempty"`
	Notes            *string  `json:"notes,omitempty"`

	// Enriched fields
	CanonicalURL   *string    `json:"canonical_url,omitempty"`
	Domain         *string    `json:"domain,omitempty"`
	SourceType     SourceType `json:"source_type"`
	Title          *string    `json:"title,omitempty"`
	Description    *string    `json:"description,omitempty"`
	PublishedAt    *time.Time `json:"published_at,omitempty"`
	RuntimeSeconds *int       `json:"runtime_seconds,omitempty"`
	MetadataJSON   []byte     `json:"metadata_json,omitempty"`

	// Enrichment status
	EnrichmentStatus ProcessingStatus `json:"enrichment_status"`
	EnrichmentError  *string          `json:"enrichment_error,omitempty"`
	EnrichedAt       *time.Time       `json:"enriched_at,omitempty"`

	// Summary fields
	SummaryText        *string          `json:"summary_text,omitempty"`
	SummaryStatus      ProcessingStatus `json:"summary_status"`
	SummaryError       *string          `json:"summary_error,omitempty"`
	SummaryProvider    *string          `json:"summary_provider,omitempty"`
	SummaryModel       *string          `json:"summary_model,omitempty"`
	SummaryVersion     *string          `json:"summary_version,omitempty"`
	SummaryGeneratedAt *time.Time       `json:"summary_generated_at,omitempty"`
}

// CreateEntryInput represents input for creating a new entry
type CreateEntryInput struct {
	SourceURL        string
	NormalizedURL    string
	Tags             []string
	TimeSpentSeconds *int
	Quantity         *int
	Notes            *string
}

// UpdateEntryInput represents input for updating an entry
type UpdateEntryInput struct {
	Tags             []string
	TimeSpentSeconds *int
	Quantity         *int
	Notes            *string
}

// SummaryCache represents a cached summary for a URL
type SummaryCache struct {
	ID           uuid.UUID `json:"id"`
	URLHash      string    `json:"url_hash"`
	CanonicalURL string    `json:"canonical_url"`
	SummaryText  string    `json:"summary_text"`
	Provider     string    `json:"provider"`
	Model        string    `json:"model"`
	Version      string    `json:"version"`
	CreatedAt    time.Time `json:"created_at"`
}
