package middleware

import (
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

const cookieName = "learnd_api_key"

// Auth middleware validates the API key cookie
func Auth(apiKeyHash string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(cookieName)
			if err != nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			if err := bcrypt.CompareHashAndPassword([]byte(apiKeyHash), []byte(cookie.Value)); err != nil {
				// Invalid cookie, clear it and redirect
				http.SetCookie(w, &http.Cookie{
					Name:     cookieName,
					Value:    "",
					Path:     "/",
					MaxAge:   -1,
					HttpOnly: true,
				})
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
