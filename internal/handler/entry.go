package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/drywaters/learnd/internal/model"
	"github.com/drywaters/learnd/internal/repository"
	"github.com/drywaters/learnd/internal/ui/partials"
	"github.com/drywaters/learnd/internal/urlutil"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// EntryHandler handles entry CRUD operations
type EntryHandler struct {
	entryRepo *repository.EntryRepository
}

// NewEntryHandler creates a new EntryHandler
func NewEntryHandler(entryRepo *repository.EntryRepository) *EntryHandler {
	return &EntryHandler{
		entryRepo: entryRepo,
	}
}

// Create handles creating a new entry
func (h *EntryHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		h.htmxError(w, "Invalid form data")
		return
	}

	url := strings.TrimSpace(r.FormValue("url"))
	if url == "" {
		h.htmxError(w, "URL is required")
		return
	}

	allowDuplicate := r.FormValue("allow_duplicate") == "1"

	normalizedURL := url
	if normalized, err := urlutil.NormalizeURL(url); err == nil {
		normalizedURL = normalized
	}

	if !allowDuplicate {
		existing, err := h.entryRepo.GetLatestByNormalizedURL(ctx, normalizedURL)
		if err != nil {
			h.htmxError(w, "Failed to check duplicates")
			return
		}
		if existing != nil {
			title := ""
			if existing.Title != nil && *existing.Title != "" {
				title = *existing.Title
			}

			partials.DuplicateWarning(true, title, existing.CreatedAt).Render(ctx, w)
			return
		}
	}

	// Parse tags from comma-separated string
	tagsStr := r.FormValue("tags")
	var tags []string
	if tagsStr != "" {
		for _, tag := range strings.Split(tagsStr, ",") {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				tags = append(tags, tag)
			}
		}
	}

	// Parse optional fields
	var timeSpent *int
	if ts := r.FormValue("time_spent"); ts != "" {
		if v, err := strconv.Atoi(ts); err == nil && v > 0 {
			// Convert minutes to seconds
			seconds := v * 60
			timeSpent = &seconds
		}
	}

	var quantity *int
	if q := r.FormValue("quantity"); q != "" {
		if v, err := strconv.Atoi(q); err == nil && v > 0 {
			quantity = &v
		}
	}

	var notes *string
	if n := strings.TrimSpace(r.FormValue("notes")); n != "" {
		notes = &n
	}

	input := &model.CreateEntryInput{
		SourceURL:        url,
		NormalizedURL:    normalizedURL,
		Tags:             tags,
		TimeSpentSeconds: timeSpent,
		Quantity:         quantity,
		Notes:            notes,
	}

	entry, err := h.entryRepo.Create(ctx, input)
	if err != nil {
		h.htmxError(w, "Failed to create entry")
		return
	}

	// Get updated entry count
	count, err := h.entryRepo.Count(ctx)
	if err != nil {
		count = 1 // Fallback to at least 1 since we just created one
	}

	w.Header().Set("X-Entry-Created", "true")

	// Trigger toast and return the new entry row
	h.htmxToast(w, "Entry saved", &entry.ID, "")

	// Render entry row
	duplicateCount := getDuplicateCount(ctx, h.entryRepo, entry)
	entryView := buildEntryView(entry, duplicateCount)
	partials.EntryRow(entryView).Render(ctx, w)

	if duplicateCount > 1 {
		duplicates, err := h.entryRepo.ListByNormalizedURL(ctx, entry.NormalizedURL)
		if err == nil {
			for _, duplicate := range duplicates {
				if duplicate.ID == entry.ID {
					continue
				}
				duplicateView := buildEntryView(&duplicate, duplicateCount)
				duplicateView.SwapOOB = true
				partials.EntryRow(duplicateView).Render(ctx, w)
			}
		}
	}

	// Render OOB swap for entry count
	partials.EntryCount(count).Render(ctx, w)

	// Render OOB swap to remove empty state
	partials.EmptyState(false).Render(ctx, w)

	// Clear duplicate warning
	partials.DuplicateWarning(false, "", time.Time{}).Render(ctx, w)
}

