package token_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"relay/internal/handoff/adapter/token"
	"relay/internal/handoff/application"
	"relay/internal/handoff/domain"
)

var secret = []byte("0123456789abcdef0123456789abcdef")

const base32Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"

func claims(now time.Time) application.TokenClaims {
	due := domain.NewDueDate(2026, 11, 28)
	return application.TokenClaims{
		JTI:       "0123456789abcdef",
		From:      uuid.New(),
		Item:      "LANケーブル",
		Qty:       2,
		DueDate:   &due,
		ExpiresAt: now.Add(time.Minute).Truncate(time.Second),
	}
}

func sameDueDate(a, b *domain.DueDate) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func TestCompactSigner_RoundTrip(t *testing.T) {
	s := token.NewCompactSigner(secret)
	now := time.Now()
	withoutDue := claims(now)
	withoutDue.DueDate = nil

	for name, in := range map[string]application.TokenClaims{"with due date": claims(now), "without due date": withoutDue} {
		t.Run(name, func(t *testing.T) {
			tok, err := s.Sign(in)
			if err != nil {
				t.Fatal(err)
			}
			got, err := s.Verify(tok, now)
			if err != nil {
				t.Fatal(err)
			}
			if !sameDueDate(got.DueDate, in.DueDate) {
				t.Fatalf("DueDate = %v; want %v", got.DueDate, in.DueDate)
			}
			got.DueDate, in.DueDate = nil, nil
			if got != in {
				t.Fatalf("got %+v; want %+v", got, in)
			}
		})
	}
}

func TestCompactSigner_IsShortAndQRAlphanumeric(t *testing.T) {
	tok, _ := token.NewCompactSigner(secret).Sign(claims(time.Now()))
	if len(tok) > 100 {
		t.Errorf("token length = %d; want <= 100", len(tok))
	}
	if strings.Trim(tok, base32Chars) != "" {
		t.Errorf("token %q contains non-alphanumeric characters", tok)
	}
}

func TestCompactSigner_ReturnRoundTrip(t *testing.T) {
	s := token.NewCompactSigner(secret)
	now := time.Now()
	in := application.ReturnClaims{HandoffID: uuid.New(), ExpiresAt: now.Add(time.Minute).Truncate(time.Second)}

	tok, err := s.SignReturn(in)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Trim(tok, base32Chars) != "" || len(tok) > 60 {
		t.Errorf("return token %q should be short and alphanumeric", tok)
	}
	got, err := s.VerifyReturn(tok, now)
	if err != nil {
		t.Fatal(err)
	}
	if got != in {
		t.Fatalf("got %+v; want %+v", got, in)
	}
	if _, err := s.VerifyReturn(tok, now.Add(2*time.Minute)); !errors.Is(err, application.ErrTokenExpired) {
		t.Fatalf("expired: err = %v; want ErrTokenExpired", err)
	}
}

func TestCompactSigner_KindsAreNotInterchangeable(t *testing.T) {
	s := token.NewCompactSigner(secret)
	now := time.Now()
	handoffTok, _ := s.Sign(claims(now))
	returnTok, _ := s.SignReturn(application.ReturnClaims{HandoffID: uuid.New(), ExpiresAt: now.Add(time.Minute)})

	if _, err := s.Verify(returnTok, now); !errors.Is(err, application.ErrReturnToken) {
		t.Errorf("Verify(return token): err = %v; want ErrReturnToken", err)
	}
	if _, err := s.VerifyReturn(handoffTok, now); !errors.Is(err, application.ErrInvalidToken) {
		t.Errorf("VerifyReturn(handoff token): err = %v; want ErrInvalidToken", err)
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
