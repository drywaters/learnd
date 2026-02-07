package worker

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/drywaters/learnd/internal/enricher"
	"github.com/drywaters/learnd/internal/model"
	"github.com/drywaters/learnd/internal/repository"
	"github.com/drywaters/learnd/internal/summarizer"
)

// Worker processes entries in the background
type Worker struct {
	entryRepo      *repository.EntryRepository
	cacheRepo      *repository.SummaryCacheRepository
	enrichRegistry *enricher.Registry
	summarizer     summarizer.Summarizer

	interval  time.Duration
	batchSize int

	stopCh chan struct{}
	wg     sync.WaitGroup
}

// Config holds worker configuration
type Config struct {
	Interval  time.Duration
	BatchSize int
}

// New creates a new background worker
func New(
	entryRepo *repository.EntryRepository,
	cacheRepo *repository.SummaryCacheRepository,
	enrichRegistry *enricher.Registry,
	sum summarizer.Summarizer,
	cfg Config,
) *Worker {
	if cfg.Interval == 0 {
		cfg.Interval = 10 * time.Second
	}
	if cfg.BatchSize == 0 {
		cfg.BatchSize = 5
	}

	return &Worker{
		entryRepo:      entryRepo,
		cacheRepo:      cacheRepo,
		enrichRegistry: enrichRegistry,
		summarizer:     sum,
		interval:       cfg.Interval,
		batchSize:      cfg.BatchSize,
		stopCh:         make(chan struct{}),
	}
}

// Start begins the background processing loops
func (w *Worker) Start(ctx context.Context) {
	slog.Info("starting background worker", "interval", w.interval, "batch_size", w.batchSize)

	w.wg.Add(2)
	go w.runEnrichmentLoop(ctx)
	go w.runSummarizationLoop(ctx)
}

// Stop gracefully stops the worker
func (w *Worker) Stop() {
	slog.Info("stopping background worker")
	close(w.stopCh)
	w.wg.Wait()
	slog.Info("background worker stopped")
}

func (w *Worker) runEnrichmentLoop(ctx context.Context) {
	defer w.wg.Done()

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processEnrichment(ctx)
		}
	}
}

func (w *Worker) runSummarizationLoop(ctx context.Context) {
	defer w.wg.Done()

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processSummarization(ctx)
		}
	}
}

func (w *Worker) processEnrichment(ctx context.Context) {
	entries, err := w.entryRepo.GetPendingEnrichment(ctx, w.batchSize)
	if err != nil {
		slog.Error("failed to get pending enrichment", "error", err)
		return
	}

	for _, entry := range entries {
		// Mark as processing
		if err := w.entryRepo.UpdateEnrichmentStatus(ctx, entry.ID, model.StatusProcessing, nil); err != nil {
			slog.Error("failed to update enrichment status", "id", entry.ID, "error", err)
			continue
		}

		// Enrich
		result, err := w.enrichRegistry.Enrich(ctx, entry.SourceURL)
		if err != nil {
			errMsg := err.Error()
			slog.Warn("enrichment failed", "id", entry.ID, "url", entry.SourceURL, "error", err)
			w.entryRepo.UpdateEnrichmentStatus(ctx, entry.ID, model.StatusFailed, &errMsg)
			continue
		}

		var metadataJSON []byte
		if len(result.Metadata) > 0 {
			metadataJSON, err = json.Marshal(result.Metadata)
			if err != nil {
				slog.Warn("failed to marshal enrichment metadata", "id", entry.ID, "error", err)
				metadataJSON = nil
			}
		}

		// Save enrichment result (sanitize text fields to remove invalid UTF-8)
		enrichResult := &repository.EnrichmentResult{
			CanonicalURL:   result.CanonicalURL,
			Domain:         result.Domain,
			SourceType:     result.SourceType,
			Title:          sanitizeUTF8(result.Title),
			Description:    sanitizeUTF8(result.Description),
			PublishedAt:    result.PublishedAt,
			RuntimeSeconds: result.RuntimeSeconds,
			MetadataJSON:   metadataJSON,
		}

		if err := w.entryRepo.UpdateEnrichmentResult(ctx, entry.ID, enrichResult); err != nil {
			slog.Error("failed to save enrichment result", "id", entry.ID, "error", err)
			continue
		}

		slog.Info("enriched entry", "id", entry.ID, "title", result.Title, "type", result.SourceType)
	}
}

