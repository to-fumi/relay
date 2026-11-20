package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"relay/internal/handoff/domain"

	"github.com/google/uuid"
)

type Service struct {
	repo   domain.Repository
	signer TokenSigner
	users  UserDirectory
	ttl    time.Duration
	now    func() time.Time
}

func NewService(repo domain.Repository, signer TokenSigner, users UserDirectory, ttl time.Duration, now func() time.Time) *Service {
	return &Service{repo: repo, signer: signer, users: users, ttl: ttl, now: now}
}

type IssuedToken struct {
	Token     string
	ExpiresAt time.Time
}

// HandoffView は表示用の受け渡し記録。
type HandoffView struct {
	ID        uuid.UUID
	From      UserRef
	To        UserRef
	Item      string
	Qty       int
	CreatedAt time.Time
}

type UserRef struct {
	ID          uuid.UUID
	DisplayName string
}

// IssueToken は渡す側（from）の QR トークンを発行する。トークンはどこにも保存しない。
func (s *Service) IssueToken(_ context.Context, from uuid.UUID, item string, qty int) (IssuedToken, error) {
	it, err := domain.NewItem(item)
	if err != nil {
		return IssuedToken{}, err
	}
	q, err := domain.NewQuantity(qty)
	if err != nil {
		return IssuedToken{}, err
	}
	exp := s.now().Add(s.ttl).Truncate(time.Second)
	token, err := s.signer.Sign(TokenClaims{
		JTI:       newJTI(),
		From:      from,
		Item:      string(it),
		Qty:       int(q),
		ExpiresAt: exp,
	})
	if err != nil {
		return IssuedToken{}, err
	}
	return IssuedToken{Token: token, ExpiresAt: exp}, nil
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
	h, err := domain.NewHandoff(uuid.New(), c.JTI, c.From, to, item, qty, now)
	if err != nil {
		return HandoffView{}, err
	}
	if err := s.repo.Save(ctx, h); err != nil {
		return HandoffView{}, err
	}
	views, err := s.toViews(ctx, []domain.Handoff{h})
	if err != nil {
		return HandoffView{}, err
	}
	return views[0], nil
}

// ListMine は user が渡した、または受け取った記録を新しい順に返す。
func (s *Service) ListMine(ctx context.Context, user uuid.UUID, limit int) ([]HandoffView, error) {
	hs, err := s.repo.ListByUser(ctx, user, limit)
	if err != nil {
		return nil, err
	}
	return s.toViews(ctx, hs)
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
		views[i] = HandoffView{
			ID:        h.ID(),
			From:      UserRef{ID: h.From(), DisplayName: names[h.From()]},
			To:        UserRef{ID: h.To(), DisplayName: names[h.To()]},
			Item:      string(h.Item()),
			Qty:       int(h.Qty()),
			CreatedAt: h.CreatedAt(),
		}
	}
	return views, nil
}

// newJTI はトークンの ID（8 バイトの乱数を 16 進数にしたもの）を作る。
func newJTI() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b) // crypto/rand.Read は失敗しない
	return hex.EncodeToString(b)
}
