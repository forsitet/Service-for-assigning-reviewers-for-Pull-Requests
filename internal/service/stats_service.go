package service

import (
	"context"
)

type AssignmentStatsRepo interface {
	CountAssignmentsByReviewer(ctx context.Context) (map[string]int64, error)
	CountAssignmentsByPR(ctx context.Context) (map[string]int64, error)
}

type StatsService struct {
	prRepo AssignmentStatsRepo
}

func NewStatsService(prRepo AssignmentStatsRepo) *StatsService {
	return &StatsService{prRepo: prRepo}
}

func (s *StatsService) GetAssignmentStats(ctx context.Context) (map[string]int64, map[string]int64, error) {
	byUser, err := s.prRepo.CountAssignmentsByReviewer(ctx)
	if err != nil {
		return nil, nil, err
	}

	byPR, err := s.prRepo.CountAssignmentsByPR(ctx)
	if err != nil {
		return nil, nil, err
	}

	return byUser, byPR, nil
}
