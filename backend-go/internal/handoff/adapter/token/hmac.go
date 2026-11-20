// Package token は QR トークンを HMAC-SHA256 の JWT として署名・検証する。
package token

import (
	"errors"
	"time"

	"relay/internal/handoff/application"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const audience = "relay-handoff"

type HMACSigner struct {
	secret []byte
}

func NewHMACSigner(secret []byte) *HMACSigner {
	return &HMACSigner{secret: secret}
}

type claims struct {
	Item string `json:"item"`
	Qty  int    `json:"qty"`
	jwt.RegisteredClaims
}

func (s *HMACSigner) Sign(c application.TokenClaims) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		Item: c.Item,
		Qty:  c.Qty,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        c.JTI,
			Subject:   c.From.String(),
			Audience:  jwt.ClaimStrings{audience},
			ExpiresAt: jwt.NewNumericDate(c.ExpiresAt),
		},
	})
	return t.SignedString(s.secret)
}

func (s *HMACSigner) Verify(token string, now time.Time) (application.TokenClaims, error) {
	var c claims
	_, err := jwt.ParseWithClaims(token, &c, func(*jwt.Token) (any, error) { return s.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithAudience(audience),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(func() time.Time { return now }),
	)
	if errors.Is(err, jwt.ErrTokenExpired) {
		return application.TokenClaims{}, application.ErrTokenExpired
	}
	if err != nil || c.ID == "" {
		return application.TokenClaims{}, application.ErrInvalidToken
	}
	from, err := uuid.Parse(c.Subject)
	if err != nil {
		return application.TokenClaims{}, application.ErrInvalidToken
	}
	return application.TokenClaims{
		JTI:       c.ID,
		From:      from,
		Item:      c.Item,
		Qty:       c.Qty,
		ExpiresAt: c.ExpiresAt.Time,
	}, nil
}
