package mocks

import (
	"context"

	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/internal/domain"
)

type MockTeamUserRepository struct {
	UpsertErr error
}

func (m *MockTeamUserRepository) UpsertForTeam(ctx context.Context, teamName string, users []domain.User) error {
	return m.UpsertErr
}

