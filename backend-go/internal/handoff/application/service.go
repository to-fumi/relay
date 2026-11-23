package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"relay/internal/handoff/domain"

	"github.com/google/uuid"
)

// MaxBatch はまとめて返却・承認できる件数の上限。
const MaxBatch = 100

// ErrInvalidSelection は、まとめて扱う受け渡しの選び方（0 件、多すぎる）が正しくないことを表す。
var ErrInvalidSelection = errors.New("select 1 to 100 handoffs")

type Service struct {
	repo   domain.Repository
	signer TokenSigner
	users  UserDirectory
	ttl    time.Duration
	now    func() time.Time
	loc    *time.Location
}

// NewService は handoff のユースケースを作る。loc は返却期限の「今日」を決めるタイムゾーン。
func NewService(repo domain.Repository, signer TokenSigner, users UserDirectory, ttl time.Duration, now func() time.Time, loc *time.Location) *Service {
	return &Service{repo: repo, signer: signer, users: users, ttl: ttl, now: now, loc: loc}
}

type IssuedToken struct {
	Token     string
	ExpiresAt time.Time
	// JTI は受け渡し用のトークンの ID（返却用のトークンでは空）。記録された HandoffView の JTI と一致する
	JTI string
}

// HandoffView は表示用の受け渡し記録。
type HandoffView struct {
	ID                uuid.UUID
	JTI               string
	From              UserRef
	To                UserRef
	Item              string
	Qty               int
	CreatedAt         time.Time
	DueDate           *time.Time // 返却期限（UTC の 0 時）。なければ nil
	ReturnRequestedAt *time.Time // 対面でない返却を送った日時。送っていなければ nil
	ReturnedAt        *time.Time // 返却された日時。まだなら nil
}

type UserRef struct {
	ID          uuid.UUID
	DisplayName string
}

// IssueToken は渡す側（from）の QR トークンを発行する。トークンはどこにも保存しない。
// due を指定すると返却期限付きの受け渡しになる。期限は今日以降でなければならない。
func (s *Service) IssueToken(_ context.Context, from uuid.UUID, item string, qty int, due *domain.DueDate) (IssuedToken, error) {
	it, err := domain.NewItem(item)
	if err != nil {
		return IssuedToken{}, err
	}
	q, err := domain.NewQuantity(qty)
	if err != nil {
		return IssuedToken{}, err
	}
	now := s.now()
	if due != nil {
		if err := due.CheckNotBefore(domain.DueDateOf(now.In(s.loc))); err != nil {
			return IssuedToken{}, err
		}
	}
	exp := now.Add(s.ttl).Truncate(time.Second)
	jti := newJTI()
	token, err := s.signer.Sign(TokenClaims{
		JTI:       jti,
		From:      from,
		Item:      string(it),
		Qty:       int(q),
		DueDate:   due,
		ExpiresAt: exp,
	})
	if err != nil {
		return IssuedToken{}, err
	}
	return IssuedToken{Token: token, ExpiresAt: exp, JTI: jti}, nil
}

// Receive は受け取る側（to）がトークンを読み取ったときに、受け渡しを記録する。
func (s *Service) Receive(ctx context.Context, to uuid.UUID, token string) (HandoffView, error) {
	now := s.now()
	c, err := s.signer.Verify(token, now)
	if err != nil {
		return HandoffView{}, err
	}
	item, err := domain.NewItem(c.Item)
	if err != nil {
		return HandoffView{}, ErrInvalidToken
	}
	qty, err := domain.NewQuantity(c.Qty)
	if err != nil {
		return HandoffView{}, ErrInvalidToken
	}
	h, err := domain.NewHandoff(uuid.New(), c.JTI, c.From, to, item, qty, c.DueDate, now)
	if err != nil {
		return HandoffView{}, err
	}
	if err := s.repo.Save(ctx, h); err != nil {
		return HandoffView{}, err
	}
	return s.toView(ctx, h)
}

// IssueReturnToken は借りた側（borrower）が、返却用の QR トークンを発行する。
func (s *Service) IssueReturnToken(ctx context.Context, borrower, handoffID uuid.UUID) (IssuedToken, error) {
	h, err := s.repo.Get(ctx, handoffID)
	if err != nil {
		return IssuedToken{}, err
	}
	if err := h.CheckReturnableBy(borrower); err != nil {
		return IssuedToken{}, err
	}
	exp := s.now().Add(s.ttl).Truncate(time.Second)
	token, err := s.signer.SignReturn(ReturnClaims{HandoffID: h.ID(), ExpiresAt: exp})
	if err != nil {
		return IssuedToken{}, err
	}
	return IssuedToken{Token: token, ExpiresAt: exp}, nil
}

