package mocks

import (
	"context"
	"time"

	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/internal/domain"
)

type MockPRRepository struct {
	ExistsResult          bool
	ExistsErr             error
	CreateErr             error
	GetByIDResult         *domain.PullRequest
	GetByIDReviewers      []string
	GetByIDErr            error
	SetMergedResult       *domain.PullRequest
	SetMergedReviewers    []string
	SetMergedErr          error
	UpdateResult          *domain.PullRequest
	UpdateReviewersResult []string
	UpdateErr             error
	DeactivateResult      domain.TeamDeactivationResult
	DeactivateErr         error
}

func (m *MockPRRepository) CreateWithReviewers(ctx context.Context, pr *domain.PullRequest, reviewerIDs []string) error {
	return m.CreateErr
}

func (m *MockPRRepository) GetByID(ctx context.Context, id string) (*domain.PullRequest, []string, error) {
	return m.GetByIDResult, m.GetByIDReviewers, m.GetByIDErr
}

func (m *MockPRRepository) SetMerged(ctx context.Context, id string, mergedAt time.Time) (*domain.PullRequest, []string, error) {
	return m.SetMergedResult, m.SetMergedReviewers, m.SetMergedErr
}

func (m *MockPRRepository) UpdateReviewers(ctx context.Context, id string, reviewerIDs []string) (*domain.PullRequest, []string, error) {
	return m.UpdateResult, m.UpdateReviewersResult, m.UpdateErr
}

func (m *MockPRRepository) ListByReviewer(ctx context.Context, userID string) ([]domain.PullRequest, error) {
	return nil, nil
}

func (m *MockPRRepository) Exists(ctx context.Context, id string) (bool, error) {
	return m.ExistsResult, m.ExistsErr
}

func (m *MockPRRepository) DeactivateTeamAndReassignOpenPRs(ctx context.Context, teamName string) (domain.TeamDeactivationResult, error) {
	return m.DeactivateResult, m.DeactivateErr
}
