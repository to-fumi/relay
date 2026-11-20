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
	_, err := domain.NewHandoff(uuid.New(), "jti", u, u, "item", 1, time.Now())
	if !errors.Is(err, domain.ErrSelfHandoff) {
		t.Fatalf("err = %v; want ErrSelfHandoff", err)
	}
}
