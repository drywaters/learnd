package handler

import (
	"net/http"

	"github.com/drywaters/learnd/internal/session"
	"github.com/drywaters/learnd/internal/ui/pages"
	"golang.org/x/crypto/bcrypt"
)

const cookieName = "learnd_session"

// AuthHandler handles authentication
type AuthHandler struct {
	apiKeyHash    string
	sessions      *session.Store
	secureCookies bool
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(apiKeyHash string, sessions *session.Store, secureCookies bool) *AuthHandler {
	return &AuthHandler{
		apiKeyHash:    apiKeyHash,
		sessions:      sessions,
		secureCookies: secureCookies,
	}
}

// LoginPage renders the login page
func (h *AuthHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	// If already authenticated via valid session, redirect to home
	if cookie, err := r.Cookie(cookieName); err == nil {
		if h.sessions.Valid(cookie.Value) {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	}

	errorType := r.URL.Query().Get("error")
	pages.LoginPage(errorType).Render(r.Context(), w)
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

	// Validate the API key with bcrypt (only happens once at login)
	if err := bcrypt.CompareHashAndPassword([]byte(h.apiKeyHash), []byte(apiKey)); err != nil {
		http.Redirect(w, r, "/login?error=invalid_key", http.StatusSeeOther)
		return
	}

	// Create a session token
	token, err := h.sessions.Create()
	if err != nil {
		http.Redirect(w, r, "/login?error=server_error", http.StatusSeeOther)
		return
	}

	// Set the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.secureCookies,
		SameSite: http.SameSiteStrictMode,
		// No MaxAge = session cookie (expires when browser closes)
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Logout clears the session cookie and invalidates the session
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Invalidate the session server-side
	if cookie, err := r.Cookie(cookieName); err == nil {
		h.sessions.Delete(cookie.Value)
	}

	// Clear the cookie
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
