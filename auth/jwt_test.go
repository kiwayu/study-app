package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret-key-for-jwt-tests"

func TestGenerateAccessToken_ValidClaims(t *testing.T) {
	tokenStr, err := GenerateAccessToken("user-123", "user@example.com", testSecret)
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}
	if tokenStr == "" {
		t.Fatal("expected non-empty token string")
	}

	// Parse and verify claims.
	claims, err := ValidateAccessToken(tokenStr, testSecret)
	if err != nil {
		t.Fatalf("ValidateAccessToken returned error: %v", err)
	}
	if claims.Subject != "user-123" {
		t.Errorf("expected subject 'user-123', got %q", claims.Subject)
	}
	if claims.Email != "user@example.com" {
		t.Errorf("expected email 'user@example.com', got %q", claims.Email)
	}
	if claims.ExpiresAt == nil {
		t.Fatal("expected ExpiresAt to be set")
	}
	// Expiry should be ~15 minutes from now.
	expiry := claims.ExpiresAt.Time
	if expiry.Before(time.Now().Add(14*time.Minute)) || expiry.After(time.Now().Add(16*time.Minute)) {
		t.Errorf("expiry %v not within expected 15-minute window", expiry)
	}
}

func TestValidateAccessToken_Valid(t *testing.T) {
	tokenStr, err := GenerateAccessToken("user-456", "test@test.com", testSecret)
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}

	claims, err := ValidateAccessToken(tokenStr, testSecret)
	if err != nil {
		t.Fatalf("ValidateAccessToken returned error: %v", err)
	}
	if claims.Subject != "user-456" {
		t.Errorf("expected subject 'user-456', got %q", claims.Subject)
	}
	if claims.Email != "test@test.com" {
		t.Errorf("expected email 'test@test.com', got %q", claims.Email)
	}
}

func TestValidateAccessToken_Expired(t *testing.T) {
	// Create a token that is already expired.
	now := time.Now()
	claims := Claims{
		Email: "expired@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-expired",
			IssuedAt:  jwt.NewNumericDate(now.Add(-1 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(now.Add(-30 * time.Minute)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("failed to sign expired token: %v", err)
	}

	_, err = ValidateAccessToken(tokenStr, testSecret)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}

func TestValidateAccessToken_WrongSecret(t *testing.T) {
	tokenStr, err := GenerateAccessToken("user-789", "wrong@secret.com", testSecret)
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}

	_, err = ValidateAccessToken(tokenStr, "wrong-secret")
	if err == nil {
		t.Fatal("expected error for wrong secret, got nil")
	}
}

func TestGenerateRefreshToken_Format(t *testing.T) {
	token, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken returned error: %v", err)
	}
	if len(token) != 64 {
		t.Errorf("expected 64-char hex string, got length %d", len(token))
	}
	// Verify it's valid hex.
	for _, c := range token {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("unexpected character %c in refresh token", c)
			break
		}
	}

	// Ensure uniqueness (two calls produce different tokens).
	token2, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken (2nd call): %v", err)
	}
	if token == token2 {
		t.Error("two refresh tokens should not be identical")
	}
}