// Return は貸した側（lender）が返却用のトークンを読み取ったときに、返却を記録する。
func (s *Service) Return(ctx context.Context, lender uuid.UUID, token string) (HandoffView, error) {
	now := s.now()
	c, err := s.signer.VerifyReturn(token, now)
	if err != nil {
		return HandoffView{}, err
	}
	h, err := s.repo.Get(ctx, c.HandoffID)
	if err != nil {
		return HandoffView{}, err
	}
	if err := h.MarkReturned(lender, now); err != nil {
		return HandoffView{}, err
	}
	if err := s.repo.SaveReturn(ctx, h); err != nil {
		return HandoffView{}, err
	}
	return s.toView(ctx, h)
}

// RequestReturns は借りた側（borrower）が、対面でない返却をまとめて送る。貸した側の承認待ちになる。
// 1 件でも送れないものがあれば、どれも記録しない。
func (s *Service) RequestReturns(ctx context.Context, borrower uuid.UUID, ids []uuid.UUID) ([]HandoffView, error) {
	now := s.now()
	return s.updateAll(ctx, ids,
		func(h *domain.Handoff) error { return h.RequestReturn(borrower, now) },
		s.repo.SaveReturnRequests,
	)
}

// ConfirmReturns は貸した側（lender）が、送られた返却をまとめて承認し、返却済みにする。
// 1 件でも承認できないものがあれば、どれも記録しない。
func (s *Service) ConfirmReturns(ctx context.Context, lender uuid.UUID, ids []uuid.UUID) ([]HandoffView, error) {
	now := s.now()
	return s.updateAll(ctx, ids,
		func(h *domain.Handoff) error { return h.ConfirmReturn(lender, now) },
		s.repo.SaveConfirmedReturns,
	)
}

func (s *Service) updateAll(ctx context.Context, ids []uuid.UUID, change func(*domain.Handoff) error, save func(context.Context, []domain.Handoff) error) ([]HandoffView, error) {
	ids = uniqueIDs(ids)
	if len(ids) == 0 || len(ids) > MaxBatch {
		return nil, ErrInvalidSelection
	}
	hs, err := s.repo.GetMany(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range hs {
		if err := change(&hs[i]); err != nil {
			return nil, err
		}
	}
	if err := save(ctx, hs); err != nil {
		return nil, err
	}
	return s.toViews(ctx, hs)
}

func uniqueIDs(ids []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(ids))
	out := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; !ok {
			seen[id] = struct{}{}
			out = append(out, id)
		}
	}
	return out
}

// ListMine は user が渡した、または受け取った記録を新しい順に返す。
func (s *Service) ListMine(ctx context.Context, user uuid.UUID, limit int) ([]HandoffView, error) {
	hs, err := s.repo.ListByUser(ctx, user, limit)
	if err != nil {
		return nil, err
	}
	return s.toViews(ctx, hs)
}

// ListOutstanding は user が貸している・借りている（返却期限があってまだ返却されていない）記録を、期限が近い順に返す。
func (s *Service) ListOutstanding(ctx context.Context, user uuid.UUID) ([]HandoffView, error) {
	hs, err := s.repo.ListOutstandingByUser(ctx, user)
	if err != nil {
		return nil, err
	}
	return s.toViews(ctx, hs)
}

func (s *Service) toView(ctx context.Context, h domain.Handoff) (HandoffView, error) {
	views, err := s.toViews(ctx, []domain.Handoff{h})
	if err != nil {
		return HandoffView{}, err
	}
	return views[0], nil
}

func (s *Service) toViews(ctx context.Context, hs []domain.Handoff) ([]HandoffView, error) {
	idSet := make(map[uuid.UUID]struct{}, len(hs)*2)
	for _, h := range hs {
		idSet[h.From()] = struct{}{}
		idSet[h.To()] = struct{}{}
	}
	ids := make([]uuid.UUID, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}
	names, err := s.users.DisplayNames(ctx, ids)
	if err != nil {
		return nil, err
	}
	views := make([]HandoffView, len(hs))
	for i, h := range hs {
		v := HandoffView{
			ID:                h.ID(),
			JTI:               h.JTI(),
			From:              UserRef{ID: h.From(), DisplayName: names[h.From()]},
			To:                UserRef{ID: h.To(), DisplayName: names[h.To()]},
			Item:              string(h.Item()),
			Qty:               int(h.Qty()),
			CreatedAt:         h.CreatedAt(),
			ReturnedAt:        h.ReturnedAt(),
			ReturnRequestedAt: h.ReturnRequestedAt(),
		}
		if d := h.DueDate(); d != nil {
			t := d.Time()
			v.DueDate = &t
		}
		views[i] = v
	}
	return views, nil
}

// newJTI はトークンの ID（8 バイトの乱数を 16 進数にしたもの）を作る。
func newJTI() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b) // crypto/rand.Read は失敗しない
	return hex.EncodeToString(b)
}
