// Package token は QR トークンを、QR を小さく（荒く）できる短いバイナリ形式で署名・検証する。
//
// 形式（Base32、パディングなし。QR の英数字モードに収まる）:
//
//	受け渡し: tag(1)=2 | jti(8) | from(16) | exp(uint32, unix 秒) | due(uint16) | qty(uvarint) | item(UTF-8) | mac(12)
//	返却:     tag(1)=3 | handoff_id(16) | exp(uint32, unix 秒) | mac(12)
//
// due は返却期限の 1970-01-01 からの日数（0 は返却期限なし）。
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
	"relay/internal/handoff/domain"
)

const (
	// tagHandoff は受け渡し用のトークン。1 は返却期限を入れる前の形式で、今は使わない。
	tagHandoff = 2
	tagReturn  = 3

	jtiLen = 8
	macLen = 12

	handoffHeadLen  = 1 + jtiLen + 16 + 4 + 2 // tag + jti + from + exp + due
	handoffMinBytes = handoffHeadLen + 1 + macLen
	returnLen       = 1 + 16 + 4 + macLen // tag + handoff_id + exp + mac

	secondsPerDay = 24 * 60 * 60
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
	due, err := encodeDueDate(c.DueDate)
	if err != nil {
		return "", err
	}
	buf := make([]byte, 0, handoffHeadLen+binary.MaxVarintLen64+len(c.Item)+macLen)
	buf = append(buf, tagHandoff)
	buf = append(buf, jti...)
	buf = append(buf, c.From[:]...)
	buf = binary.BigEndian.AppendUint32(buf, uint32(c.ExpiresAt.Unix()))
	buf = binary.BigEndian.AppendUint16(buf, due)
	buf = binary.AppendUvarint(buf, uint64(c.Qty))
	buf = append(buf, c.Item...)
	buf = append(buf, s.mac(buf)...)
	return encoding.EncodeToString(buf), nil
}

func (s *CompactSigner) Verify(token string, now time.Time) (application.TokenClaims, error) {
	invalid := application.TokenClaims{}
	body, err := s.open(token)
	if err != nil {
		return invalid, err
	}
	if body[0] == tagReturn {
		return invalid, application.ErrReturnToken
	}
	if body[0] != tagHandoff || len(body)+macLen < handoffMinBytes {
		return invalid, application.ErrInvalidToken
	}

	jti := body[1 : 1+jtiLen]
	from, err := uuid.FromBytes(body[1+jtiLen : 1+jtiLen+16])
	if err != nil {
		return invalid, application.ErrInvalidToken
	}
	exp := time.Unix(int64(binary.BigEndian.Uint32(body[1+jtiLen+16:1+jtiLen+16+4])), 0)
	due := decodeDueDate(binary.BigEndian.Uint16(body[1+jtiLen+16+4 : handoffHeadLen]))
	qty, n := binary.Uvarint(body[handoffHeadLen:])
	if n <= 0 || qty > 1<<31 {
		return invalid, application.ErrInvalidToken
	}
	item := body[handoffHeadLen+n:]
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
		DueDate:   due,
		ExpiresAt: exp,
	}, nil
}

func (s *CompactSigner) SignReturn(c application.ReturnClaims) (string, error) {
	buf := make([]byte, 0, returnLen)
	buf = append(buf, tagReturn)
	buf = append(buf, c.HandoffID[:]...)
	buf = binary.BigEndian.AppendUint32(buf, uint32(c.ExpiresAt.Unix()))
	buf = append(buf, s.mac(buf)...)
	return encoding.EncodeToString(buf), nil
}

func (s *CompactSigner) VerifyReturn(token string, now time.Time) (application.ReturnClaims, error) {
	invalid := application.ReturnClaims{}
	body, err := s.open(token)
	if err != nil {
		return invalid, err
	}
	if body[0] != tagReturn || len(body)+macLen != returnLen {
		return invalid, application.ErrInvalidToken
	}
	id, err := uuid.FromBytes(body[1:17])
	if err != nil {
		return invalid, application.ErrInvalidToken
	}
	exp := time.Unix(int64(binary.BigEndian.Uint32(body[17:21])), 0)
	if !now.Before(exp) {
		return invalid, application.ErrTokenExpired
	}
	return application.ReturnClaims{HandoffID: id, ExpiresAt: exp}, nil
}

// open は Base32 を戻して署名を検証し、mac を除いたバイト列を返す。
func (s *CompactSigner) open(token string) ([]byte, error) {
	b, err := encoding.DecodeString(strings.ToUpper(strings.TrimSpace(token)))
	if err != nil || len(b) < 1+macLen {
		return nil, application.ErrInvalidToken
	}
	body, sig := b[:len(b)-macLen], b[len(b)-macLen:]
	if !hmac.Equal(sig, s.mac(body)) {
		return nil, application.ErrInvalidToken
	}
	return body, nil
}

func (s *CompactSigner) mac(b []byte) []byte {
	m := hmac.New(sha256.New, s.secret)
	m.Write(b)
	return m.Sum(nil)[:macLen]
}

func encodeDueDate(d *domain.DueDate) (uint16, error) {
	if d == nil {
		return 0, nil
	}
	days := d.Time().Unix() / secondsPerDay
	if days <= 0 || days > 0xFFFF {
		return 0, application.ErrInvalidToken
	}
	return uint16(days), nil
}

func decodeDueDate(days uint16) *domain.DueDate {
	if days == 0 {
		return nil
	}
	d := domain.DueDateOf(time.Unix(int64(days)*secondsPerDay, 0).UTC())
	return &d
}
