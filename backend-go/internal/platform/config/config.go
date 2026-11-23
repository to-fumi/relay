// Package config は環境変数からアプリの設定を読み込む。
package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

type AuthMode string

const (
	// AuthModeDev は開発用ログイン（メールアドレスだけでログインする）
	AuthModeDev AuthMode = "dev"
	// AuthModeOIDC は OIDC（Google など）でログインする
	AuthModeOIDC AuthMode = "oidc"
)

type Config struct {
	Addr               string
	DatabaseURL        string
	StaticDir          string
	AuthMode           AuthMode
	OIDCIssuer         string
	OIDCClientID       string
	OIDCClientSecret   string
	OIDCRedirectURL    string
	AllowedEmails      []string
	SessionTTL         time.Duration
	HandoffTokenSecret []byte
	HandoffTokenTTL    time.Duration
	AdminEmails        []string
}

func Load() (Config, error) {
	cfg := Config{
		Addr:        getenv("ADDR", ":8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		StaticDir:   os.Getenv("STATIC_DIR"),
		// 書き忘れて開発用ログイン（誰にでもなりすませる）で動かないよう、初期値は置かない
		AuthMode:           AuthMode(os.Getenv("AUTH_MODE")),
		OIDCIssuer:         getenv("OIDC_ISSUER", "https://accounts.google.com"),
		OIDCClientID:       os.Getenv("OIDC_CLIENT_ID"),
		OIDCClientSecret:   os.Getenv("OIDC_CLIENT_SECRET"),
		OIDCRedirectURL:    os.Getenv("OIDC_REDIRECT_URL"),
		AllowedEmails:      splitList(os.Getenv("ALLOWED_EMAILS")),
		SessionTTL:         24 * time.Hour,
		HandoffTokenSecret: []byte(os.Getenv("HANDOFF_TOKEN_SECRET")),
		HandoffTokenTTL:    60 * time.Second,
		AdminEmails:        splitList(os.Getenv("ADMIN_EMAILS")),
	}

	if ttl := os.Getenv("HANDOFF_TOKEN_TTL"); ttl != "" {
		d, err := time.ParseDuration(ttl)
		if err != nil {
			return Config{}, fmt.Errorf("HANDOFF_TOKEN_TTL: %w", err)
		}
		cfg.HandoffTokenTTL = d
	}

	if ttl := os.Getenv("SESSION_TTL"); ttl != "" {
		d, err := time.ParseDuration(ttl)
		if err != nil {
			return Config{}, fmt.Errorf("SESSION_TTL: %w", err)
		}
		cfg.SessionTTL = d
	}

	var errs []error
	if cfg.DatabaseURL == "" {
		errs = append(errs, errors.New("DATABASE_URL is required"))
	}
	if len(cfg.HandoffTokenSecret) < 32 {
		errs = append(errs, errors.New("HANDOFF_TOKEN_SECRET must be at least 32 bytes"))
	}
	switch cfg.AuthMode {
	case "":
		errs = append(errs, errors.New("AUTH_MODE is required (dev or oidc)"))
	case AuthModeDev:
	case AuthModeOIDC:
		if cfg.OIDCClientID == "" || cfg.OIDCClientSecret == "" || cfg.OIDCRedirectURL == "" {
			errs = append(errs, errors.New("OIDC_CLIENT_ID, OIDC_CLIENT_SECRET and OIDC_REDIRECT_URL are required when AUTH_MODE=oidc"))
		}
		// 誰でもログインできる状態にしない
		if len(cfg.AllowedEmails) == 0 {
			errs = append(errs, errors.New("ALLOWED_EMAILS is required when AUTH_MODE=oidc"))
		}
		// .env.example の値のまま本番で動かさない（QR トークンを誰でも作れてしまう）
		if strings.Contains(string(cfg.HandoffTokenSecret), "change-me") {
			errs = append(errs, errors.New("HANDOFF_TOKEN_SECRET must be changed from the example value"))
		}
	default:
		errs = append(errs, fmt.Errorf("unknown AUTH_MODE %q", cfg.AuthMode))
	}
	return cfg, errors.Join(errs...)
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func splitList(s string) []string {
	var out []string
	for v := range strings.SplitSeq(s, ",") {
		if v = strings.TrimSpace(v); v != "" {
			out = append(out, v)
		}
	}
	return out
}
