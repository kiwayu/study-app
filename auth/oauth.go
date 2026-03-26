package auth

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"studysession/config"
)

// NewGoogleOAuthConfig returns an oauth2.Config for Google sign-in.
func NewGoogleOAuthConfig(cfg *config.Config) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.BaseURL + "/auth/callback/google",
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

// GitHubEndpoint is the OAuth2 endpoint for GitHub.
var GitHubEndpoint = oauth2.Endpoint{
	AuthURL:  "https://github.com/login/oauth/authorize",
	TokenURL: "https://github.com/login/oauth/access_token",
}

// NewGitHubOAuthConfig returns an oauth2.Config for GitHub sign-in.
func NewGitHubOAuthConfig(cfg *config.Config) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     cfg.GitHubClientID,
		ClientSecret: cfg.GitHubClientSecret,
		RedirectURL:  cfg.BaseURL + "/auth/callback/github",
		Scopes:       []string{"user:email"},
		Endpoint:     GitHubEndpoint,
	}
}
