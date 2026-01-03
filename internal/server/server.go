package server

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/danielmerrison/learnd/internal/config"
	"github.com/danielmerrison/learnd/internal/handler"
	"github.com/danielmerrison/learnd/internal/middleware"
	"github.com/danielmerrison/learnd/internal/repository"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

// Server represents the HTTP server
type Server struct {
	cfg              *config.Config
	entryRepo        *repository.EntryRepository
	summaryCacheRepo *repository.SummaryCacheRepository
	templates        handler.TemplateRenderer
}

// New creates a new Server
func New(cfg *config.Config, entryRepo *repository.EntryRepository, summaryCacheRepo *repository.SummaryCacheRepository) *Server {
	// Load templates with custom functions
	funcMap := template.FuncMap{
		"divide": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"formatDuration":    formatDuration,
		"formatReadingTime": formatReadingTime,
		"formatDate":        formatDate,
	}

	templates := newTemplateRenderer(funcMap)

	return &Server{
		cfg:              cfg,
		entryRepo:        entryRepo,
		summaryCacheRepo: summaryCacheRepo,
		templates:        templates,
	}
}

func formatDuration(seconds *int) string {
	if seconds == nil || *seconds <= 0 {
		return ""
	}

	total := *seconds
	hours := total / 3600
	minutes := (total % 3600) / 60
	secs := total % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %02dm", hours, minutes)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %02ds", minutes, secs)
	}
	return fmt.Sprintf("%ds", secs)
}

func formatReadingTime(seconds *int) string {
	if seconds == nil || *seconds <= 0 {
		return ""
	}

	minutes := *seconds / 60
	if *seconds%60 != 0 {
		minutes++
	}

	hours := minutes / 60
	remaining := minutes % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, remaining)
	}
	return fmt.Sprintf("%dm", minutes)
}

func formatDate(t time.Time) string {
	return t.Format("Jan 2, 2006")
}

// Router returns the configured chi router
func (s *Server) Router() http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(middleware.Logger)
	r.Use(chimw.Recoverer)

	// Static files
	fileServer := http.FileServer(http.Dir("static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Auth handlers
	authHandler := handler.NewAuthHandler(s.cfg.APIKeyHash, s.templates)
	r.Get("/login", authHandler.LoginPage)
	r.Post("/login", authHandler.Login)
	r.Post("/logout", authHandler.Logout)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(s.cfg.APIKeyHash))

		// Capture handler
		captureHandler := handler.NewCaptureHandler(s.entryRepo, s.templates)
		r.Get("/", captureHandler.CapturePage)

		// Entry API
		entryHandler := handler.NewEntryHandler(s.entryRepo, s.templates)
		r.Post("/api/entries", entryHandler.Create)
		r.Get("/api/entries", entryHandler.List)
		r.Get("/api/entries/{id}", entryHandler.Get)
		r.Put("/api/entries/{id}", entryHandler.Update)
		r.Delete("/api/entries/{id}", entryHandler.Delete)
		r.Post("/api/entries/{id}/refresh-enrichment", entryHandler.RefreshEnrichment)
		r.Post("/api/entries/{id}/refresh-summary", entryHandler.RefreshSummary)
		r.Get("/entries/{id}/status", entryHandler.Status)

		// Report handler
		reportHandler := handler.NewReportHandler(s.entryRepo, s.templates)
		r.Get("/reports", reportHandler.ReportsPage)
		r.Get("/api/reports", reportHandler.GetReport)
		r.Get("/api/reports/export", reportHandler.ExportCSV)
	})

	return r
}
