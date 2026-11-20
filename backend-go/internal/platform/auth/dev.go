package auth

import (
	"context"
	"errors"
	"strings"
)

// DevVerifier は開発用。"dev:<email>" をそのまま利用者として受け入れる。
type DevVerifier struct{}

func (DevVerifier) Verify(_ context.Context, token string) (Principal, error) {
	email, ok := strings.CutPrefix(token, "dev:")
	if !ok || !strings.Contains(email, "@") {
		return Principal{}, errors.New("invalid dev token")
	}
	name, _, _ := strings.Cut(email, "@")
	return Principal{Subject: token, Email: email, Name: name}, nil
}
