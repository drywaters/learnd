package handler

import (
	"net/http"

	"github.com/danielmerrison/learnd/internal/repository"
)

// CaptureHandler handles the main capture UI
type CaptureHandler struct {
	entryRepo *repository.EntryRepository
	templates TemplateRenderer
}

// NewCaptureHandler creates a new CaptureHandler
func NewCaptureHandler(entryRepo *repository.EntryRepository, templates TemplateRenderer) *CaptureHandler {
	return &CaptureHandler{
		entryRepo: entryRepo,
		templates: templates,
	}
}

// CapturePage renders the main capture page
func (h *CaptureHandler) CapturePage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get recent entries
	entries, err := h.entryRepo.List(ctx, repository.ListOptions{Limit: 20})
	if err != nil {
		http.Error(w, "Failed to load entries", http.StatusInternalServerError)
		return
	}

	entryViews := buildEntryViews(ctx, h.entryRepo, entries)

	data := map[string]interface{}{
		"Entries": entryViews,
	}

	if err := h.templates.RenderPage(w, "capture.html", data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}
