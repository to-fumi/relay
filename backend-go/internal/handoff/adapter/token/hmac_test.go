package token_test

import (
	"errors"
	"testing"
	"time"

	"relay/internal/handoff/adapter/token"
	"relay/internal/handoff/application"

	"github.com/google/uuid"
)

var secret = []byte("0123456789abcdef0123456789abcdef")

func TestHMACSigner_RoundTrip(t *testing.T) {
	s := token.NewHMACSigner(secret)
	now := time.Now()
	in := application.TokenClaims{JTI: "j1", From: uuid.New(), Item: "ケーブル", Qty: 2, ExpiresAt: now.Add(time.Minute)}

	tok, err := s.Sign(in)
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.Verify(tok, now)
	if err != nil {
		t.Fatal(err)
	}
	if got.JTI != in.JTI || got.From != in.From || got.Item != in.Item || got.Qty != in.Qty {
		t.Fatalf("got %+v; want %+v", got, in)
	}
}

func TestHMACSigner_Expired(t *testing.T) {
	s := token.NewHMACSigner(secret)
	now := time.Now()
	tok, _ := s.Sign(application.TokenClaims{JTI: "j1", From: uuid.New(), Item: "x", Qty: 1, ExpiresAt: now.Add(time.Minute)})

	if _, err := s.Verify(tok, now.Add(2*time.Minute)); !errors.Is(err, application.ErrTokenExpired) {
		t.Fatalf("err = %v; want ErrTokenExpired", err)
	}
}

func TestHMACSigner_WrongSecret(t *testing.T) {
	now := time.Now()
	tok, _ := token.NewHMACSigner(secret).Sign(application.TokenClaims{JTI: "j1", From: uuid.New(), Item: "x", Qty: 1, ExpiresAt: now.Add(time.Minute)})

	other := token.NewHMACSigner([]byte("ffffffffffffffffffffffffffffffff"))
	if _, err := other.Verify(tok, now); !errors.Is(err, application.ErrInvalidToken) {
		t.Fatalf("err = %v; want ErrInvalidToken", err)
	}
}
