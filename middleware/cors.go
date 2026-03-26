package middleware

import (
	"net/http"
	"strings"
)

// CORS returns middleware that sets CORS headers. Requests whose Origin header
// matches any of the supplied allowedOrigins (or starts with any of them) are
// echoed back; all others get no Access-Control-Allow-Origin header.
//
// Pass an empty slice to allow all localhost origins (desktop mode default).
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if origin != "" {
				allowed := false
				if len(allowedOrigins) == 0 {
					// Desktop mode: allow any localhost origin.
					if strings.HasPrefix(origin, "http://127.0.0.1") ||
						strings.HasPrefix(origin, "http://localhost") {
						allowed = true
					}
				} else {
					for _, ao := range allowedOrigins {
						if origin == ao || strings.HasPrefix(origin, ao) {
							allowed = true
							break
						}
					}
				}
				if allowed {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-CSRF-Token")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
