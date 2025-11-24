package mocks

import (
	"context"

	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/internal/domain"
)

type MockPRUserRepository struct {
	GetByIDResult    *domain.User
	GetByIDErr       error
	ListByTeamResult []domain.User
	ListByTeamErr    error
}

func (m *MockPRUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return m.GetByIDResult, m.GetByIDErr
}

func (m *MockPRUserRepository) ListByTeam(ctx context.Context, teamName string) ([]domain.User, error) {
	return m.ListByTeamResult, m.ListByTeamErr
}
