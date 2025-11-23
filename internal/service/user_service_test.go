package service

import (
	"context"
	"errors"
	"testing"

	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/internal/domain"
	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/internal/service/mocks"
)

func TestUserService_SetActive(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		active        bool
		mockUser      *domain.User
		mockSetErr    error
		wantErr       bool
		validateResult func(t *testing.T, user *domain.User)
	}{
		{
			name:   "успешная активация пользователя",
			userID: "user-1",
			active: true,
			mockUser: &domain.User{
				ID:       "user-1",
				Username: "testuser",
				TeamName: "team-1",
				IsActive: true,
			},
			wantErr: false,
			validateResult: func(t *testing.T, user *domain.User) {
				if user.ID != "user-1" {
					t.Errorf("expected ID user-1, got %s", user.ID)
				}
				if !user.IsActive {
					t.Errorf("expected IsActive true, got false")
				}
			},
		},
		{
			name:   "успешная деактивация пользователя",
			userID: "user-1",
			active: false,
			mockUser: &domain.User{
				ID:       "user-1",
				Username: "testuser",
				TeamName: "team-1",
				IsActive: false,
			},
			wantErr: false,
			validateResult: func(t *testing.T, user *domain.User) {
				if user.IsActive {
					t.Errorf("expected IsActive false, got true")
				}
			},
		},
		{
			name:      "ошибка при установке статуса",
			userID:    "user-1",
			active:    true,
			mockSetErr: errors.New("database error"),
			wantErr:   true,
		},
		{
			name:      "пользователь не найден",
			userID:    "user-999",
			active:    true,
			mockSetErr: errors.New("user not found"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := &mocks.MockUserRepository{
				SetIsActiveResult: tt.mockUser,
				SetIsActiveErr:    tt.mockSetErr,
			}
			mockPRRepo := &mocks.MockUserPRRepository{}

			service := NewUserService(mockUserRepo, mockPRRepo)
			ctx := context.Background()

			result, err := service.SetActive(ctx, tt.userID, tt.active)

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
					tt.validateResult(t, result)
				}
			}
		})
	}
}

func TestUserService_ListAssignedPullRequests(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		mockPRs       []domain.PullRequest
		mockListErr   error
		wantErr       bool
		validateResult func(t *testing.T, prs []domain.PullRequest)
	}{
		{
			name:   "успешное получение списка PR",
			userID: "user-1",
			mockPRs: []domain.PullRequest{
				{
					ID:                "pr-1",
					Name:              "PR 1",
					AuthorID:          "user-2",
					Status:            domain.PRStatusOpen,
					AssignedReviewers: []string{"user-1", "user-3"},
				},
				{
					ID:                "pr-2",
					Name:              "PR 2",
					AuthorID:          "user-3",
					Status:            domain.PRStatusOpen,
					AssignedReviewers: []string{"user-1"},
				},
			},
			wantErr: false,
			validateResult: func(t *testing.T, prs []domain.PullRequest) {
				if len(prs) != 2 {
					t.Errorf("expected 2 PRs, got %d", len(prs))
				}
				for _, pr := range prs {
					found := false
					for _, reviewer := range pr.AssignedReviewers {
						if reviewer == "user-1" {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected user-1 in reviewers for PR %s", pr.ID)
					}
				}
			},
		},
		{
			name:   "пустой список PR",
			userID: "user-1",
			mockPRs: []domain.PullRequest{},
			wantErr: false,
			validateResult: func(t *testing.T, prs []domain.PullRequest) {
				if len(prs) != 0 {
					t.Errorf("expected 0 PRs, got %d", len(prs))
				}
			},
		},
		{
			name:        "ошибка при получении списка",
			userID:      "user-1",
			mockListErr: errors.New("database error"),
			wantErr:     true,
		},
		{
			name:   "список с мердженными PR",
			userID: "user-1",
			mockPRs: []domain.PullRequest{
				{
					ID:                "pr-1",
					Name:              "PR 1",
					AuthorID:          "user-2",
					Status:            domain.PRStatusMerged,
					AssignedReviewers: []string{"user-1"},
				},
				{
					ID:                "pr-2",
					Name:              "PR 2",
					AuthorID:          "user-3",
					Status:            domain.PRStatusOpen,
					AssignedReviewers: []string{"user-1"},
				},
			},
			wantErr: false,
			validateResult: func(t *testing.T, prs []domain.PullRequest) {
				if len(prs) != 2 {
					t.Errorf("expected 2 PRs, got %d", len(prs))
				}
				mergedCount := 0
				openCount := 0
				for _, pr := range prs {
					if pr.Status == domain.PRStatusMerged {
						mergedCount++
					} else if pr.Status == domain.PRStatusOpen {
						openCount++
					}
				}
				if mergedCount != 1 || openCount != 1 {
					t.Errorf("expected 1 merged and 1 open PR, got %d merged and %d open", mergedCount, openCount)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := &mocks.MockUserRepository{}
			mockPRRepo := &mocks.MockUserPRRepository{
				ListByReviewerResult: tt.mockPRs,
				ListByReviewerErr:    tt.mockListErr,
			}

			service := NewUserService(mockUserRepo, mockPRRepo)
			ctx := context.Background()

			result, err := service.ListAssignedPullRequests(ctx, tt.userID)

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
					tt.validateResult(t, result)
				}
			}
		})
	}
}

