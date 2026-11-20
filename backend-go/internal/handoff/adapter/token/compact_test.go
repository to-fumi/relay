package token_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"relay/internal/handoff/adapter/token"
	"relay/internal/handoff/application"
)

var secret = []byte("0123456789abcdef0123456789abcdef")

func claims(now time.Time) application.TokenClaims {
	return application.TokenClaims{
		JTI:       "0123456789abcdef",
		From:      uuid.New(),
		Item:      "LANケーブル",
		Qty:       2,
		ExpiresAt: now.Add(time.Minute).Truncate(time.Second),
	}
}

func TestCompactSigner_RoundTrip(t *testing.T) {
	s := token.NewCompactSigner(secret)
	now := time.Now()
	in := claims(now)

	tok, err := s.Sign(in)
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.Verify(tok, now)
	if err != nil {
		t.Fatal(err)
	}
	if got != in {
		t.Fatalf("got %+v; want %+v", got, in)
	}
}

func TestCompactSigner_IsShortAndQRAlphanumeric(t *testing.T) {
	tok, _ := token.NewCompactSigner(secret).Sign(claims(time.Now()))
	if len(tok) > 100 {
		t.Errorf("token length = %d; want <= 100", len(tok))
	}
	if strings.Trim(tok, "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567") != "" {
		t.Errorf("token %q contains non-alphanumeric characters", tok)
	}
}

func TestCompactSigner_Expired(t *testing.T) {
	s := token.NewCompactSigner(secret)
	now := time.Now()
	tok, _ := s.Sign(claims(now))

	if _, err := s.Verify(tok, now.Add(2*time.Minute)); !errors.Is(err, application.ErrTokenExpired) {
		t.Fatalf("err = %v; want ErrTokenExpired", err)
	}
}

func TestCompactSigner_Rejects(t *testing.T) {
	now := time.Now()
	tok, _ := token.NewCompactSigner(secret).Sign(claims(now))
	flipped := []byte(tok)
	flipped[10] = map[bool]byte{true: 'B', false: 'A'}[flipped[10] == 'A']

	tests := map[string]struct {
		signer *token.CompactSigner
		token  string
	}{
		"wrong secret": {token.NewCompactSigner([]byte("ffffffffffffffffffffffffffffffff")), tok},
		"tampered":     {token.NewCompactSigner(secret), string(flipped)},
		"truncated":    {token.NewCompactSigner(secret), tok[:20]},
		"garbage":      {token.NewCompactSigner(secret), "not-a-token"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := tt.signer.Verify(tt.token, now); !errors.Is(err, application.ErrInvalidToken) {
				t.Fatalf("err = %v; want ErrInvalidToken", err)
			}
		})
	}
}