func (w *Worker) processSummarization(ctx context.Context) {
	if w.summarizer == nil {
		return
	}

	entries, err := w.entryRepo.GetPendingSummary(ctx, w.batchSize)
	if err != nil {
		slog.Error("failed to get pending summary", "error", err)
		return
	}

	for _, entry := range entries {
		// Skip if no content to summarize
		if entry.Title == nil && entry.Description == nil {
			w.entryRepo.UpdateSummaryStatus(ctx, entry.ID, model.StatusSkipped, nil)
			continue
		}

		// Check cache first
		canonicalURL := entry.SourceURL
		if entry.CanonicalURL != nil {
			canonicalURL = *entry.CanonicalURL
		}
		urlHash := hashURL(canonicalURL)

		cached, err := w.cacheRepo.GetByURLHash(ctx, urlHash)
		if err == nil && cached != nil {
			// Use cached summary
			result := &repository.SummaryResult{
				Text:        cached.SummaryText,
				Provider:    cached.Provider,
				Model:       cached.Model,
				Version:     cached.Version,
				GeneratedAt: cached.CreatedAt,
			}
			if err := w.entryRepo.UpdateSummaryResult(ctx, entry.ID, result); err != nil {
				slog.Error("failed to save cached summary", "id", entry.ID, "error", err)
			}
			slog.Info("used cached summary", "id", entry.ID)
			continue
		}

		// Mark as processing
		if err := w.entryRepo.UpdateSummaryStatus(ctx, entry.ID, model.StatusProcessing, nil); err != nil {
			slog.Error("failed to update summary status", "id", entry.ID, "error", err)
			continue
		}

		// Build input
		tag := ""
		if entry.Tag != nil {
			tag = *entry.Tag
		}
		input := summarizer.Input{
			SourceType: entry.SourceType,
			URL:        entry.SourceURL,
			Tag:        tag,
		}
		if entry.Title != nil {
			input.Title = *entry.Title
		}
		if entry.Description != nil {
			input.Description = *entry.Description
		}

		// Generate summary
		result, err := w.summarizer.Summarize(ctx, input)
		if err != nil {
			errMsg := err.Error()
			slog.Warn("summarization failed", "id", entry.ID, "error", err)
			w.entryRepo.UpdateSummaryStatus(ctx, entry.ID, model.StatusFailed, &errMsg)
			continue
		}

		// Save to entry
		summaryResult := &repository.SummaryResult{
			Text:        result.Text,
			Provider:    result.Provider,
			Model:       result.Model,
			Version:     result.Version,
			GeneratedAt: result.GeneratedAt,
		}

		if err := w.entryRepo.UpdateSummaryResult(ctx, entry.ID, summaryResult); err != nil {
			slog.Error("failed to save summary result", "id", entry.ID, "error", err)
			continue
		}

		// Cache the summary
		cache := &model.SummaryCache{
			URLHash:      urlHash,
			CanonicalURL: canonicalURL,
			SummaryText:  result.Text,
			Provider:     result.Provider,
			Model:        result.Model,
			Version:      result.Version,
		}
		if err := w.cacheRepo.Store(ctx, cache); err != nil {
			slog.Warn("failed to cache summary", "id", entry.ID, "error", err)
		}

		slog.Info("summarized entry", "id", entry.ID)
	}
}

func hashURL(url string) string {
	h := sha256.New()
	h.Write([]byte(url))
	return hex.EncodeToString(h.Sum(nil))
}

// sanitizeUTF8 removes invalid UTF-8 byte sequences from a string.
// This prevents PostgreSQL errors when storing text that may contain
// malformed characters from web scraping.
func sanitizeUTF8(s string) string {
	return strings.ToValidUTF8(s, "")
}
