package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/oauth2"

	"studysession/db"
	"studysession/models"
)

// AuthHandler holds dependencies for OAuth route handlers.
type AuthHandler struct {
	googleOAuth *oauth2.Config
	gitHubOAuth *oauth2.Config
	jwtSecret   string
	baseURL     string
	env         string
	userRepo    *db.UserRepo
	tokenRepo   *db.TokenRepo
}

// NewAuthHandler constructs an AuthHandler.
func NewAuthHandler(
	googleOAuth *oauth2.Config,
	gitHubOAuth *oauth2.Config,
	jwtSecret string,
	baseURL string,
	env string,
	userRepo *db.UserRepo,
	tokenRepo *db.TokenRepo,
) *AuthHandler {
	return &AuthHandler{
		googleOAuth: googleOAuth,
		gitHubOAuth: gitHubOAuth,
		jwtSecret:   jwtSecret,
		baseURL:     baseURL,
		env:         env,
		userRepo:    userRepo,
		tokenRepo:   tokenRepo,
	}
}

// secureCookie returns true when cookies should be marked Secure.
func (h *AuthHandler) secureCookie() bool {
	return h.env == "production"
}

// LoginGoogle redirects the user to Google's OAuth2 consent screen.
func (h *AuthHandler) LoginGoogle(w http.ResponseWriter, r *http.Request) {
	state, err := randomState()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to generate state")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   300, // 5 minutes
		HttpOnly: true,
		Secure:   h.secureCookie(),
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, h.googleOAuth.AuthCodeURL(state), http.StatusTemporaryRedirect)
}

// LoginGitHub redirects the user to GitHub's OAuth2 consent screen.
func (h *AuthHandler) LoginGitHub(w http.ResponseWriter, r *http.Request) {
	state, err := randomState()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to generate state")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   300,
		HttpOnly: true,
		Secure:   h.secureCookie(),
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, h.gitHubOAuth.AuthCodeURL(state), http.StatusTemporaryRedirect)
}

// CallbackGoogle handles the OAuth2 callback from Google.
func (h *AuthHandler) CallbackGoogle(w http.ResponseWriter, r *http.Request) {
	if err := h.validateState(r); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid oauth state")
		return
	}

	code := r.URL.Query().Get("code")
	token, err := h.googleOAuth.Exchange(r.Context(), code)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "failed to exchange code")
		return
	}

	client := h.googleOAuth.Client(r.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to fetch user info")
		return
	}
	defer resp.Body.Close()

	var info struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to decode user info")
		return
	}

	user, err := h.findOrCreateUser(r, "google", info.ID, info.Email, info.Name, info.Picture)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to find or create user")
		return
	}

	h.issueTokensAndRedirect(w, r, user)
}

// CallbackGitHub handles the OAuth2 callback from GitHub.
func (h *AuthHandler) CallbackGitHub(w http.ResponseWriter, r *http.Request) {
	if err := h.validateState(r); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid oauth state")
		return
	}

	code := r.URL.Query().Get("code")
	token, err := h.gitHubOAuth.Exchange(r.Context(), code)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "failed to exchange code")
		return
	}

	client := h.gitHubOAuth.Client(r.Context(), token)

	// Fetch user profile.
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to fetch GitHub user")
		return
	}
	defer resp.Body.Close()

	var ghUser struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&ghUser); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to decode GitHub user")
		return
	}

	email := ghUser.Email
	if email == "" {
		email, err = h.fetchGitHubPrimaryEmail(r, client)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to fetch GitHub email")
			return
		}
	}

	name := ghUser.Name
	if name == "" {
		name = ghUser.Login
	}

	providerID := fmt.Sprintf("%d", ghUser.ID)
	user, err := h.findOrCreateUser(r, "github", providerID, email, name, ghUser.AvatarURL)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to find or create user")
		return
	}

	h.issueTokensAndRedirect(w, r, user)
}

// Refresh issues a new access token using a valid refresh token cookie.
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "missing refresh token")
		return
	}

	hash := hashToken(cookie.Value)
	stored, err := h.tokenRepo.FindByHash(r.Context(), hash)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "token lookup failed")
		return
	}
	if stored == nil || time.Now().After(stored.ExpiresAt) {
		writeJSONError(w, http.StatusUnauthorized, "invalid or expired refresh token")
		return
	}

	// Look up user to get email for new access token.
	user, err := h.userRepo.GetByID(r.Context(), stored.UserID)
	if err != nil || user == nil {
		writeJSONError(w, http.StatusUnauthorized, "user not found")
		return
	}

	accessToken, err := GenerateAccessToken(user.ID, user.Email, h.jwtSecret)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to generate access token")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		MaxAge:   900, // 15 minutes
		HttpOnly: true,
		Secure:   h.secureCookie(),
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// Logout clears tokens from the database and cookies.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("refresh_token"); err == nil {
		hash := hashToken(cookie.Value)
		if stored, err := h.tokenRepo.FindByHash(r.Context(), hash); err == nil && stored != nil {
			_ = h.tokenRepo.Delete(r.Context(), stored.ID)
		}
	}

	// Clear cookies.
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.secureCookie(),
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.secureCookie(),
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// Me returns the authenticated user's profile.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		writeJSONError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	user, err := h.userRepo.GetByID(r.Context(), userID)
	if err != nil || user == nil {
		writeJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":        user.ID,
		"email":     user.Email,
		"name":      user.Name,
		"avatarUrl": user.AvatarURL,
		"provider":  user.Provider,
		"createdAt": user.CreatedAt,
	})
}

// --- helpers ---

func randomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func (h *AuthHandler) validateState(r *http.Request) error {
	cookie, err := r.Cookie("oauth_state")
	if err != nil {
		return fmt.Errorf("missing state cookie")
	}
	if r.URL.Query().Get("state") != cookie.Value {
		return fmt.Errorf("state mismatch")
	}
	return nil
}

func (h *AuthHandler) findOrCreateUser(r *http.Request, provider, providerID, email, name, avatarURL string) (*models.User, error) {
	user, err := h.userRepo.FindByProviderID(r.Context(), provider, providerID)
	if err != nil {
		return nil, err
	}
	if user != nil {
		return user, nil
	}

	user = &models.User{
		Email:      email,
		Name:       name,
		AvatarURL:  avatarURL,
		Provider:   provider,
		ProviderID: providerID,
	}
	return h.userRepo.Create(r.Context(), user)
}

func (h *AuthHandler) issueTokensAndRedirect(w http.ResponseWriter, r *http.Request, user *models.User) {
	accessToken, err := GenerateAccessToken(user.ID, user.Email, h.jwtSecret)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to generate access token")
		return
	}

	refreshToken, err := GenerateRefreshToken()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to generate refresh token")
		return
	}

	// Store hashed refresh token in DB.
	hash := hashToken(refreshToken)
	_, err = h.tokenRepo.Create(r.Context(), &models.RefreshToken{
		UserID:    user.ID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	})
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to store refresh token")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		MaxAge:   900, // 15 minutes
		HttpOnly: true,
		Secure:   h.secureCookie(),
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		MaxAge:   604800, // 7 days
		HttpOnly: true,
		Secure:   h.secureCookie(),
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (h *AuthHandler) fetchGitHubPrimaryEmail(r *http.Request, client *http.Client) (string, error) {
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.Unmarshal(body, &emails); err != nil {
		return "", err
	}

	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}
	// Fallback: return first verified email.
	for _, e := range emails {
		if e.Verified {
			return e.Email, nil
		}
	}
	return "", fmt.Errorf("no verified email found")
}
