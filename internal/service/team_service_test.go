package service

import (
	"context"
	"errors"
	"testing"

	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/internal/domain"
	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/internal/service/mocks"
)

func TestTeamService_CreateTeam(t *testing.T) {
	tests := []struct {
		name           string
		teamName       string
		members        []domain.User
		mockExists     bool
		mockExistsErr  error
		mockCreateErr  error
		mockUpsertErr  error
		wantErr        bool
		wantErrCode    domain.ErrorCode
		validateResult func(t *testing.T, team *domain.Team)
	}{
		{
			name:     "успешное создание команды",
			teamName: "team-1",
			members: []domain.User{
				{ID: "user-1", Username: "user1", IsActive: true},
				{ID: "user-2", Username: "user2", IsActive: true},
			},
			mockExists: false,
			wantErr:    false,
			validateResult: func(t *testing.T, team *domain.Team) {
				if team.Name != "team-1" {
					t.Errorf("expected team name team-1, got %s", team.Name)
				}
				if len(team.Members) != 2 {
					t.Errorf("expected 2 members, got %d", len(team.Members))
				}
				for _, member := range team.Members {
					if member.TeamName != "team-1" {
						t.Errorf("expected team name team-1 for member, got %s", member.TeamName)
					}
				}
			},
		},
		{
			name:     "команда уже существует",
			teamName: "team-1",
			members: []domain.User{
				{ID: "user-1", Username: "user1", IsActive: true},
			},
			mockExists:  true,
			wantErr:     true,
			wantErrCode: domain.ErrorCodeTeamExists,
		},
		{
			name:          "ошибка при проверке существования команды",
			teamName:      "team-1",
			members:       []domain.User{},
			mockExistsErr: errors.New("database error"),
			wantErr:       true,
		},
		{
			name:     "ошибка при создании команды",
			teamName: "team-1",
			members: []domain.User{
				{ID: "user-1", Username: "user1", IsActive: true},
			},
			mockExists:    false,
			mockCreateErr: errors.New("create error"),
			wantErr:       true,
		},
		{
			name:     "ошибка при добавлении членов команды",
			teamName: "team-1",
			members: []domain.User{
				{ID: "user-1", Username: "user1", IsActive: true},
			},
			mockExists:    false,
			mockUpsertErr: errors.New("upsert error"),
			wantErr:       true,
		},
		{
			name:       "создание команды без членов",
			teamName:   "team-1",
			members:    []domain.User{},
			mockExists: false,
			wantErr:    false,
			validateResult: func(t *testing.T, team *domain.Team) {
				if team.Name != "team-1" {
					t.Errorf("expected team name team-1, got %s", team.Name)
				}
				if len(team.Members) != 0 {
					t.Errorf("expected 0 members, got %d", len(team.Members))
				}
			},
		},
		{
			name:     "создание команды с большим количеством членов",
			teamName: "team-1",
			members: []domain.User{
				{ID: "user-1", Username: "user1", IsActive: true},
				{ID: "user-2", Username: "user2", IsActive: true},
				{ID: "user-3", Username: "user3", IsActive: true},
				{ID: "user-4", Username: "user4", IsActive: true},
				{ID: "user-5", Username: "user5", IsActive: true},
			},
			mockExists: false,
			wantErr:    false,
			validateResult: func(t *testing.T, team *domain.Team) {
				if len(team.Members) != 5 {
					t.Errorf("expected 5 members, got %d", len(team.Members))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTeamRepo := &mocks.MockTeamRepository{
				ExistsResult: tt.mockExists,
				ExistsErr:    tt.mockExistsErr,
				CreateErr:    tt.mockCreateErr,
			}
			mockUserRepo := &mocks.MockTeamUserRepository{
				UpsertErr: tt.mockUpsertErr,
			}

			service := NewTeamService(mockTeamRepo, mockUserRepo)
			ctx := context.Background()

			result, err := service.CreateTeam(ctx, tt.teamName, tt.members)

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

func TestTeamService_GetTeam(t *testing.T) {
	tests := []struct {
		name           string
		teamName       string
		mockTeam       *domain.Team
		mockGetErr     error
		wantErr        bool
		validateResult func(t *testing.T, team *domain.Team)
	}{
		{
			name:     "успешное получение команды",
			teamName: "team-1",
			mockTeam: &domain.Team{
				Name: "team-1",
				Members: []domain.User{
					{ID: "user-1", Username: "user1", TeamName: "team-1", IsActive: true},
					{ID: "user-2", Username: "user2", TeamName: "team-1", IsActive: true},
				},
			},
			wantErr: false,
			validateResult: func(t *testing.T, team *domain.Team) {
				if team.Name != "team-1" {
					t.Errorf("expected team name team-1, got %s", team.Name)
				}
				if len(team.Members) != 2 {
					t.Errorf("expected 2 members, got %d", len(team.Members))
				}
			},
		},
		{
			name:       "команда не найдена",
			teamName:   "team-999",
			mockGetErr: errors.New("team not found"),
			wantErr:    true,
		},
		{
			name:     "команда без членов",
			teamName: "team-1",
			mockTeam: &domain.Team{
				Name:    "team-1",
				Members: []domain.User{},
			},
			wantErr: false,
			validateResult: func(t *testing.T, team *domain.Team) {
				if team.Name != "team-1" {
					t.Errorf("expected team name team-1, got %s", team.Name)
				}
				if len(team.Members) != 0 {
					t.Errorf("expected 0 members, got %d", len(team.Members))
				}
			},
		},
		{
			name:       "ошибка базы данных",
			teamName:   "team-1",
			mockGetErr: errors.New("database error"),
			wantErr:    true,
		},
		{
			name:     "команда с неактивными членами",
			teamName: "team-1",
			mockTeam: &domain.Team{
				Name: "team-1",
				Members: []domain.User{
					{ID: "user-1", Username: "user1", TeamName: "team-1", IsActive: true},
					{ID: "user-2", Username: "user2", TeamName: "team-1", IsActive: false},
					{ID: "user-3", Username: "user3", TeamName: "team-1", IsActive: true},
				},
			},
			wantErr: false,
			validateResult: func(t *testing.T, team *domain.Team) {
				if len(team.Members) != 3 {
					t.Errorf("expected 3 members, got %d", len(team.Members))
				}
				activeCount := 0
				for _, member := range team.Members {
					if member.IsActive {
						activeCount++
					}
				}
				if activeCount != 2 {
					t.Errorf("expected 2 active members, got %d", activeCount)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTeamRepo := &mocks.MockTeamRepository{
				GetWithMembersResult: tt.mockTeam,
				GetWithMembersErr:    tt.mockGetErr,
			}
			mockUserRepo := &mocks.MockTeamUserRepository{}

			service := NewTeamService(mockTeamRepo, mockUserRepo)
			ctx := context.Background()

			result, err := service.GetTeam(ctx, tt.teamName)

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
