package middleware

import (
	"net/http"

	"github.com/drywaters/learnd/internal/session"
)

const cookieName = "learnd_session"

// Auth middleware validates the session token cookie
func Auth(sessions *session.Store, secureCookies bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(cookieName)
			if err != nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			// O(1) token lookup instead of expensive bcrypt comparison
			if !sessions.Valid(cookie.Value) {
			// Invalid/expired session, clear cookie and redirect
			http.SetCookie(w, &http.Cookie{
				Name:     cookieName,
				Value:    "",
				Path:     "/",
				MaxAge:   -1,
				HttpOnly: true,
				Secure:   secureCookies,
				SameSite: http.SameSiteStrictMode,
			})
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			// Refresh session TTL on activity
			sessions.Refresh(cookie.Value)

			next.ServeHTTP(w, r)
		})
	}
}
