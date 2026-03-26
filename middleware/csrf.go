package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
)

// CSRFProtect returns middleware that enforces double-submit cookie CSRF
// protection. State-changing methods (POST, PUT, DELETE) must include an
// X-CSRF-Token header whose value matches the csrf_token cookie.
//
// secureCookie controls whether the cookie is set with the Secure flag
// (should be true in production behind HTTPS).
func CSRFProtect(secureCookie bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip CSRF for OAuth callback and login routes — they use
			// GET + state parameter for protection.
			if strings.HasPrefix(r.URL.Path, "/auth/callback/") ||
				strings.HasPrefix(r.URL.Path, "/auth/login/") {
				next.ServeHTTP(w, r)
				return
			}

			// Ensure the client has a CSRF cookie.
			cookie, err := r.Cookie("csrf_token")
			if err != nil || cookie.Value == "" {
				token, genErr := generateCSRFToken()
				if genErr != nil {
					http.Error(w, "internal server error", http.StatusInternalServerError)
					return
				}
				http.SetCookie(w, &http.Cookie{
					Name:     "csrf_token",
					Value:    token,
					Path:     "/",
					HttpOnly: false, // JS needs to read this
					Secure:   secureCookie,
					SameSite: http.SameSiteStrictMode,
				})
				cookie = &http.Cookie{Value: token}
			}

			// For state-changing methods, validate the token.
			if r.Method == http.MethodPost ||
				r.Method == http.MethodPut ||
				r.Method == http.MethodDelete {
				headerToken := r.Header.Get("X-CSRF-Token")
				if headerToken == "" || headerToken != cookie.Value {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusForbidden)
					json.NewEncoder(w).Encode(map[string]string{
						"error": "CSRF token missing or invalid",
					})
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// generateCSRFToken returns a cryptographically random 32-byte hex string.
func generateCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
