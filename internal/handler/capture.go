package handler

import (
	"log/slog"
	"net/http"

	"github.com/drywaters/learnd/internal/repository"
	"github.com/drywaters/learnd/internal/ui/pages"
)

// CaptureHandler handles the main capture UI
type CaptureHandler struct {
	entryRepo *repository.EntryRepository
}

// NewCaptureHandler creates a new CaptureHandler
func NewCaptureHandler(entryRepo *repository.EntryRepository) *CaptureHandler {
	return &CaptureHandler{
		entryRepo: entryRepo,
	}
}

// CapturePage renders the main capture page
func (h *CaptureHandler) CapturePage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get recent entries
	entries, err := h.entryRepo.List(ctx, repository.ListOptions{Limit: 20})
	if err != nil {
		slog.Error("failed to list entries", "handler", "CapturePage", "error", err)
		http.Error(w, "Failed to load entries", http.StatusInternalServerError)
		return
	}

	entryViews := buildEntryViews(ctx, h.entryRepo, entries)

	if err := pages.CapturePage(entryViews).Render(ctx, w); err != nil {
		// Log only - response may already be partially written, can't send clean http.Error
		slog.Error("failed to render page", "handler", "CapturePage", "error", err)
	}
}
