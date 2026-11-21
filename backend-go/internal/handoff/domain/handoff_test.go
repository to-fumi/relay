package domain_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"relay/internal/handoff/domain"

	"github.com/google/uuid"
)

func TestNewItem(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    domain.Item
		wantErr error
	}{
		{"trims spaces", "  ケーブル ", "ケーブル", nil},
		{"empty", "  ", "", domain.ErrInvalidItem},
		{"200 runes ok", strings.Repeat("あ", 200), domain.Item(strings.Repeat("あ", 200)), nil},
		{"201 runes", strings.Repeat("あ", 201), "", domain.ErrInvalidItem},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.NewItem(tt.in)
			if !errors.Is(err, tt.wantErr) || got != tt.want {
				t.Fatalf("NewItem(%q) = %q, %v; want %q, %v", tt.in, got, err, tt.want, tt.wantErr)
			}
		})
	}
}

func TestNewQuantity(t *testing.T) {
	for _, n := range []int{0, -1} {
		if _, err := domain.NewQuantity(n); !errors.Is(err, domain.ErrInvalidQuantity) {
			t.Errorf("NewQuantity(%d) err = %v; want ErrInvalidQuantity", n, err)
		}
	}
	if q, err := domain.NewQuantity(3); err != nil || q != 3 {
		t.Errorf("NewQuantity(3) = %v, %v", q, err)
	}
}

func TestNewHandoff_RejectsSelf(t *testing.T) {
	u := uuid.New()
	_, err := domain.NewHandoff(uuid.New(), "jti", u, u, "item", 1, nil, time.Now())
	if !errors.Is(err, domain.ErrSelfHandoff) {
		t.Fatalf("err = %v; want ErrSelfHandoff", err)
	}
}

func TestDueDate_CheckNotBefore(t *testing.T) {
	today := domain.NewDueDate(2026, 11, 22)
	if err := domain.NewDueDate(2026, 11, 21).CheckNotBefore(today); !errors.Is(err, domain.ErrDueDateInPast) {
		t.Errorf("yesterday: err = %v; want ErrDueDateInPast", err)
	}
	for _, d := range []domain.DueDate{today, domain.NewDueDate(2026, 12, 1)} {
		if err := d.CheckNotBefore(today); err != nil {
			t.Errorf("%v: err = %v; want nil", d.Time(), err)
		}
	}
}

func TestDueDateOf_UsesLocalDate(t *testing.T) {
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	got := domain.DueDateOf(time.Date(2026, 11, 22, 0, 30, 0, 0, jst))
	if got != domain.NewDueDate(2026, 11, 22) {
		t.Fatalf("got %v; want 2026-11-22", got.Time())
	}
}

func TestMarkReturned(t *testing.T) {
	lender, borrower := uuid.New(), uuid.New()
	due := domain.NewDueDate(2026, 11, 28)
	newLoan := func(due *domain.DueDate) domain.Handoff {
		h, err := domain.NewHandoff(uuid.New(), "jti", lender, borrower, "item", 1, due, time.Now())
		if err != nil {
			t.Fatal(err)
		}
		return h
	}

	t.Run("lender records once", func(t *testing.T) {
		h := newLoan(&due)
		if err := h.CheckReturnableBy(borrower); err != nil {
			t.Fatalf("CheckReturnableBy: %v", err)
		}
		at := time.Now()
		if err := h.MarkReturned(lender, at); err != nil {
			t.Fatal(err)
		}
		if h.ReturnedAt() == nil || !h.ReturnedAt().Equal(at) {
			t.Fatalf("ReturnedAt = %v; want %v", h.ReturnedAt(), at)
		}
		if err := h.MarkReturned(lender, at); !errors.Is(err, domain.ErrAlreadyReturned) {
			t.Fatalf("second: err = %v; want ErrAlreadyReturned", err)
		}
		if err := h.CheckReturnableBy(borrower); !errors.Is(err, domain.ErrAlreadyReturned) {
			t.Fatalf("CheckReturnableBy after return: err = %v; want ErrAlreadyReturned", err)
		}
	})
	t.Run("not lender", func(t *testing.T) {
		h := newLoan(&due)
		if err := h.MarkReturned(borrower, time.Now()); !errors.Is(err, domain.ErrNotLender) {
			t.Fatalf("err = %v; want ErrNotLender", err)
		}
	})
	t.Run("not borrower", func(t *testing.T) {
		h := newLoan(&due)
		if err := h.CheckReturnableBy(lender); !errors.Is(err, domain.ErrNotBorrower) {
			t.Fatalf("err = %v; want ErrNotBorrower", err)
		}
	})
	t.Run("no due date", func(t *testing.T) {
		h := newLoan(nil)
		if err := h.MarkReturned(lender, time.Now()); !errors.Is(err, domain.ErrNotReturnable) {
			t.Fatalf("err = %v; want ErrNotReturnable", err)
		}
	})
}
