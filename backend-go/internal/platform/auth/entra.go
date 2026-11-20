package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
)

// EntraVerifier は Entra ID が発行した v2.0 のアクセストークンを検証する。
type EntraVerifier struct {
	keys     keyfunc.Keyfunc
	issuer   string
	audience string
}

func NewEntraVerifier(ctx context.Context, tenantID, audience string) (*EntraVerifier, error) {
	jwksURL := fmt.Sprintf("https://login.microsoftonline.com/%s/discovery/v2.0/keys", tenantID)
	k, err := keyfunc.NewDefaultCtx(ctx, []string{jwksURL})
	if err != nil {
		return nil, fmt.Errorf("load JWKS: %w", err)
	}
	return &EntraVerifier{
		keys:     k,
		issuer:   fmt.Sprintf("https://login.microsoftonline.com/%s/v2.0", tenantID),
		audience: audience,
	}, nil
}

type entraClaims struct {
	OID               string `json:"oid"`
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
	jwt.RegisteredClaims
}

func (v *EntraVerifier) Verify(_ context.Context, token string) (Principal, error) {
	var c entraClaims
	_, err := jwt.ParseWithClaims(token, &c, v.keys.Keyfunc,
		jwt.WithValidMethods([]string{"RS256"}),
		jwt.WithIssuer(v.issuer),
		jwt.WithAudience(v.audience),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return Principal{}, err
	}
	if c.OID == "" {
		return Principal{}, errors.New("oid claim is missing")
	}
	return Principal{Subject: c.OID, Email: c.PreferredUsername, Name: c.Name}, nil
}
