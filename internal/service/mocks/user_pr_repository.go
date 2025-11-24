package mocks

import (
	"context"

	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/internal/domain"
)

type MockUserPRRepository struct {
	ListByReviewerResult []domain.PullRequest
	ListByReviewerErr    error
}

func (m *MockUserPRRepository) ListByReviewer(ctx context.Context, userID string) ([]domain.PullRequest, error) {
	return m.ListByReviewerResult, m.ListByReviewerErr
}
