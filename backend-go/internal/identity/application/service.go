// Package application は identity モジュールのユースケース。domain だけに依存する。
package application

import (
	"context"
	"strings"

	"relay/internal/identity/domain"

	"github.com/google/uuid"
)

// Identity はログインで確かめた利用者の情報（platform/auth の Principal から変換して渡す）。
type Identity struct {
	Issuer  string
	Subject string
	Email   string
	Name    string
}

type Service struct {
	repo        domain.Repository
	adminEmails map[string]struct{}
}

func NewService(repo domain.Repository, adminEmails []string) *Service {
	admins := make(map[string]struct{}, len(adminEmails))
	for _, e := range adminEmails {
		admins[strings.ToLower(e)] = struct{}{}
	}
	return &Service{repo: repo, adminEmails: admins}
}

// EnsureUser はログインした利用者をユーザーとして登録（初回）または更新する。
// 名前がなければ、メールアドレスの @ より前を名前にする。
func (s *Service) EnsureUser(ctx context.Context, id Identity) (domain.User, error) {
	_, isAdmin := s.adminEmails[strings.ToLower(id.Email)]
	name := strings.TrimSpace(id.Name)
	if name == "" {
		name, _, _ = strings.Cut(id.Email, "@")
	}
	return s.repo.UpsertBySubject(ctx, domain.User{
		ID:          uuid.New(),
		Issuer:      id.Issuer,
		Subject:     id.Subject,
		DisplayName: name,
		Email:       id.Email,
		IsAdmin:     isAdmin,
	})
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (domain.User, error) {
	return s.repo.FindByID(ctx, id)
}

// DisplayNames はユーザーIDから表示名を引く（handoff.application.UserDirectory を満たす）。
func (s *Service) DisplayNames(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]string, error) {
	users, err := s.repo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := make(map[uuid.UUID]string, len(users))
	for _, u := range users {
		out[u.ID] = u.DisplayName
	}
	return out, nil
}
