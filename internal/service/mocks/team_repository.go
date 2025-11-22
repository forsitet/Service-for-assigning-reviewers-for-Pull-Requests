package mocks

import (
	"context"

	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/internal/domain"
)

type MockTeamRepository struct {
	ExistsResult         bool
	ExistsErr            error
	CreateErr            error
	GetWithMembersResult *domain.Team
	GetWithMembersErr    error
}

func (m *MockTeamRepository) Create(ctx context.Context, name string) error {
	return m.CreateErr
}

func (m *MockTeamRepository) Exists(ctx context.Context, name string) (bool, error) {
	return m.ExistsResult, m.ExistsErr
}

func (m *MockTeamRepository) GetWithMembers(ctx context.Context, name string) (*domain.Team, error) {
	return m.GetWithMembersResult, m.GetWithMembersErr
}

