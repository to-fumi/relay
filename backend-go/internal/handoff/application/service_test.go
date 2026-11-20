package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"relay/internal/handoff/application"
	"relay/internal/handoff/domain"

	"github.com/google/uuid"
)

type memRepo struct{ saved []domain.Handoff }

func (r *memRepo) Save(_ context.Context, h domain.Handoff) error {
	for _, s := range r.saved {
		if s.JTI() == h.JTI() {
			return domain.ErrAlreadyRecorded
		}
	}
	r.saved = append(r.saved, h)
	return nil
}

func (r *memRepo) ListByUser(context.Context, uuid.UUID, int) ([]domain.Handoff, error) {
	return r.saved, nil
}

// fakeSigner はトークンの中身をメモリに持つ。署名の検証は adapter/token のテストで行う。
type fakeSigner struct {
	issued map[string]application.TokenClaims
}

func (s *fakeSigner) Sign(c application.TokenClaims) (string, error) {
	tok := "tok-" + c.JTI
	s.issued[tok] = c
	return tok, nil
}

func (s *fakeSigner) Verify(tok string, now time.Time) (application.TokenClaims, error) {
	c, ok := s.issued[tok]
	if !ok {
		return application.TokenClaims{}, application.ErrInvalidToken
	}
	if !now.Before(c.ExpiresAt) {
		return application.TokenClaims{}, application.ErrTokenExpired
	}
	return c, nil
}

type names map[uuid.UUID]string

func (n names) DisplayNames(context.Context, []uuid.UUID) (map[uuid.UUID]string, error) {
	return n, nil
}

type fixture struct {
	svc      *application.Service
	repo     *memRepo
	now      time.Time
	from, to uuid.UUID
}

func newFixture() *fixture {
	f := &fixture{repo: &memRepo{}, now: time.Date(2026, 11, 21, 9, 0, 0, 0, time.UTC), from: uuid.New(), to: uuid.New()}
	signer := &fakeSigner{issued: map[string]application.TokenClaims{}}
	f.svc = application.NewService(f.repo, signer, names{f.from: "A", f.to: "B"}, time.Minute, func() time.Time { return f.now })
	return f
}

func TestReceive_RecordsHandoffFromIssuer(t *testing.T) {
	f := newFixture()
	ctx := context.Background()

	issued, err := f.svc.IssueToken(ctx, f.from, " ケーブル ", 2)
	if err != nil {
		t.Fatal(err)
	}
	got, err := f.svc.Receive(ctx, f.to, issued.Token)
	if err != nil {
		t.Fatal(err)
	}
	if got.From.ID != f.from || got.To.ID != f.to || got.Item != "ケーブル" || got.Qty != 2 {
		t.Fatalf("unexpected handoff: %+v", got)
	}
	if got.From.DisplayName != "A" || got.To.DisplayName != "B" {
		t.Fatalf("unexpected names: %+v", got)
	}
}

func TestReceive_TokenCanBeUsedOnce(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	issued, _ := f.svc.IssueToken(ctx, f.from, "x", 1)

	if _, err := f.svc.Receive(ctx, f.to, issued.Token); err != nil {
		t.Fatal(err)
	}
	if _, err := f.svc.Receive(ctx, f.to, issued.Token); !errors.Is(err, domain.ErrAlreadyRecorded) {
		t.Fatalf("err = %v; want ErrAlreadyRecorded", err)
	}
}

func TestReceive_Expired(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	issued, _ := f.svc.IssueToken(ctx, f.from, "x", 1)

	f.now = f.now.Add(2 * time.Minute)
	if _, err := f.svc.Receive(ctx, f.to, issued.Token); !errors.Is(err, application.ErrTokenExpired) {
		t.Fatalf("err = %v; want ErrTokenExpired", err)
	}
}

func TestReceive_RejectsSelf(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	issued, _ := f.svc.IssueToken(ctx, f.from, "x", 1)

	if _, err := f.svc.Receive(ctx, f.from, issued.Token); !errors.Is(err, domain.ErrSelfHandoff) {
		t.Fatalf("err = %v; want ErrSelfHandoff", err)
	}
}

func TestIssueToken_ValidatesInput(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	if _, err := f.svc.IssueToken(ctx, f.from, "", 1); !errors.Is(err, domain.ErrInvalidItem) {
		t.Errorf("err = %v; want ErrInvalidItem", err)
	}
	if _, err := f.svc.IssueToken(ctx, f.from, "x", 0); !errors.Is(err, domain.ErrInvalidQuantity) {
		t.Errorf("err = %v; want ErrInvalidQuantity", err)
	}
}
