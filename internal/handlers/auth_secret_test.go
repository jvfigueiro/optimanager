package handlers_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"optimanager/internal/database"
	"optimanager/internal/handlers"
)

func TestResolveJWTSecret_MissingFailsWhenNotDev(t *testing.T) {
	t.Setenv("JWT_SECRET", "")
	t.Setenv("APP_ENV", "")

	sec, err := handlers.ResolveJWTSecret()
	if err == nil {
		t.Fatalf("expected error when JWT_SECRET is empty and APP_ENV is empty, got secret: %s", string(sec))
	}
	if !strings.Contains(err.Error(), "JWT_SECRET environment variable is missing") {
		t.Fatalf("expected missing env error, got: %v", err)
	}

	// Also test with APP_ENV=production
	t.Setenv("APP_ENV", "production")
	_, err = handlers.ResolveJWTSecret()
	if err == nil {
		t.Fatalf("expected error when JWT_SECRET is empty and APP_ENV=production")
	}
}

func TestResolveJWTSecret_ShortSecretFails(t *testing.T) {
	shortKeys := []string{
		"short",
		"1234567890123456789012345678901", // 31 chars
		"too_short_key_under_32_chars!",
	}

	for _, k := range shortKeys {
		t.Run("len_"+k, func(t *testing.T) {
			t.Setenv("JWT_SECRET", k)
			t.Setenv("APP_ENV", "")
			_, err := handlers.ResolveJWTSecret()
			if err == nil {
				t.Fatalf("expected error for short key %q (len %d)", k, len(k))
			}
			if !strings.Contains(err.Error(), "at least 32 characters long") {
				t.Fatalf("expected short key error message, got: %v", err)
			}
		})
	}

	// Even in APP_ENV=dev, an explicitly provided short secret must fail
	t.Setenv("APP_ENV", "dev")
	t.Setenv("JWT_SECRET", "too_short")
	if _, err := handlers.ResolveJWTSecret(); err == nil {
		t.Fatalf("expected error when short key is provided even in APP_ENV=dev")
	}
}

func TestResolveJWTSecret_OldFallbackFails(t *testing.T) {
	t.Setenv("JWT_SECRET", "optimanager-default-secret-key-2026-hmac-sha256")
	t.Setenv("APP_ENV", "")

	_, err := handlers.ResolveJWTSecret()
	if err == nil {
		t.Fatalf("expected error when JWT_SECRET is the old insecure fallback key")
	}
	if !strings.Contains(err.Error(), "deprecated insecure default fallback key") {
		t.Fatalf("expected old fallback error message, got: %v", err)
	}
}

func TestResolveJWTSecret_PlaceholderFails(t *testing.T) {
	placeholders := []string{
		"CONFIGURE_UMA_CHAVE_COM_PELO_MENOS_32_CHARS_AQUI",
		"CONFIGURE_SUA_CHAVE_SUPER_SECRETA_AQUI_123456",
	}

	for _, ph := range placeholders {
		t.Run(ph, func(t *testing.T) {
			t.Setenv("JWT_SECRET", ph)
			t.Setenv("APP_ENV", "")
			_, err := handlers.ResolveJWTSecret()
			if err == nil {
				t.Fatalf("expected error when JWT_SECRET is placeholder %q", ph)
			}
			if !strings.Contains(err.Error(), "placeholder example key") {
				t.Fatalf("expected placeholder error message, got: %v", err)
			}
		})
	}
}

func TestResolveJWTSecret_ValidSecretSucceeds(t *testing.T) {
	validKey := "e7f3c4a1b8d2e9f0a5b6c7d8e9f01234567890abcdef1234567890abcdef12"
	t.Setenv("JWT_SECRET", validKey)
	t.Setenv("APP_ENV", "production")

	sec, err := handlers.ResolveJWTSecret()
	if err != nil {
		t.Fatalf("expected success with valid 64-char key, got error: %v", err)
	}
	if string(sec) != validKey {
		t.Fatalf("expected secret %q, got %q", validKey, string(sec))
	}
}

func TestResolveJWTSecret_DevModeGeneratesRandomEphemeral(t *testing.T) {
	t.Setenv("JWT_SECRET", "")
	t.Setenv("APP_ENV", "dev")

	sec1, err1 := handlers.ResolveJWTSecret()
	if err1 != nil {
		t.Fatalf("expected dev mode to generate ephemeral secret, got: %v", err1)
	}
	if len(sec1) != 32 {
		t.Fatalf("expected 32-byte secret, got %d bytes", len(sec1))
	}

	sec2, err2 := handlers.ResolveJWTSecret()
	if err2 != nil {
		t.Fatalf("expected second call in dev mode to succeed, got: %v", err2)
	}
	if len(sec2) != 32 {
		t.Fatalf("expected 32-byte secret, got %d bytes", len(sec2))
	}

	if bytes.Equal(sec1, sec2) {
		t.Fatalf("expected two ephemeral calls to produce different cryptographically random secrets")
	}
}

func TestToken_OldFallbackSecretTokenRejected(t *testing.T) {
	// 1. Generate a token using the deprecated fallback secret
	oldSecret := []byte("optimanager-default-secret-key-2026-hmac-sha256")
	payload := map[string]interface{}{
		"uid":   1,
		"email": "admin@example.com",
		"role":  "admin",
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
	}
	payloadBytes, _ := json.Marshal(payload)
	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadBytes)

	macOld := hmac.New(sha256.New, oldSecret)
	macOld.Write([]byte(encodedPayload))
	oldSig := base64.RawURLEncoding.EncodeToString(macOld.Sum(nil))
	oldToken := encodedPayload + "." + oldSig

	// 2. Initialize database and an AuthHandler with a modern valid secret
	testDBPath := "./test_auth_secret.db"
	db, err := database.New(testDBPath)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	defer func() {
		db.Close()
		_ = os.Remove(testDBPath)
		_ = os.Remove(testDBPath + "-wal")
		_ = os.Remove(testDBPath + "-shm")
	}()

	validNewSecret := []byte("super-secure-production-secret-key-with-at-least-32-chars-long!")
	authHandler := handlers.NewAuthHandlerWithSecret(db, validNewSecret)

	// 3. Set up a protected route requiring authentication
	r := chi.NewRouter()
	r.Use(authHandler.RequireAuth)
	r.Get("/api/protected", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// 4. Send request with the old token
	req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
	req.Header.Set("Authorization", "Bearer "+oldToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// 5. Must be rejected with 401 Unauthorized
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized when presenting token signed with old secret, got %d: %s", w.Code, w.Body.String())
	}

	// 6. Verify that a valid token signed with the new secret IS accepted
	// Let's create an admin user in the db first
	_, err = db.Writer.Exec("INSERT OR IGNORE INTO usuarios (id, nome, email, senha_hash, role) VALUES (1, 'Admin', 'admin@example.com', 'hash', 'admin')")
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	// Sign a valid token with validNewSecret
	macNew := hmac.New(sha256.New, validNewSecret)
	macNew.Write([]byte(encodedPayload))
	newSig := base64.RawURLEncoding.EncodeToString(macNew.Sum(nil))
	validToken := encodedPayload + "." + newSig

	reqValid := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
	reqValid.Header.Set("Authorization", "Bearer "+validToken)
	wValid := httptest.NewRecorder()
	r.ServeHTTP(wValid, reqValid)

	if wValid.Code != http.StatusOK {
		t.Fatalf("expected 200 OK with valid new token, got %d: %s", wValid.Code, wValid.Body.String())
	}
}
