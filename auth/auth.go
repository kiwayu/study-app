package auth

import "net/http"

// RegisterRoutes registers all authentication routes on the given mux.
func RegisterRoutes(mux *http.ServeMux, h *AuthHandler) {
	mux.HandleFunc("GET /auth/login/google", h.LoginGoogle)
	mux.HandleFunc("GET /auth/login/github", h.LoginGitHub)
	mux.HandleFunc("GET /auth/callback/google", h.CallbackGoogle)
	mux.HandleFunc("GET /auth/callback/github", h.CallbackGitHub)
	mux.HandleFunc("POST /auth/refresh", h.Refresh)
	mux.HandleFunc("POST /auth/logout", h.Logout)

	// /api/me requires authentication.
	authMiddleware := RequireAuth(h.jwtSecret)
	mux.Handle("GET /api/me", authMiddleware(http.HandlerFunc(h.Me)))
}
