package handler

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"time"

	"github.com/danielmerrison/learnd/internal/repository"
)

// ReportHandler handles reporting
type ReportHandler struct {
	entryRepo *repository.EntryRepository
	templates TemplateRenderer
}

// NewReportHandler creates a new ReportHandler
func NewReportHandler(entryRepo *repository.EntryRepository, templates TemplateRenderer) *ReportHandler {
	return &ReportHandler{
		entryRepo: entryRepo,
		templates: templates,
	}
}

// ReportsPage renders the reports page
func (h *ReportHandler) ReportsPage(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{}

	if err := h.templates.RenderPage(w, "reports.html", data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
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
	var filtered []interface{}
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
			http.Error(w, "Failed to get entries", http.StatusInternalServerError)
			return
		}
		if len(entries) == 0 {
			break
		}

		for _, entry := range entries {
			totalEntries++
			if entry.TimeSpentSeconds != nil {
				totalTime += *entry.TimeSpentSeconds
			}

			// Aggregate by tags
			for _, tag := range entry.Tags {
				tagCounts[tag]++
				if entry.TimeSpentSeconds != nil {
					tagTime[tag] += *entry.TimeSpentSeconds
				}
			}

			// Aggregate by type
			typeCounts[string(entry.SourceType)]++
			if entry.TimeSpentSeconds != nil {
				typeTime[string(entry.SourceType)] += *entry.TimeSpentSeconds
			}

			filtered = append(filtered, entry)
		}

		opts.Offset += len(entries)
	}

	// Build report data
	var tagReport []map[string]interface{}
	for tag, count := range tagCounts {
		tagReport = append(tagReport, map[string]interface{}{
			"tag":   tag,
			"count": count,
			"time":  tagTime[tag] / 60, // Convert to minutes
		})
	}

	var typeReport []map[string]interface{}
	for typ, count := range typeCounts {
		typeReport = append(typeReport, map[string]interface{}{
			"type":  typ,
			"count": count,
			"time":  typeTime[typ] / 60,
		})
	}

	data := map[string]interface{}{
		"Start":        start.Format("2006-01-02"),
		"End":          end.Format("2006-01-02"),
		"TotalEntries": totalEntries,
		"TotalTime":    totalTime / 60, // Minutes
		"ByTag":        tagReport,
		"ByType":       typeReport,
		"Entries":      filtered,
	}

	if err := h.templates.RenderPartial(w, "report_results.html", data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
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

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=learnd-export-%s.csv", time.Now().Format("2006-01-02")))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	writer.Write([]string{
		"Date", "URL", "Title", "Type", "Tags", "Time (min)", "Quantity", "Notes", "Summary",
	})

	opts := repository.ListOptions{
		Limit:  pageSize,
		Offset: 0,
		Start:  &start,
		End:    &end,
	}

	for {
		entries, err := h.entryRepo.List(ctx, opts)
		if err != nil {
			http.Error(w, "Failed to get entries", http.StatusInternalServerError)
			return
		}
		if len(entries) == 0 {
			break
		}

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
			if entry.TimeSpentSeconds != nil {
				timeSpent = fmt.Sprintf("%d", *entry.TimeSpentSeconds/60)
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

		opts.Offset += len(entries)
	}
}
