package handler

import (
	"encoding/csv"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/drywaters/learnd/internal/model"
	"github.com/drywaters/learnd/internal/repository"
	"github.com/drywaters/learnd/internal/ui/pages"
	"github.com/drywaters/learnd/internal/ui/partials"
)

// ReportHandler handles reporting
type ReportHandler struct {
	entryRepo *repository.EntryRepository
}

// NewReportHandler creates a new ReportHandler
func NewReportHandler(entryRepo *repository.EntryRepository) *ReportHandler {
	return &ReportHandler{
		entryRepo: entryRepo,
	}
}

// ReportsPage renders the reports page
func (h *ReportHandler) ReportsPage(w http.ResponseWriter, r *http.Request) {
	pages.ReportsPage().Render(r.Context(), w)
}

// GetReport generates a report for the specified date range
func (h *ReportHandler) GetReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	var start, end time.Time
	var err error

	if startStr != "" {
		start, err = time.Parse("2006-01-02", startStr)
		if err != nil {
			http.Error(w, "Invalid start date", http.StatusBadRequest)
			return
		}
	} else {
		// Default to 30 days ago
		start = time.Now().AddDate(0, 0, -30)
	}

	if endStr != "" {
		end, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			http.Error(w, "Invalid end date", http.StatusBadRequest)
			return
		}
		// Include the full end day
		end = end.Add(24*time.Hour - time.Second)
	} else {
		end = time.Now()
	}

	const pageSize = 1000

	// Get all entries in the date range and compute aggregations
	tagCounts := make(map[string]int)
	tagTime := make(map[string]int)
	typeCounts := make(map[string]int)
	typeTime := make(map[string]int)
	totalEntries := 0
	totalTime := 0

	opts := repository.ListOptions{
		Limit:  pageSize,
		Offset: 0,
		Start:  &start,
		End:    &end,
	}

	for {
		entries, err := h.entryRepo.List(ctx, opts)
		if err != nil {
			slog.Error("failed to list entries", "handler", "GetReport", "offset", opts.Offset, "error", err)
			http.Error(w, "Failed to get entries", http.StatusInternalServerError)
			return
		}
		if len(entries) == 0 {
			break
		}

		for _, entry := range entries {
			totalEntries++
			trackedSeconds := reportTrackedSeconds(entry)
			if trackedSeconds > 0 {
				totalTime += trackedSeconds
			}

			// Aggregate by tags
			for _, tag := range entry.Tags {
				tagCounts[tag]++
				if trackedSeconds > 0 {
					tagTime[tag] += trackedSeconds
				}
			}

			// Aggregate by type
			typeCounts[string(entry.SourceType)]++
			if trackedSeconds > 0 {
				typeTime[string(entry.SourceType)] += trackedSeconds
			}
		}

		opts.Offset += len(entries)
	}

	// Build report data
	var tagReport []partials.TagReport
	totalTagEntries := 0
	totalTagTime := 0
	for tag, count := range tagCounts {
		minutes := minutesFromSeconds(tagTime[tag])
		totalTagEntries += count
		totalTagTime += minutes
		tagReport = append(tagReport, partials.TagReport{
			Tag:   tag,
			Count: count,
			Time:  minutes,
		})
	}

	var typeReport []partials.TypeReport
	totalTypeEntries := 0
	totalTypeTime := 0
	for typ, count := range typeCounts {
		minutes := minutesFromSeconds(typeTime[typ])
		totalTypeEntries += count
		totalTypeTime += minutes
		typeReport = append(typeReport, partials.TypeReport{
			Type:  typ,
			Count: count,
			Time:  minutes,
		})
	}

	data := partials.ReportData{
		Start:            start.Format("2006-01-02"),
		End:              end.Format("2006-01-02"),
		TotalEntries:     totalEntries,
		TotalTime:        minutesFromSeconds(totalTime),
		TotalTagEntries:  totalTagEntries,
		TotalTagTime:     totalTagTime,
		TotalTypeEntries: totalTypeEntries,
		TotalTypeTime:    totalTypeTime,
		ByTag:            tagReport,
		ByType:           typeReport,
	}

	partials.ReportResults(data).Render(ctx, w)
}

// ExportCSV exports entries as CSV
func (h *ReportHandler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	var start, end time.Time
	var err error

	if startStr != "" {
		start, err = time.Parse("2006-01-02", startStr)
		if err != nil {
			http.Error(w, "Invalid start date", http.StatusBadRequest)
			return
		}
	} else {
		start = time.Now().AddDate(0, 0, -30)
	}

	if endStr != "" {
		end, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			http.Error(w, "Invalid end date", http.StatusBadRequest)
			return
		}
		end = end.Add(24*time.Hour - time.Second)
	} else {
		end = time.Now()
	}

	const pageSize = 1000

	opts := repository.ListOptions{
		Limit:  pageSize,
		Offset: 0,
		Start:  &start,
		End:    &end,
	}

	// Fetch first page before writing headers to allow clean error response
	entries, err := h.entryRepo.List(ctx, opts)
	if err != nil {
		slog.Error("failed to list entries", "handler", "ExportCSV", "offset", opts.Offset, "error", err)
		http.Error(w, "Failed to get entries", http.StatusInternalServerError)
		return
	}

	// Now safe to write headers and begin streaming
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=learnd-export-%s.csv", time.Now().Format("2006-01-02")))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write CSV header
	writer.Write([]string{
		"Date", "URL", "Title", "Type", "Tags", "Time (min)", "Quantity", "Notes", "Summary",
	})

	for {
		// Process current page
		for _, entry := range entries {
			title := ""
			if entry.Title != nil {
				title = *entry.Title
			}

			tags := ""
			if len(entry.Tags) > 0 {
				for i, tag := range entry.Tags {
					if i > 0 {
						tags += ", "
					}
					tags += tag
				}
			}

			timeSpent := ""
			if trackedSeconds := reportTrackedSeconds(entry); trackedSeconds > 0 {
				timeSpent = fmt.Sprintf("%d", minutesFromSeconds(trackedSeconds))
			}

			quantity := ""
			if entry.Quantity != nil {
				quantity = fmt.Sprintf("%d", *entry.Quantity)
			}

			notes := ""
			if entry.Notes != nil {
				notes = *entry.Notes
			}

			summary := ""
			if entry.SummaryText != nil {
				summary = *entry.SummaryText
			}

			writer.Write([]string{
				entry.CreatedAt.Format("2006-01-02"),
				entry.SourceURL,
				title,
				string(entry.SourceType),
				tags,
				timeSpent,
				quantity,
				notes,
				summary,
			})
		}

		// Check if this was the last page
		if len(entries) < pageSize {
			break
		}

		// Fetch next page
		opts.Offset += len(entries)
		entries, err = h.entryRepo.List(ctx, opts)
		if err != nil {
			// Headers already sent, can only log and stop
			slog.Error("failed to list entries", "handler", "ExportCSV", "offset", opts.Offset, "error", err)
			return
		}
	}
}

func reportTrackedSeconds(entry model.Entry) int {
	if entry.TimeSpentSeconds != nil && *entry.TimeSpentSeconds > 0 {
		return *entry.TimeSpentSeconds
	}
	if entry.RuntimeSeconds != nil && *entry.RuntimeSeconds > 0 {
		return *entry.RuntimeSeconds
	}
	return 0
}

func minutesFromSeconds(seconds int) int {
	if seconds <= 0 {
		return 0
	}
	return (seconds + 59) / 60
}
