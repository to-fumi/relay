// Package token は QR トークンを、QR を小さく（荒く）できる短いバイナリ形式で署名・検証する。
//
// 形式（Base32、パディングなし。QR の英数字モードに収まる）:
//
//	version(1) | jti(8) | from(16) | exp(uint32, unix 秒) | qty(uvarint) | item(UTF-8) | mac(12)
//
// mac は HMAC-SHA256(secret, mac より前のバイト列) の先頭 12 バイト。
package token

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base32"
	"encoding/binary"
	"encoding/hex"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"relay/internal/handoff/application"
)

const (
	version  = 1
	jtiLen   = 8
	macLen   = 12
	headLen  = 1 + jtiLen + 16 + 4 // version + jti + from + exp
	minBytes = headLen + 1 + macLen
)

var encoding = base32.StdEncoding.WithPadding(base32.NoPadding)

type CompactSigner struct {
	secret []byte
}

func NewCompactSigner(secret []byte) *CompactSigner {
	return &CompactSigner{secret: secret}
}

func (s *CompactSigner) Sign(c application.TokenClaims) (string, error) {
	jti, err := hex.DecodeString(c.JTI)
	if err != nil || len(jti) != jtiLen {
		return "", application.ErrInvalidToken
	}
	buf := make([]byte, 0, headLen+binary.MaxVarintLen64+len(c.Item)+macLen)
	buf = append(buf, version)
	buf = append(buf, jti...)
	buf = append(buf, c.From[:]...)
	buf = binary.BigEndian.AppendUint32(buf, uint32(c.ExpiresAt.Unix()))
	buf = binary.AppendUvarint(buf, uint64(c.Qty))
	buf = append(buf, c.Item...)
	buf = append(buf, s.mac(buf)...)
	return encoding.EncodeToString(buf), nil
}

func (s *CompactSigner) Verify(token string, now time.Time) (application.TokenClaims, error) {
	invalid := application.TokenClaims{}
	b, err := encoding.DecodeString(strings.ToUpper(strings.TrimSpace(token)))
	if err != nil || len(b) < minBytes || b[0] != version {
		return invalid, application.ErrInvalidToken
	}
	body, sig := b[:len(b)-macLen], b[len(b)-macLen:]
	if !hmac.Equal(sig, s.mac(body)) {
		return invalid, application.ErrInvalidToken
	}

	jti := body[1 : 1+jtiLen]
	from, err := uuid.FromBytes(body[1+jtiLen : 1+jtiLen+16])
	if err != nil {
		return invalid, application.ErrInvalidToken
	}
	exp := time.Unix(int64(binary.BigEndian.Uint32(body[1+jtiLen+16:headLen])), 0)
	qty, n := binary.Uvarint(body[headLen:])
	if n <= 0 || qty > 1<<31 {
		return invalid, application.ErrInvalidToken
	}
	item := body[headLen+n:]
	if !utf8.Valid(item) {
		return invalid, application.ErrInvalidToken
	}
	if !now.Before(exp) {
		return invalid, application.ErrTokenExpired
	}
	return application.TokenClaims{
		JTI:       hex.EncodeToString(jti),
		From:      from,
		Item:      string(item),
		Qty:       int(qty),
		ExpiresAt: exp,
	}, nil
}

func (s *CompactSigner) mac(b []byte) []byte {
	m := hmac.New(sha256.New, s.secret)
	m.Write(b)
	return m.Sum(nil)[:macLen]
}
