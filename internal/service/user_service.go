package service

import (
	"context"
	"fmt"

	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/internal/domain"
)

type UserRepository interface {
	GetByID(ctx context.Context, id string) (*domain.User, error)
	SetIsActive(ctx context.Context, id string, active bool) (*domain.User, error)
}

type UserPRRepository interface {
	ListByReviewer(ctx context.Context, userID string) ([]domain.PullRequest, error)
}

type UserService struct {
	users UserRepository
	prs   UserPRRepository
}

func NewUserService(users UserRepository, prs UserPRRepository) *UserService {
	return &UserService{
		users: users,
		prs:   prs,
	}
}

func (s *UserService) SetActive(ctx context.Context, userID string, active bool) (*domain.User, error) {
	user, err := s.users.SetIsActive(ctx, userID, active)
	if err != nil {
		return nil, fmt.Errorf("set user active: %w", err)
	}
	return user, nil
}

func (s *UserService) ListAssignedPullRequests(ctx context.Context, userID string) ([]domain.PullRequest, error) {
	prs, err := s.prs.ListByReviewer(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list pull_requests by reviewer: %w", err)
	}
	return prs, nil
}
