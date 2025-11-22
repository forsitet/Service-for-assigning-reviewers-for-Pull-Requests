package service

import (
	"context"
	"errors"
	"testing"

	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/internal/service/mocks"
)

func TestStatsService_GetAssignmentStats(t *testing.T) {
	tests := []struct {
		name              string
		mockByReviewer    map[string]int64
		mockByReviewerErr error
		mockByPR          map[string]int64
		mockByPRErr       error
		wantErr           bool
		validateResult    func(t *testing.T, byUser, byPR map[string]int64)
	}{
		{
			name: "успешное получение статистики",
			mockByReviewer: map[string]int64{
				"user-1": 5,
				"user-2": 3,
				"user-3": 7,
			},
			mockByPR: map[string]int64{
				"pr-1": 2,
				"pr-2": 2,
				"pr-3": 2,
			},
			wantErr: false,
			validateResult: func(t *testing.T, byUser, byPR map[string]int64) {
				if len(byUser) != 3 {
					t.Errorf("expected 3 users in stats, got %d", len(byUser))
				}
				if byUser["user-1"] != 5 {
					t.Errorf("expected user-1 to have 5 assignments, got %d", byUser["user-1"])
				}
				if byUser["user-2"] != 3 {
					t.Errorf("expected user-2 to have 3 assignments, got %d", byUser["user-2"])
				}
				if byUser["user-3"] != 7 {
					t.Errorf("expected user-3 to have 7 assignments, got %d", byUser["user-3"])
				}

				if len(byPR) != 3 {
					t.Errorf("expected 3 PRs in stats, got %d", len(byPR))
				}
				if byPR["pr-1"] != 2 {
					t.Errorf("expected pr-1 to have 2 assignments, got %d", byPR["pr-1"])
				}
			},
		},
		{
			name:           "пустая статистика",
			mockByReviewer: map[string]int64{},
			mockByPR:       map[string]int64{},
			wantErr:        false,
			validateResult: func(t *testing.T, byUser, byPR map[string]int64) {
				if len(byUser) != 0 {
					t.Errorf("expected 0 users in stats, got %d", len(byUser))
				}
				if len(byPR) != 0 {
					t.Errorf("expected 0 PRs in stats, got %d", len(byPR))
				}
			},
		},
		{
			name:              "ошибка при получении статистики по ревьюерам",
			mockByReviewerErr: errors.New("database error"),
			wantErr:           true,
		},
		{
			name: "ошибка при получении статистики по PR",
			mockByReviewer: map[string]int64{
				"user-1": 5,
			},
			mockByPRErr: errors.New("database error"),
			wantErr:     true,
		},
		{
			name: "статистика с нулевыми значениями",
			mockByReviewer: map[string]int64{
				"user-1": 0,
				"user-2": 0,
			},
			mockByPR: map[string]int64{
				"pr-1": 0,
			},
			wantErr: false,
			validateResult: func(t *testing.T, byUser, byPR map[string]int64) {
				if byUser["user-1"] != 0 {
					t.Errorf("expected user-1 to have 0 assignments, got %d", byUser["user-1"])
				}
				if byPR["pr-1"] != 0 {
					t.Errorf("expected pr-1 to have 0 assignments, got %d", byPR["pr-1"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.MockAssignmentStatsRepo{
				CountByReviewerResult: tt.mockByReviewer,
				CountByReviewerErr:    tt.mockByReviewerErr,
				CountByPRResult:       tt.mockByPR,
				CountByPRErr:          tt.mockByPRErr,
			}

			service := NewStatsService(mockRepo)
			ctx := context.Background()

			byUser, byPR, err := service.GetAssignmentStats(ctx)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if tt.validateResult != nil {
					tt.validateResult(t, byUser, byPR)
				}
			}
		})
	}
}
