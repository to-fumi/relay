package config_test

import (
	"testing"

	"relay/internal/platform/config"
)

const secret = "0123456789abcdef0123456789abcdef0123"

func setBase(t *testing.T) {
	t.Helper()
	t.Setenv("DATABASE_URL", "postgres://localhost/relay")
	t.Setenv("HANDOFF_TOKEN_SECRET", secret)
	t.Setenv("AUTH_MODE", "")
}

func setOIDC(t *testing.T) {
	t.Helper()
	t.Setenv("AUTH_MODE", "oidc")
	t.Setenv("OIDC_CLIENT_ID", "id")
	t.Setenv("OIDC_CLIENT_SECRET", "secret")
	t.Setenv("OIDC_REDIRECT_URL", "https://relay.example/api/auth/callback")
	t.Setenv("ALLOWED_EMAILS", "a@example.com")
}

func TestLoad_RequiresAuthMode(t *testing.T) {
	setBase(t)
	if _, err := config.Load(); err == nil {
		t.Fatal("AUTH_MODE unset: want error (must not fall back to dev login)")
	}
	t.Setenv("AUTH_MODE", "dev")
	if _, err := config.Load(); err != nil {
		t.Fatalf("AUTH_MODE=dev: %v", err)
	}
}

func TestLoad_OIDC(t *testing.T) {
	setBase(t)
	setOIDC(t)
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.OIDCIssuer != "https://accounts.google.com" || cfg.SessionTTL.Hours() != 24 {
		t.Fatalf("defaults: issuer=%q ttl=%v", cfg.OIDCIssuer, cfg.SessionTTL)
	}

	t.Setenv("ALLOWED_EMAILS", "")
	if _, err := config.Load(); err == nil {
		t.Error("ALLOWED_EMAILS empty: want error")
	}
	t.Setenv("ALLOWED_EMAILS", "a@example.com")

	t.Setenv("HANDOFF_TOKEN_SECRET", "change-me-change-me-change-me-change-me")
	if _, err := config.Load(); err == nil {
		t.Error("example HANDOFF_TOKEN_SECRET: want error")
	}
}
