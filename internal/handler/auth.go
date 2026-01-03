package handler

import (
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

const cookieName = "learnd_api_key"

// AuthHandler handles authentication
type AuthHandler struct {
	apiKeyHash string
	templates  TemplateRenderer
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(apiKeyHash string, templates TemplateRenderer) *AuthHandler {
	return &AuthHandler{
		apiKeyHash: apiKeyHash,
		templates:  templates,
	}
}

// LoginPage renders the login page
func (h *AuthHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	// If already authenticated, redirect to home
	if cookie, err := r.Cookie(cookieName); err == nil {
		if bcrypt.CompareHashAndPassword([]byte(h.apiKeyHash), []byte(cookie.Value)) == nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	}

	data := map[string]interface{}{
		"Error": r.URL.Query().Get("error"),
	}

	if err := h.templates.RenderPage(w, "login.html", data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

// Login handles the login form submission
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/login?error=invalid_request", http.StatusSeeOther)
		return
	}

	apiKey := r.FormValue("api_key")
	if apiKey == "" {
		http.Redirect(w, r, "/login?error=missing_key", http.StatusSeeOther)
		return
	}

	// Validate the API key
	if err := bcrypt.CompareHashAndPassword([]byte(h.apiKeyHash), []byte(apiKey)); err != nil {
		http.Redirect(w, r, "/login?error=invalid_key", http.StatusSeeOther)
		return
	}

	// Set the cookie with the raw API key (will be validated against hash on each request)
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    apiKey,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		// No MaxAge = session cookie (expires when browser closes)
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Logout clears the session cookie
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
