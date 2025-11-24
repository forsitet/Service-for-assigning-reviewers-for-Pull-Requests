package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/internal/domain"
	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/internal/service/mocks"
)

func TestPRService_CreatePullRequest(t *testing.T) {
	tests := []struct {
		name               string
		id                 string
		prName             string
		authorID           string
		mockPRExists       bool
		mockPRExistsErr    error
		mockAuthor         *domain.User
		mockAuthorErr      error
		mockTeamMembers    []domain.User
		mockTeamMembersErr error
		mockCreateErr      error
		nowFunc            func() time.Time
		wantErr            bool
		wantErrCode        domain.ErrorCode
		validateResult     func(t *testing.T, pr *domain.PullRequest)
	}{
		{
			name:         "успешное создание PR",
			id:           "pr-1",
			prName:       "Test PR",
			authorID:     "user-1",
			mockPRExists: false,
			mockAuthor: &domain.User{
				ID:       "user-1",
				Username: "author",
				TeamName: "team-1",
				IsActive: true,
			},
			mockTeamMembers: []domain.User{
				{ID: "user-2", Username: "reviewer1", TeamName: "team-1", IsActive: true},
				{ID: "user-3", Username: "reviewer2", TeamName: "team-1", IsActive: true},
			},
			nowFunc: func() time.Time { return time.Unix(1000, 0) },
			wantErr: false,
			validateResult: func(t *testing.T, pr *domain.PullRequest) {
				if pr.ID != "pr-1" {
					t.Errorf("expected ID pr-1, got %s", pr.ID)
				}
				if pr.Name != "Test PR" {
					t.Errorf("expected Name Test PR, got %s", pr.Name)
				}
				if pr.AuthorID != "user-1" {
					t.Errorf("expected AuthorID user-1, got %s", pr.AuthorID)
				}
				if pr.Status != domain.PRStatusOpen {
					t.Errorf("expected Status OPEN, got %s", pr.Status)
				}
				if len(pr.AssignedReviewers) != 2 {
					t.Errorf("expected 2 reviewers, got %d", len(pr.AssignedReviewers))
				}
				if pr.CreatedAt != 1000 {
					t.Errorf("expected CreatedAt 1000, got %d", pr.CreatedAt)
				}
			},
		},
		{
			name:         "PR уже существует",
			id:           "pr-1",
			prName:       "Test PR",
			authorID:     "user-1",
			mockPRExists: true,
			wantErr:      true,
			wantErrCode:  domain.ErrorCodePRExists,
		},
		{
			name:            "ошибка при проверке существования PR",
			id:              "pr-1",
			prName:          "Test PR",
			authorID:        "user-1",
			mockPRExistsErr: errors.New("database error"),
			wantErr:         true,
		},
		{
			name:          "автор не найден",
			id:            "pr-1",
			prName:        "Test PR",
			authorID:      "user-1",
			mockPRExists:  false,
			mockAuthorErr: errors.New("user not found"),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPRRepo := &mocks.MockPRRepository{
				ExistsResult: tt.mockPRExists,
				ExistsErr:    tt.mockPRExistsErr,
				CreateErr:    tt.mockCreateErr,
			}
			mockUserRepo := &mocks.MockPRUserRepository{
				GetByIDResult:    tt.mockAuthor,
				GetByIDErr:       tt.mockAuthorErr,
				ListByTeamResult: tt.mockTeamMembers,
				ListByTeamErr:    tt.mockTeamMembersErr,
			}

			nowFunc := tt.nowFunc
			if nowFunc == nil {
				nowFunc = time.Now
			}

			service := NewPRService(mockPRRepo, mockUserRepo, nowFunc)
			ctx := context.Background()

			result, err := service.CreatePullRequest(ctx, tt.id, tt.prName, tt.authorID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
					return
				}
				if tt.wantErrCode != "" {
					var domainErr *domain.DomainError
					if errors.As(err, &domainErr) {
						if domainErr.Code != tt.wantErrCode {
							t.Errorf("expected error code %s, got %s", tt.wantErrCode, domainErr.Code)
						}
					} else {
						t.Errorf("expected domain error, got %T", err)
					}
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

func TestPRService_MergePullRequest(t *testing.T) {
	tests := []struct {
		name                   string
		id                     string
		mockPR                 *domain.PullRequest
		mockReviewers          []string
		mockGetErr             error
		mockSetMergedPR        *domain.PullRequest
		mockSetMergedReviewers []string
		mockSetMergedErr       error
		nowFunc                func() time.Time
		wantErr                bool
		validateResult         func(t *testing.T, pr *domain.PullRequest)
	}{
		{
			name: "успешное мерджирование PR",
			id:   "pr-1",
			mockPR: &domain.PullRequest{
				ID:     "pr-1",
				Name:   "Test PR",
				Status: domain.PRStatusOpen,
			},
			mockReviewers: []string{"user-2", "user-3"},
			mockSetMergedPR: &domain.PullRequest{
				ID:     "pr-1",
				Name:   "Test PR",
				Status: domain.PRStatusMerged,
			},
			mockSetMergedReviewers: []string{"user-2", "user-3"},
			nowFunc:                func() time.Time { return time.Unix(2000, 0) },
			wantErr:                false,
			validateResult: func(t *testing.T, pr *domain.PullRequest) {
				if pr.Status != domain.PRStatusMerged {
					t.Errorf("expected Status MERGED, got %s", pr.Status)
				}
				if len(pr.AssignedReviewers) != 2 {
					t.Errorf("expected 2 reviewers, got %d", len(pr.AssignedReviewers))
				}
			},
		},
		{
			name: "PR уже смерджен",
			id:   "pr-1",
			mockPR: &domain.PullRequest{
				ID:     "pr-1",
				Name:   "Test PR",
				Status: domain.PRStatusMerged,
			},
			mockReviewers: []string{"user-2", "user-3"},
			nowFunc:       func() time.Time { return time.Unix(2000, 0) },
			wantErr:       false,
			validateResult: func(t *testing.T, pr *domain.PullRequest) {
				if pr.Status != domain.PRStatusMerged {
					t.Errorf("expected Status MERGED, got %s", pr.Status)
				}
			},
		},
		{
			name:       "PR не найден",
			id:         "pr-1",
			mockGetErr: errors.New("PR not found"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPRRepo := &mocks.MockPRRepository{
				GetByIDResult:      tt.mockPR,
				GetByIDReviewers:   tt.mockReviewers,
				GetByIDErr:         tt.mockGetErr,
				SetMergedResult:    tt.mockSetMergedPR,
				SetMergedReviewers: tt.mockSetMergedReviewers,
				SetMergedErr:       tt.mockSetMergedErr,
			}
			mockUserRepo := &mocks.MockPRUserRepository{}

			nowFunc := tt.nowFunc
			if nowFunc == nil {
				nowFunc = time.Now
			}

			service := NewPRService(mockPRRepo, mockUserRepo, nowFunc)
			ctx := context.Background()

			result, err := service.MergePullRequest(ctx, tt.id)

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

func TestPRService_ReassignReviewer(t *testing.T) {
	tests := []struct {
		name                 string
		prID                 string
		oldReviewerID        string
		mockPR               *domain.PullRequest
		mockReviewers        []string
		mockGetErr           error
		mockOldReviewer      *domain.User
		mockOldReviewerErr   error
		mockTeamMembers      []domain.User
		mockTeamMembersErr   error
		mockUpdatedPR        *domain.PullRequest
		mockUpdatedReviewers []string
		mockUpdateErr        error
		wantErr              bool
		wantErrCode          domain.ErrorCode
		validateResult       func(t *testing.T, pr *domain.PullRequest)
	}{
		{
			name:          "PR уже смерджен",
			prID:          "pr-1",
			oldReviewerID: "user-2",
			mockPR: &domain.PullRequest{
				ID:     "pr-1",
				Name:   "Test PR",
				Status: domain.PRStatusMerged,
			},
			mockReviewers: []string{"user-2", "user-3"},
			wantErr:       true,
			wantErrCode:   domain.ErrorCodePRMerged,
		},
		{
			name:          "ревьюер не назначен на PR",
			prID:          "pr-1",
			oldReviewerID: "user-5",
			mockPR: &domain.PullRequest{
				ID:     "pr-1",
				Name:   "Test PR",
				Status: domain.PRStatusOpen,
			},
			mockReviewers: []string{"user-2", "user-3"},
			wantErr:       true,
			wantErrCode:   domain.ErrorCodeNotAssigned,
		},
		{
			name:          "нет доступных кандидатов",
			prID:          "pr-1",
			oldReviewerID: "user-2",
			mockPR: &domain.PullRequest{
				ID:       "pr-1",
				Name:     "Test PR",
				Status:   domain.PRStatusOpen,
				AuthorID: "user-1",
			},
			mockReviewers: []string{"user-2", "user-3"},
			mockOldReviewer: &domain.User{
				ID:       "user-2",
				Username: "reviewer1",
				TeamName: "team-1",
				IsActive: true,
			},
			mockTeamMembers: []domain.User{
				{ID: "user-1", Username: "author", TeamName: "team-1", IsActive: true},
				{ID: "user-2", Username: "reviewer1", TeamName: "team-1", IsActive: true},
				{ID: "user-3", Username: "reviewer2", TeamName: "team-1", IsActive: true},
			},
			wantErr:     true,
			wantErrCode: domain.ErrorCodeNoCandidate,
		},
		{
			name:          "PR не найден",
			prID:          "pr-1",
			oldReviewerID: "user-2",
			mockGetErr:    errors.New("PR not found"),
			wantErr:       true,
		},
		{
			name:          "старый ревьюер не найден",
			prID:          "pr-1",
			oldReviewerID: "user-2",
			mockPR: &domain.PullRequest{
				ID:     "pr-1",
				Name:   "Test PR",
				Status: domain.PRStatusOpen,
			},
			mockReviewers:      []string{"user-2", "user-3"},
			mockOldReviewerErr: errors.New("user not found"),
			wantErr:            true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPRRepo := &mocks.MockPRRepository{
				GetByIDResult:         tt.mockPR,
				GetByIDReviewers:      tt.mockReviewers,
				GetByIDErr:            tt.mockGetErr,
				UpdateResult:          tt.mockUpdatedPR,
				UpdateReviewersResult: tt.mockUpdatedReviewers,
				UpdateErr:             tt.mockUpdateErr,
			}
			mockUserRepo := &mocks.MockPRUserRepository{
				GetByIDResult:    tt.mockOldReviewer,
				GetByIDErr:       tt.mockOldReviewerErr,
				ListByTeamResult: tt.mockTeamMembers,
				ListByTeamErr:    tt.mockTeamMembersErr,
			}

			service := NewPRService(mockPRRepo, mockUserRepo, time.Now)
			ctx := context.Background()

			result, err := service.ReassignReviewer(ctx, tt.prID, tt.oldReviewerID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
					return
				}
				if tt.wantErrCode != "" {
					var domainErr *domain.DomainError
					if errors.As(err, &domainErr) {
						if domainErr.Code != tt.wantErrCode {
							t.Errorf("expected error code %s, got %s", tt.wantErrCode, domainErr.Code)
						}
					} else {
						t.Errorf("expected domain error, got %T", err)
					}
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

func TestSelectInitialReviewers(t *testing.T) {
	tests := []struct {
		name         string
		authorID     string
		members      []domain.User
		wantCount    int
		wantMaxCount int
	}{
		{
			name:     "выбор 2 ревьюеров из 3 доступных",
			authorID: "user-1",
			members: []domain.User{
				{ID: "user-1", IsActive: true},
				{ID: "user-2", IsActive: true},
				{ID: "user-3", IsActive: true},
			},
			wantCount:    2,
			wantMaxCount: 2,
		},
		{
			name:     "только автор в команде",
			authorID: "user-1",
			members: []domain.User{
				{ID: "user-1", IsActive: true},
			},
			wantCount:    0,
			wantMaxCount: 0,
		},
		{
			name:     "исключение неактивных пользователей",
			authorID: "user-1",
			members: []domain.User{
				{ID: "user-1", IsActive: true},
				{ID: "user-2", IsActive: false},
				{ID: "user-3", IsActive: true},
			},
			wantCount:    1,
			wantMaxCount: 1,
		},
		{
			name:     "меньше кандидатов чем требуется",
			authorID: "user-1",
			members: []domain.User{
				{ID: "user-1", IsActive: true},
				{ID: "user-2", IsActive: true},
			},
			wantCount:    1,
			wantMaxCount: 1,
		},
		{
			name:         "пустой список членов команлы",
			authorID:     "user-1",
			members:      []domain.User{},
			wantCount:    0,
			wantMaxCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := selectInitialReviewers(tt.authorID, tt.members)
			if len(result) < tt.wantCount || len(result) > tt.wantMaxCount {
				t.Errorf("expected %d-%d reviewers, got %d", tt.wantCount, tt.wantMaxCount, len(result))
			}
			for _, reviewerID := range result {
				if reviewerID == tt.authorID {
					t.Errorf("author should not be in reviewers list")
				}
			}
		})
	}
}
