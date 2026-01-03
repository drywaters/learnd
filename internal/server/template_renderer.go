package server

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/danielmerrison/learnd/internal/handler"
)

type templateRenderer struct {
	pages    map[string]*template.Template
	partials *template.Template
}

func newTemplateRenderer(funcMap template.FuncMap) handler.TemplateRenderer {
	base := template.New("").Funcs(funcMap)

	if _, err := base.ParseGlob("templates/layouts/*.html"); err != nil {
		slog.Warn("failed to parse layout templates", "error", err)
	}

	if _, err := base.ParseGlob("templates/partials/*.html"); err != nil {
		slog.Warn("failed to parse partial templates", "error", err)
	}

	pages := map[string]*template.Template{}
	pageFiles, err := filepath.Glob("templates/pages/*.html")
	if err != nil {
		slog.Warn("failed to list page templates", "error", err)
	} else {
		for _, file := range pageFiles {
			clone, err := base.Clone()
			if err != nil {
				slog.Warn("failed to clone templates for page", "page", file, "error", err)
				continue
			}

			if _, err := clone.ParseFiles(file); err != nil {
				slog.Warn("failed to parse page template", "page", file, "error", err)
				continue
			}

			pages[filepath.Base(file)] = clone
		}
	}

	partials, err := base.Clone()
	if err != nil {
		slog.Warn("failed to clone partial templates", "error", err)
		partials = template.New("").Funcs(funcMap)
	}

	return &templateRenderer{
		pages:    pages,
		partials: partials,
	}
}

func (r *templateRenderer) RenderPage(w http.ResponseWriter, name string, data any) error {
	tmpl, ok := r.pages[name]
	if !ok {
		return fmt.Errorf("page template not found: %s", name)
	}

	return tmpl.ExecuteTemplate(w, name, data)
}

func (r *templateRenderer) RenderPartial(w http.ResponseWriter, name string, data any) error {
	return r.partials.ExecuteTemplate(w, name, data)
}