// List returns entries as JSON
func (h *EntryHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	entries, err := h.entryRepo.List(ctx, repository.ListOptions{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		slog.Error("failed to list entries", "handler", "List", "error", err)
		http.Error(w, "Failed to list entries", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

// Get returns a single entry
func (h *EntryHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	entry, err := h.entryRepo.GetByID(ctx, id)
	if err != nil {
		slog.Error("failed to get entry", "handler", "Get", "id", id, "error", err)
		http.Error(w, "Failed to get entry", http.StatusInternalServerError)
		return
	}
	if entry == nil {
		http.Error(w, "Entry not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}

// Update updates an entry
func (h *EntryHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Parse tags
	tagsStr := r.FormValue("tags")
	var tags []string
	if tagsStr != "" {
		for _, tag := range strings.Split(tagsStr, ",") {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				tags = append(tags, tag)
			}
		}
	}

	// Parse optional fields
	var timeSpent *int
	if ts := r.FormValue("time_spent"); ts != "" {
		if v, err := strconv.Atoi(ts); err == nil && v > 0 {
			seconds := v * 60
			timeSpent = &seconds
		}
	}

	var quantity *int
	if q := r.FormValue("quantity"); q != "" {
		if v, err := strconv.Atoi(q); err == nil && v > 0 {
			quantity = &v
		}
	}

	var notes *string
	if n := strings.TrimSpace(r.FormValue("notes")); n != "" {
		notes = &n
	}

	input := &model.UpdateEntryInput{
		Tags:             tags,
		TimeSpentSeconds: timeSpent,
		Quantity:         quantity,
		Notes:            notes,
	}

	entry, err := h.entryRepo.Update(ctx, id, input)
	if err != nil {
		slog.Error("failed to update entry", "handler", "Update", "id", id, "error", err)
		http.Error(w, "Failed to update entry", http.StatusInternalServerError)
		return
	}
	if entry == nil {
		http.Error(w, "Entry not found", http.StatusNotFound)
		return
	}

	h.htmxToast(w, "Entry updated", &entry.ID, "")

	duplicateCount := getDuplicateCount(ctx, h.entryRepo, entry)
	entryView := buildEntryView(entry, duplicateCount)
	partials.EntryRow(entryView).Render(ctx, w)
}

// Delete removes an entry
func (h *EntryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	entry, err := h.entryRepo.GetByID(ctx, id)
	if err != nil {
		slog.Error("failed to get entry", "handler", "Delete", "id", id, "error", err)
		http.Error(w, "Failed to get entry", http.StatusInternalServerError)
		return
	}
	if entry == nil {
		http.Error(w, "Entry not found", http.StatusNotFound)
		return
	}

	normalizedURL := entry.NormalizedURL
	if normalizedURL == "" {
		if normalized, err := urlutil.NormalizeURL(entry.SourceURL); err == nil {
			normalizedURL = normalized
		}
	}

	if err := h.entryRepo.Delete(ctx, id); err != nil {
		slog.Error("failed to delete entry", "handler", "Delete", "id", id, "error", err)
		http.Error(w, "Failed to delete entry", http.StatusInternalServerError)
		return
	}

	// Get updated entry count
	count, err := h.entryRepo.Count(ctx)
	if err != nil {
		count = 0
	}

	h.htmxToast(w, "Entry deleted", &id, "")

	// Render OOB swap for entry count
	partials.EntryCount(count).Render(ctx, w)

	// Render OOB swap for empty state (show if no entries left)
	partials.EmptyState(count == 0).Render(ctx, w)

	if normalizedURL != "" {
		duplicates, err := h.entryRepo.ListByNormalizedURL(ctx, normalizedURL)
		if err == nil && len(duplicates) > 0 {
			duplicateCount := len(duplicates)
			for _, duplicate := range duplicates {
				duplicateView := buildEntryView(&duplicate, duplicateCount)
				duplicateView.SwapOOB = true
				partials.EntryRow(duplicateView).Render(ctx, w)
			}
		}
	}
}

// RefreshEnrichment resets enrichment status to pending
func (h *EntryHandler) RefreshEnrichment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.entryRepo.ResetEnrichment(ctx, id); err != nil {
		slog.Error("failed to reset enrichment", "handler", "RefreshEnrichment", "id", id, "error", err)
		http.Error(w, "Failed to reset enrichment", http.StatusInternalServerError)
		return
	}

	entry, err := h.entryRepo.GetByID(ctx, id)
	if err != nil || entry == nil {
		http.Error(w, "Entry not found", http.StatusNotFound)
		return
	}

	h.htmxToast(w, "Enrichment queued", &id, "")

	duplicateCount := getDuplicateCount(ctx, h.entryRepo, entry)
	entryView := buildEntryView(entry, duplicateCount)
	partials.EntryRow(entryView).Render(ctx, w)
}

// RefreshSummary resets summary status to pending
func (h *EntryHandler) RefreshSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.entryRepo.ResetSummary(ctx, id); err != nil {
		slog.Error("failed to reset summary", "handler", "RefreshSummary", "id", id, "error", err)
		http.Error(w, "Failed to reset summary", http.StatusInternalServerError)
		return
	}

	entry, err := h.entryRepo.GetByID(ctx, id)
	if err != nil || entry == nil {
		http.Error(w, "Entry not found", http.StatusNotFound)
		return
	}

	h.htmxToast(w, "Summary queued", &id, "")

	duplicateCount := getDuplicateCount(ctx, h.entryRepo, entry)
	entryView := buildEntryView(entry, duplicateCount)
	partials.EntryRow(entryView).Render(ctx, w)
}

// Status returns the status partial for an entry (for polling)
func (h *EntryHandler) Status(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	entry, err := h.entryRepo.GetByID(ctx, id)
	if err != nil || entry == nil {
		http.Error(w, "Entry not found", http.StatusNotFound)
		return
	}

	duplicateCount := getDuplicateCount(ctx, h.entryRepo, entry)
	entryView := buildEntryView(entry, duplicateCount)
	partials.EntryRow(entryView).Render(ctx, w)
}

func (h *EntryHandler) htmxError(w http.ResponseWriter, msg string) {
	h.htmxToast(w, msg, nil, "error")
	w.WriteHeader(http.StatusBadRequest)
}

func (h *EntryHandler) htmxToast(w http.ResponseWriter, msg string, entryID *uuid.UUID, toastType string) {
	showToast := map[string]string{
		"message": msg,
	}
	if entryID != nil {
		showToast["id"] = entryID.String()
	}
	if toastType != "" {
		showToast["type"] = toastType
	}

	payload := map[string]map[string]string{
		"showToast": showToast,
	}
	if data, err := json.Marshal(payload); err == nil {
		w.Header().Set("HX-Trigger", string(data))
	}
}
