package handler

import (
	"context"

	"github.com/drywaters/learnd/internal/model"
	"github.com/drywaters/learnd/internal/repository"
	"github.com/drywaters/learnd/internal/ui"
)

func buildEntryView(ctx context.Context, repo *repository.EntryRepository, entry *model.Entry) ui.EntryView {
	duplicateCount := 1
	normalizedURL := entry.NormalizedURL
	if normalizedURL == "" {
		normalizedURL = entry.SourceURL
	}
	if normalizedURL != "" {
		if count, err := repo.CountByNormalizedURL(ctx, normalizedURL); err == nil && count > 0 {
			duplicateCount = count
		}
	}

	return ui.EntryView{
		Entry:          *entry,
		DuplicateCount: duplicateCount,
		SwapOOB:        false,
	}
}

func buildEntryViews(ctx context.Context, repo *repository.EntryRepository, entries []model.Entry) []ui.EntryView {
	views := make([]ui.EntryView, 0, len(entries))
	if len(entries) == 0 {
		return views
	}

	normalizedURLs := make([]string, 0, len(entries))
	for _, entry := range entries {
		normalizedURL := entry.NormalizedURL
		if normalizedURL == "" {
			normalizedURL = entry.SourceURL
		}
		if normalizedURL != "" {
			normalizedURLs = append(normalizedURLs, normalizedURL)
		}
	}

	counts := map[string]int{}
	if len(normalizedURLs) > 0 {
		if fetched, err := repo.GetDuplicateCountsByNormalizedURL(ctx, normalizedURLs); err == nil {
			counts = fetched
		}
	}

	for _, entry := range entries {
		normalizedURL := entry.NormalizedURL
		if normalizedURL == "" {
			normalizedURL = entry.SourceURL
		}

		duplicateCount := 1
		if normalizedURL != "" {
			if count, ok := counts[normalizedURL]; ok && count > 0 {
				duplicateCount = count
			}
		}

		views = append(views, ui.EntryView{
			Entry:          entry,
			DuplicateCount: duplicateCount,
			SwapOOB:        false,
		})
	}

	return views
}
