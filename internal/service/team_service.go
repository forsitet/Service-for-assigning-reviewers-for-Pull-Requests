package service

import (
	"context"
	"fmt"

	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/internal/domain"
)

type TeamRepository interface {
	Create(ctx context.Context, name string) error
	Exists(ctx context.Context, name string) (bool, error)
	GetWithMembers(ctx context.Context, name string) (*domain.Team, error)
}

type TeamUserRepository interface {
	UpsertForTeam(ctx context.Context, teamName string, users []domain.User) error
}

type TeamService struct {
	teams TeamRepository
	users TeamUserRepository
}

func NewTeamService(teams TeamRepository, users TeamUserRepository) *TeamService {
	return &TeamService{
		teams: teams,
		users: users,
	}
}

func (s *TeamService) CreateTeam(ctx context.Context, teamName string, members []domain.User) (*domain.Team, error) {
	exists, err := s.teams.Exists(ctx, teamName)
	if err != nil {
		return nil, fmt.Errorf("check team exists: %w", err)
	}
	if exists {
		return nil, domain.NewDomainError(domain.ErrorCodeTeamExists, "team already exists")
	}

	if err := s.teams.Create(ctx, teamName); err != nil {
		return nil, fmt.Errorf("create team: %w", err)
	}

	for i := range members {
		members[i].TeamName = teamName
	}

	if err := s.users.UpsertForTeam(ctx, teamName, members); err != nil {
		return nil, fmt.Errorf("upsert team members: %w", err)
	}

	return &domain.Team{
		Name:    teamName,
		Members: members,
	}, nil
}

func (s *TeamService) GetTeam(ctx context.Context, teamName string) (*domain.Team, error) {
	team, err := s.teams.GetWithMembers(ctx, teamName)
	if err != nil {
		return nil, fmt.Errorf("get team: %w", err)
	}
	return team, nil
}
