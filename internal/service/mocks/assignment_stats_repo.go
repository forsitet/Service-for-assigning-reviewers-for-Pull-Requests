package mocks

import (
	"context"
)

type MockAssignmentStatsRepo struct {
	CountByReviewerResult map[string]int64
	CountByReviewerErr    error
	CountByPRResult       map[string]int64
	CountByPRErr          error
}

func (m *MockAssignmentStatsRepo) CountAssignmentsByReviewer(ctx context.Context) (map[string]int64, error) {
	return m.CountByReviewerResult, m.CountByReviewerErr
}

func (m *MockAssignmentStatsRepo) CountAssignmentsByPR(ctx context.Context) (map[string]int64, error) {
	return m.CountByPRResult, m.CountByPRErr
}

