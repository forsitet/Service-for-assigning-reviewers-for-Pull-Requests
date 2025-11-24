package mocks

import (
	"context"

	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/internal/domain"
)

type MockUserRepository struct {
	GetByIDResult     *domain.User
	GetByIDErr        error
	SetIsActiveResult *domain.User
	SetIsActiveErr    error
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return m.GetByIDResult, m.GetByIDErr
}

func (m *MockUserRepository) SetIsActive(ctx context.Context, id string, active bool) (*domain.User, error) {
	return m.SetIsActiveResult, m.SetIsActiveErr
}

