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
	AuthModeDev   AuthMode = "dev"
	AuthModeEntra AuthMode = "entra"
)

type Config struct {
	Addr               string
	DatabaseURL        string
	StaticDir          string
	AuthMode           AuthMode
	EntraTenantID      string
	EntraAudience      string
	HandoffTokenSecret []byte
	HandoffTokenTTL    time.Duration
	AdminEmails        []string
}

func Load() (Config, error) {
	cfg := Config{
		Addr:               getenv("ADDR", ":8080"),
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		StaticDir:          os.Getenv("STATIC_DIR"),
		AuthMode:           AuthMode(getenv("AUTH_MODE", string(AuthModeDev))),
		EntraTenantID:      os.Getenv("ENTRA_TENANT_ID"),
		EntraAudience:      os.Getenv("ENTRA_AUDIENCE"),
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

	var errs []error
	if cfg.DatabaseURL == "" {
		errs = append(errs, errors.New("DATABASE_URL is required"))
	}
	if len(cfg.HandoffTokenSecret) < 32 {
		errs = append(errs, errors.New("HANDOFF_TOKEN_SECRET must be at least 32 bytes"))
	}
	switch cfg.AuthMode {
	case AuthModeDev:
	case AuthModeEntra:
		if cfg.EntraTenantID == "" || cfg.EntraAudience == "" {
			errs = append(errs, errors.New("ENTRA_TENANT_ID and ENTRA_AUDIENCE are required when AUTH_MODE=entra"))
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
