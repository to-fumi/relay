package application_test

import (
	"context"
	"errors"
	"slices"
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

func (r *memRepo) Get(_ context.Context, id uuid.UUID) (domain.Handoff, error) {
	for _, s := range r.saved {
		if s.ID() == id {
			return s, nil
		}
	}
	return domain.Handoff{}, domain.ErrNotFound
}

func (r *memRepo) SaveReturn(_ context.Context, h domain.Handoff) error {
	for i, s := range r.saved {
		if s.ID() == h.ID() {
			if s.ReturnedAt() != nil {
				return domain.ErrAlreadyReturned
			}
			r.saved[i] = h
			return nil
		}
	}
	return domain.ErrNotFound
}

func (r *memRepo) GetMany(ctx context.Context, ids []uuid.UUID) ([]domain.Handoff, error) {
	out := make([]domain.Handoff, len(ids))
	for i, id := range ids {
		h, err := r.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		out[i] = h
	}
	return out, nil
}

func (r *memRepo) SaveReturnRequests(_ context.Context, hs []domain.Handoff) error {
	return r.replaceAll(hs, func(s domain.Handoff) bool { return s.ReturnedAt() == nil && s.ReturnRequestedAt() == nil })
}

func (r *memRepo) SaveConfirmedReturns(_ context.Context, hs []domain.Handoff) error {
	return r.replaceAll(hs, func(s domain.Handoff) bool { return s.ReturnedAt() == nil && s.ReturnRequestedAt() != nil })
}

// replaceAll は、保存済みのものがすべて ok を満たすときだけ hs で置き換える。
func (r *memRepo) replaceAll(hs []domain.Handoff, ok func(domain.Handoff) bool) error {
	next := slices.Clone(r.saved)
	for _, h := range hs {
		i := slices.IndexFunc(next, func(s domain.Handoff) bool { return s.ID() == h.ID() })
		if i < 0 || !ok(next[i]) {
			return domain.ErrAlreadyReturned
		}
		next[i] = h
	}
	r.saved = next
	return nil
}

func (r *memRepo) ListByUser(context.Context, uuid.UUID, int) ([]domain.Handoff, error) {
	return r.saved, nil
}

func (r *memRepo) ListOutstandingByUser(context.Context, uuid.UUID) ([]domain.Handoff, error) {
	return slices.DeleteFunc(slices.Clone(r.saved), func(h domain.Handoff) bool {
		return h.DueDate() == nil || h.ReturnedAt() != nil
	}), nil
}

// fakeSigner はトークンの中身をメモリに持つ。署名の検証は adapter/token のテストで行う。
type fakeSigner struct {
	issued   map[string]application.TokenClaims
	returns  map[string]application.ReturnClaims
	sequence int
}

func (s *fakeSigner) Sign(c application.TokenClaims) (string, error) {
	tok := "tok-" + c.JTI
	s.issued[tok] = c
	return tok, nil
}

func (s *fakeSigner) Verify(tok string, now time.Time) (application.TokenClaims, error) {
	if _, ok := s.returns[tok]; ok {
		return application.TokenClaims{}, application.ErrReturnToken
	}
	c, ok := s.issued[tok]
	if !ok {
		return application.TokenClaims{}, application.ErrInvalidToken
	}
	if !now.Before(c.ExpiresAt) {
		return application.TokenClaims{}, application.ErrTokenExpired
	}
	return c, nil
}

func (s *fakeSigner) SignReturn(c application.ReturnClaims) (string, error) {
	s.sequence++
	tok := "ret-" + c.HandoffID.String() + "-" + string(rune('a'+s.sequence))
	s.returns[tok] = c
	return tok, nil
}

func (s *fakeSigner) VerifyReturn(tok string, now time.Time) (application.ReturnClaims, error) {
	c, ok := s.returns[tok]
	if !ok {
		return application.ReturnClaims{}, application.ErrInvalidToken
	}
	if !now.Before(c.ExpiresAt) {
		return application.ReturnClaims{}, application.ErrTokenExpired
	}
	return c, nil
}

type names map[uuid.UUID]string

func (n names) DisplayNames(context.Context, []uuid.UUID) (map[uuid.UUID]string, error) {
	return n, nil
}

var jst = time.FixedZone("Asia/Tokyo", 9*60*60)

type fixture struct {
	svc      *application.Service
	repo     *memRepo
	now      time.Time
	from, to uuid.UUID
}

func newFixture() *fixture {
	// 日本時間では 2026-11-22 の 0:30（UTC ではまだ 11-21）
	f := &fixture{repo: &memRepo{}, now: time.Date(2026, 11, 21, 15, 30, 0, 0, time.UTC), from: uuid.New(), to: uuid.New()}
	signer := &fakeSigner{issued: map[string]application.TokenClaims{}, returns: map[string]application.ReturnClaims{}}
	f.svc = application.NewService(f.repo, signer, names{f.from: "A", f.to: "B"}, time.Minute, func() time.Time { return f.now }, jst)
	return f
}

func dueDate(y int, m time.Month, d int) *domain.DueDate {
	v := domain.NewDueDate(y, m, d)
	return &v
}

// lend は from が to に、返却期限付きで渡した記録を作る。
func (f *fixture) lend(t *testing.T) application.HandoffView {
	t.Helper()
	ctx := context.Background()
	issued, err := f.svc.IssueToken(ctx, f.from, "ケーブル", 1, dueDate(2026, 11, 28))
	if err != nil {
		t.Fatal(err)
	}
	v, err := f.svc.Receive(ctx, f.to, issued.Token)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestReceive_RecordsHandoffFromIssuer(t *testing.T) {
	f := newFixture()
	ctx := context.Background()

	issued, err := f.svc.IssueToken(ctx, f.from, " ケーブル ", 2, nil)
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
	if got.DueDate != nil || got.ReturnedAt != nil {
		t.Fatalf("handoff without due date has DueDate=%v ReturnedAt=%v", got.DueDate, got.ReturnedAt)
	}
	// 送る側は、発行したときの JTI で読み取られたことを知る
	if issued.JTI == "" || got.JTI != issued.JTI {
		t.Fatalf("JTI = %q; want %q (issued)", got.JTI, issued.JTI)
	}
}

func TestReceive_RecordsDueDate(t *testing.T) {
	f := newFixture()
	got := f.lend(t)
	if got.DueDate == nil || !got.DueDate.Equal(time.Date(2026, 11, 28, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("DueDate = %v; want 2026-11-28", got.DueDate)
	}
}

func TestReceive_TokenCanBeUsedOnce(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	issued, _ := f.svc.IssueToken(ctx, f.from, "x", 1, nil)

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
	issued, _ := f.svc.IssueToken(ctx, f.from, "x", 1, nil)

	f.now = f.now.Add(2 * time.Minute)
	if _, err := f.svc.Receive(ctx, f.to, issued.Token); !errors.Is(err, application.ErrTokenExpired) {
		t.Fatalf("err = %v; want ErrTokenExpired", err)
	}
}

func TestReceive_RejectsSelf(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	issued, _ := f.svc.IssueToken(ctx, f.from, "x", 1, nil)

	if _, err := f.svc.Receive(ctx, f.from, issued.Token); !errors.Is(err, domain.ErrSelfHandoff) {
		t.Fatalf("err = %v; want ErrSelfHandoff", err)
	}
}

func TestReceive_RejectsReturnToken(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	h := f.lend(t)
	ret, _ := f.svc.IssueReturnToken(ctx, f.to, h.ID)

	if _, err := f.svc.Receive(ctx, f.from, ret.Token); !errors.Is(err, application.ErrReturnToken) {
		t.Fatalf("err = %v; want ErrReturnToken", err)
	}
}

func TestIssueToken_ValidatesInput(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	if _, err := f.svc.IssueToken(ctx, f.from, "", 1, nil); !errors.Is(err, domain.ErrInvalidItem) {
		t.Errorf("err = %v; want ErrInvalidItem", err)
	}
	if _, err := f.svc.IssueToken(ctx, f.from, "x", 0, nil); !errors.Is(err, domain.ErrInvalidQuantity) {
		t.Errorf("err = %v; want ErrInvalidQuantity", err)
	}
}

func TestIssueToken_DueDateIsTodayOrLaterInJST(t *testing.T) {
	f := newFixture() // 日本時間では 2026-11-22
	ctx := context.Background()
	if _, err := f.svc.IssueToken(ctx, f.from, "x", 1, dueDate(2026, 11, 21)); !errors.Is(err, domain.ErrDueDateInPast) {
		t.Errorf("yesterday: err = %v; want ErrDueDateInPast", err)
	}
	if _, err := f.svc.IssueToken(ctx, f.from, "x", 1, dueDate(2026, 11, 22)); err != nil {
		t.Errorf("today: err = %v; want nil", err)
	}
}

func TestReturn_LenderScansBorrowersToken(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	h := f.lend(t)

	if out, _ := f.svc.ListOutstanding(ctx, f.from); len(out) != 1 {
		t.Fatalf("outstanding before return = %d; want 1", len(out))
	}
	ret, err := f.svc.IssueReturnToken(ctx, f.to, h.ID)
	if err != nil {
		t.Fatal(err)
	}
	got, err := f.svc.Return(ctx, f.from, ret.Token)
	if err != nil {
		t.Fatal(err)
	}
	if got.ReturnedAt == nil || !got.ReturnedAt.Equal(f.now) {
		t.Fatalf("ReturnedAt = %v; want %v", got.ReturnedAt, f.now)
	}
	if out, _ := f.svc.ListOutstanding(ctx, f.from); len(out) != 0 {
		t.Fatalf("outstanding after return = %d; want 0", len(out))
	}
}

func TestReturn_CanBeRecordedOnce(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	h := f.lend(t)
	ret, _ := f.svc.IssueReturnToken(ctx, f.to, h.ID)
	if _, err := f.svc.Return(ctx, f.from, ret.Token); err != nil {
		t.Fatal(err)
	}

	if _, err := f.svc.Return(ctx, f.from, ret.Token); !errors.Is(err, domain.ErrAlreadyReturned) {
		t.Fatalf("Return twice: err = %v; want ErrAlreadyReturned", err)
	}
	if _, err := f.svc.IssueReturnToken(ctx, f.to, h.ID); !errors.Is(err, domain.ErrAlreadyReturned) {
		t.Fatalf("IssueReturnToken after return: err = %v; want ErrAlreadyReturned", err)
	}
}

func TestReturn_OnlyLenderCanRecord(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	h := f.lend(t)
	ret, _ := f.svc.IssueReturnToken(ctx, f.to, h.ID)

	for _, who := range []uuid.UUID{f.to, uuid.New()} {
		if _, err := f.svc.Return(ctx, who, ret.Token); !errors.Is(err, domain.ErrNotLender) {
			t.Fatalf("err = %v; want ErrNotLender", err)
		}
	}
}

func TestReturn_Expired(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	h := f.lend(t)
	ret, _ := f.svc.IssueReturnToken(ctx, f.to, h.ID)

	f.now = f.now.Add(2 * time.Minute)
	if _, err := f.svc.Return(ctx, f.from, ret.Token); !errors.Is(err, application.ErrTokenExpired) {
		t.Fatalf("err = %v; want ErrTokenExpired", err)
	}
}

func TestIssueReturnToken_OnlyBorrowerOfLoan(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	h := f.lend(t)

	if _, err := f.svc.IssueReturnToken(ctx, f.from, h.ID); !errors.Is(err, domain.ErrNotBorrower) {
		t.Errorf("lender: err = %v; want ErrNotBorrower", err)
	}
	if _, err := f.svc.IssueReturnToken(ctx, f.to, uuid.New()); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("unknown id: err = %v; want ErrNotFound", err)
	}

	issued, _ := f.svc.IssueToken(ctx, f.from, "x", 1, nil)
	noDue, _ := f.svc.Receive(ctx, f.to, issued.Token)
	if _, err := f.svc.IssueReturnToken(ctx, f.to, noDue.ID); !errors.Is(err, domain.ErrNotReturnable) {
		t.Errorf("no due date: err = %v; want ErrNotReturnable", err)
	}
}

// lendFrom は lender が f.to に、返却期限付きで渡した記録を作る。
func (f *fixture) lendFrom(t *testing.T, lender uuid.UUID) application.HandoffView {
	t.Helper()
	ctx := context.Background()
	issued, err := f.svc.IssueToken(ctx, lender, "ケーブル", 1, dueDate(2026, 11, 28))
	if err != nil {
		t.Fatal(err)
	}
	v, err := f.svc.Receive(ctx, f.to, issued.Token)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestRequestReturns_ThenLendersConfirm(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	other := uuid.New()
	a, b := f.lend(t), f.lendFrom(t, other)

	// 貸した人が違うものも、まとめて送れる
	got, err := f.svc.RequestReturns(ctx, f.to, []uuid.UUID{a.ID, b.ID, a.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d; want 2 (duplicates removed)", len(got))
	}
	for _, v := range got {
		if v.ReturnRequestedAt == nil || !v.ReturnRequestedAt.Equal(f.now) || v.ReturnedAt != nil {
			t.Fatalf("after request: %+v", v)
		}
	}
	// 承認されるまでは返却リストに残る
	if out, _ := f.svc.ListOutstanding(ctx, f.to); len(out) != 2 {
		t.Fatalf("outstanding = %d; want 2", len(out))
	}

	// それぞれの貸した人が承認する。他人に貸したものは承認できない
	if _, err := f.svc.ConfirmReturns(ctx, f.from, []uuid.UUID{a.ID, b.ID}); !errors.Is(err, domain.ErrNotLender) {
		t.Fatalf("confirm other's loan: err = %v; want ErrNotLender", err)
	}
	f.now = f.now.Add(time.Hour)
	done, err := f.svc.ConfirmReturns(ctx, f.from, []uuid.UUID{a.ID})
	if err != nil {
		t.Fatal(err)
	}
	if done[0].ReturnedAt == nil || !done[0].ReturnedAt.Equal(f.now) {
		t.Fatalf("ReturnedAt = %v; want %v", done[0].ReturnedAt, f.now)
	}
	if _, err := f.svc.ConfirmReturns(ctx, other, []uuid.UUID{b.ID}); err != nil {
		t.Fatal(err)
	}
	if out, _ := f.svc.ListOutstanding(ctx, f.to); len(out) != 0 {
		t.Fatalf("outstanding after confirm = %d; want 0", len(out))
	}
}

func TestRequestReturns_AllOrNothing(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	a, b := f.lend(t), f.lend(t)
	if _, err := f.svc.RequestReturns(ctx, f.to, []uuid.UUID{a.ID}); err != nil {
		t.Fatal(err)
	}

	issued, _ := f.svc.IssueToken(ctx, f.from, "x", 1, nil)
	noDue, _ := f.svc.Receive(ctx, f.to, issued.Token)

	tests := []struct {
		name string
		ids  []uuid.UUID
		who  uuid.UUID
		want error
	}{
		{"already requested", []uuid.UUID{b.ID, a.ID}, f.to, domain.ErrReturnAlreadyRequested},
		{"not borrower", []uuid.UUID{b.ID}, f.from, domain.ErrNotBorrower},
		{"unknown id", []uuid.UUID{b.ID, uuid.New()}, f.to, domain.ErrNotFound},
		{"no due date", []uuid.UUID{b.ID, noDue.ID}, f.to, domain.ErrNotReturnable},
		{"empty", nil, f.to, application.ErrInvalidSelection},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := f.svc.RequestReturns(ctx, tt.who, tt.ids); !errors.Is(err, tt.want) {
				t.Fatalf("err = %v; want %v", err, tt.want)
			}
			// b はどの失敗でも記録されない
			if h, _ := f.repo.Get(ctx, b.ID); h.ReturnRequestedAt() != nil {
				t.Fatal("b was requested despite the failure")
			}
		})
	}
}

func TestRequestReturns_TooMany(t *testing.T) {
	f := newFixture()
	ids := make([]uuid.UUID, application.MaxBatch+1)
	for i := range ids {
		ids[i] = uuid.New()
	}
	if _, err := f.svc.RequestReturns(context.Background(), f.to, ids); !errors.Is(err, application.ErrInvalidSelection) {
		t.Fatalf("err = %v; want ErrInvalidSelection", err)
	}
}

func TestConfirmReturns_AllOrNothing(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	a, b := f.lend(t), f.lend(t)
	if _, err := f.svc.RequestReturns(ctx, f.to, []uuid.UUID{a.ID}); err != nil {
		t.Fatal(err)
	}

	if _, err := f.svc.ConfirmReturns(ctx, f.from, []uuid.UUID{a.ID, b.ID}); !errors.Is(err, domain.ErrReturnNotRequested) {
		t.Fatalf("err = %v; want ErrReturnNotRequested", err)
	}
	if h, _ := f.repo.Get(ctx, a.ID); h.ReturnedAt() != nil {
		t.Fatal("a was returned despite the failure")
	}
	if _, err := f.svc.ConfirmReturns(ctx, f.from, []uuid.UUID{a.ID}); err != nil {
		t.Fatal(err)
	}
	if _, err := f.svc.ConfirmReturns(ctx, f.from, []uuid.UUID{a.ID}); !errors.Is(err, domain.ErrAlreadyReturned) {
		t.Fatalf("confirm twice: err = %v; want ErrAlreadyReturned", err)
	}
}

func TestReturn_QRStillWorksWhileRequested(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	h := f.lend(t)
	if _, err := f.svc.RequestReturns(ctx, f.to, []uuid.UUID{h.ID}); err != nil {
		t.Fatal(err)
	}
	ret, err := f.svc.IssueReturnToken(ctx, f.to, h.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.svc.Return(ctx, f.from, ret.Token); err != nil {
		t.Fatal(err)
	}
	if _, err := f.svc.ConfirmReturns(ctx, f.from, []uuid.UUID{h.ID}); !errors.Is(err, domain.ErrAlreadyReturned) {
		t.Fatalf("err = %v; want ErrAlreadyReturned", err)
	}
}
