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
	ErrDueDateInPast   = errors.New("due date must be today or later")
	ErrNotReturnable   = errors.New("handoff has no due date and does not need to be returned")
	ErrAlreadyReturned = errors.New("handoff already returned")
	ErrNotBorrower     = errors.New("only the borrower can return this handoff")
	ErrNotLender       = errors.New("only the lender can record the return")
	// ErrReturnAlreadyRequested は、対面でない返却がすでに送られて承認待ちであることを表す。
	ErrReturnAlreadyRequested = errors.New("return already requested")
	// ErrReturnNotRequested は、承認しようとした返却がまだ送られていないことを表す。
	ErrReturnNotRequested = errors.New("return has not been requested")
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

// DueDate は返却期限（日付のみ。その日の終わりまでに返す）。UTC の 0 時として持つ。
type DueDate struct{ t time.Time }

func NewDueDate(year int, month time.Month, day int) DueDate {
	return DueDate{t: time.Date(year, month, day, 0, 0, 0, 0, time.UTC)}
}

// DueDateOf は t の（t の Location での）日付を返す。
func DueDateOf(t time.Time) DueDate {
	return NewDueDate(t.Date())
}

// Time は期限の日付を UTC の 0 時で返す。
func (d DueDate) Time() time.Time { return d.t }

func (d DueDate) Before(o DueDate) bool { return d.t.Before(o.t) }

// CheckNotBefore は期限が today より前でないことを確かめる。
func (d DueDate) CheckNotBefore(today DueDate) error {
	if d.Before(today) {
		return ErrDueDateInPast
	}
	return nil
}

// Handoff は 1 回の受け渡しの記録。返却期限がある場合は、返却の手続きで returnRequestedAt と returnedAt だけを変更する。
type Handoff struct {
	id                uuid.UUID
	jti               string
	from              uuid.UUID
	to                uuid.UUID
	item              Item
	qty               Quantity
	createdAt         time.Time
	due               *DueDate
	returnRequestedAt *time.Time
	returnedAt        *time.Time
}

// NewHandoff は新しい受け渡しを作る。jti は QR トークンの ID で、同じトークンによる二重記録を防ぐために使う。
// due が nil の場合は、返却の必要がない受け渡しになる。
func NewHandoff(id uuid.UUID, jti string, from, to uuid.UUID, item Item, qty Quantity, due *DueDate, at time.Time) (Handoff, error) {
	if from == to {
		return Handoff{}, ErrSelfHandoff
	}
	return Handoff{id: id, jti: jti, from: from, to: to, item: item, qty: qty, createdAt: at, due: due}, nil
}

// Restore は永続化された値から Handoff を復元する（リポジトリの実装から使う）。
func Restore(id uuid.UUID, jti string, from, to uuid.UUID, item string, qty int, at time.Time, due *DueDate, returnRequestedAt, returnedAt *time.Time) Handoff {
	return Handoff{
		id: id, jti: jti, from: from, to: to, item: Item(item), qty: Quantity(qty),
		createdAt: at, due: due, returnRequestedAt: returnRequestedAt, returnedAt: returnedAt,
	}
}

func (h Handoff) ID() uuid.UUID        { return h.id }
func (h Handoff) JTI() string          { return h.jti }
func (h Handoff) From() uuid.UUID      { return h.from }
func (h Handoff) To() uuid.UUID        { return h.to }
func (h Handoff) Item() Item           { return h.item }
func (h Handoff) Qty() Quantity        { return h.qty }
func (h Handoff) CreatedAt() time.Time { return h.createdAt }

// DueDate は返却期限を返す。返却の必要がない受け渡しでは nil。
func (h Handoff) DueDate() *DueDate { return h.due }

// ReturnRequestedAt は対面でない返却を送った日時を返す。送っていなければ nil。
func (h Handoff) ReturnRequestedAt() *time.Time { return h.returnRequestedAt }

// ReturnedAt は返却された日時を返す。まだ返却されていなければ nil。
func (h Handoff) ReturnedAt() *time.Time { return h.returnedAt }

// CheckReturnableBy は borrower（受け取った人）が返却の手続きを始められるかを確かめる。
func (h Handoff) CheckReturnableBy(borrower uuid.UUID) error {
	if borrower != h.to {
		return ErrNotBorrower
	}
	return h.checkOutstanding()
}

// RequestReturn は borrower（受け取った人）が、対面でない返却を送ったことを記録する。貸した人の承認待ちになる。
func (h *Handoff) RequestReturn(borrower uuid.UUID, at time.Time) error {
	if err := h.CheckReturnableBy(borrower); err != nil {
		return err
	}
	if h.returnRequestedAt != nil {
		return ErrReturnAlreadyRequested
	}
	h.returnRequestedAt = &at
	return nil
}

// ConfirmReturn は lender（渡した人）が、送られた返却を承認したことを記録する。
func (h *Handoff) ConfirmReturn(lender uuid.UUID, at time.Time) error {
	if lender != h.from {
		return ErrNotLender
	}
	if err := h.checkOutstanding(); err != nil {
		return err
	}
	if h.returnRequestedAt == nil {
		return ErrReturnNotRequested
	}
	h.returnedAt = &at
	return nil
}

// MarkReturned は lender（渡した人）が返却を受け取ったことを記録する。
func (h *Handoff) MarkReturned(lender uuid.UUID, at time.Time) error {
	if lender != h.from {
		return ErrNotLender
	}
	if err := h.checkOutstanding(); err != nil {
		return err
	}
	h.returnedAt = &at
	return nil
}

func (h Handoff) checkOutstanding() error {
	if h.due == nil {
		return ErrNotReturnable
	}
	if h.returnedAt != nil {
		return ErrAlreadyReturned
	}
	return nil
}
