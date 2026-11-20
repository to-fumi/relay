// Package domain は受け渡し（Handoff）のドメインモデル。外部のパッケージには依存しない。
package domain

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const maxItemLength = 200

var (
	ErrInvalidItem     = errors.New("item must be 1 to 200 characters")
	ErrInvalidQuantity = errors.New("qty must be greater than 0")
	ErrSelfHandoff     = errors.New("cannot hand off to yourself")
)

// Item は受け渡す品目名（自由入力）。
type Item string

func NewItem(s string) (Item, error) {
	s = strings.TrimSpace(s)
	if s == "" || utf8.RuneCountInString(s) > maxItemLength {
		return "", ErrInvalidItem
	}
	return Item(s), nil
}

// Quantity は受け渡す数量。
type Quantity int

func NewQuantity(n int) (Quantity, error) {
	if n <= 0 {
		return 0, ErrInvalidQuantity
	}
	return Quantity(n), nil
}

// Handoff は 1 回の受け渡しの記録。作成後は変更しない。
type Handoff struct {
	id        uuid.UUID
	jti       string
	from      uuid.UUID
	to        uuid.UUID
	item      Item
	qty       Quantity
	createdAt time.Time
}

// NewHandoff は新しい受け渡しを作る。jti は QR トークンの ID で、同じトークンによる二重記録を防ぐために使う。
func NewHandoff(id uuid.UUID, jti string, from, to uuid.UUID, item Item, qty Quantity, at time.Time) (Handoff, error) {
	if from == to {
		return Handoff{}, ErrSelfHandoff
	}
	return Handoff{id: id, jti: jti, from: from, to: to, item: item, qty: qty, createdAt: at}, nil
}

// Restore は永続化された値から Handoff を復元する（リポジトリの実装から使う）。
func Restore(id uuid.UUID, jti string, from, to uuid.UUID, item string, qty int, at time.Time) Handoff {
	return Handoff{id: id, jti: jti, from: from, to: to, item: Item(item), qty: Quantity(qty), createdAt: at}
}

func (h Handoff) ID() uuid.UUID        { return h.id }
func (h Handoff) JTI() string          { return h.jti }
func (h Handoff) From() uuid.UUID      { return h.from }
func (h Handoff) To() uuid.UUID        { return h.to }
func (h Handoff) Item() Item           { return h.item }
func (h Handoff) Qty() Quantity        { return h.qty }
func (h Handoff) CreatedAt() time.Time { return h.createdAt }
