package handler

import "net/http"

// TemplateRenderer defines template rendering used by handlers.
type TemplateRenderer interface {
	RenderPage(w http.ResponseWriter, name string, data any) error
	RenderPartial(w http.ResponseWriter, name string, data any) error
}
