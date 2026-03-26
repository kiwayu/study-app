package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireAuth_NoCookie(t *testing.T) {
	middleware := RequireAuth(testSecret)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called when no cookie is present")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}

	var body map[string]string
	json.NewDecoder(w.Body).Decode(&body)
	if body["error"] == "" {
		t.Error("expected error message in response body")
	}
}

func TestRequireAuth_InvalidJWT(t *testing.T) {
	middleware := RequireAuth(testSecret)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called with invalid JWT")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "not-a-valid-jwt"})
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestRequireAuth_ValidJWT(t *testing.T) {
	tokenStr, err := GenerateAccessToken("user-abc", "abc@example.com", testSecret)
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}

	var gotUserID, gotEmail string
	middleware := RequireAuth(testSecret)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID = UserIDFromContext(r.Context())
		gotEmail = EmailFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: tokenStr})
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if gotUserID != "user-abc" {
		t.Errorf("expected userID 'user-abc', got %q", gotUserID)
	}
	if gotEmail != "abc@example.com" {
		t.Errorf("expected email 'abc@example.com', got %q", gotEmail)
	}
}
