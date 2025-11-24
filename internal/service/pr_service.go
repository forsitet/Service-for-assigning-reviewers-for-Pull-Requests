package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/internal/domain"
)

type PRRepository interface {
	CreateWithReviewers(ctx context.Context, pr *domain.PullRequest, reviewerIDs []string) error
	GetByID(ctx context.Context, id string) (*domain.PullRequest, []string, error)
	SetMerged(ctx context.Context, id string, mergedAt time.Time) (*domain.PullRequest, []string, error)
	UpdateReviewers(ctx context.Context, id string, reviewerIDs []string) (*domain.PullRequest, []string, error)
	ListByReviewer(ctx context.Context, userID string) ([]domain.PullRequest, error)
	Exists(ctx context.Context, id string) (bool, error)
	DeactivateTeamAndReassignOpenPRs(ctx context.Context, teamName string) (domain.TeamDeactivationResult, error)
}

type PRUserRepository interface {
	GetByID(ctx context.Context, id string) (*domain.User, error)
	ListByTeam(ctx context.Context, teamName string) ([]domain.User, error)
}

type PRService struct {
	prs     PRRepository
	users   PRUserRepository
	nowFunc func() time.Time
}

func NewPRService(prs PRRepository, users PRUserRepository, nowFunc func() time.Time) *PRService {
	if nowFunc == nil {
		nowFunc = time.Now
	}
	return &PRService{
		prs:     prs,
		users:   users,
		nowFunc: nowFunc,
	}
}

func (s *PRService) CreatePullRequest(
	ctx context.Context,
	id string,
	name string,
	authorID string,
) (*domain.PullRequest, error) {
	exists, err := s.prs.Exists(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("check PR exists: %w", err)
	}
	if exists {
		return nil, domain.NewDomainError(domain.ErrorCodePRExists, "pull request already exists")
	}

	author, err := s.users.GetByID(ctx, authorID)
	if err != nil {
		return nil, fmt.Errorf("get author: %w", err)
	}

	teamMembers, err := s.users.ListByTeam(ctx, author.TeamName)
	if err != nil {
		return nil, fmt.Errorf("list team members: %w", err)
	}

	reviewerIDs := selectInitialReviewers(author.ID, teamMembers)

	now := s.nowFunc()
	pr := &domain.PullRequest{
		ID:                id,
		Name:              name,
		AuthorID:          author.ID,
		Status:            domain.PRStatusOpen,
		AssignedReviewers: reviewerIDs,
		CreatedAt:         now.Unix(),
		MergedAt:          0,
	}

	if err := s.prs.CreateWithReviewers(ctx, pr, reviewerIDs); err != nil {
		return nil, fmt.Errorf("create PR with reviewers: %w", err)
	}

	return pr, nil
}

func (s *PRService) MergePullRequest(ctx context.Context, id string) (*domain.PullRequest, error) {
	pr, reviewers, err := s.prs.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get PR for merge: %w", err)
	}

	if pr.IsMerged() {
		pr.AssignedReviewers = reviewers
		return pr, nil
	}

	mergedAt := s.nowFunc()
	pr, reviewers, err = s.prs.SetMerged(ctx, id, mergedAt)
	if err != nil {
		return nil, fmt.Errorf("set PR merged: %w", err)
	}
	pr.AssignedReviewers = reviewers

	return pr, nil
}

func (s *PRService) ReassignReviewer(
	ctx context.Context,
	prID string,
	oldReviewerID string,
) (*domain.PullRequest, error) {
	pr, reviewers, err := s.prs.GetByID(ctx, prID)
	if err != nil {
		return nil, fmt.Errorf("get PR for reassign: %w", err)
	}

	if pr.IsMerged() {
		return nil, domain.NewDomainError(domain.ErrorCodePRMerged, "cannot reassign reviewers for merged PR")
	}

	reviewerIndex := -1
	for i, rID := range reviewers {
		if rID == oldReviewerID {
			reviewerIndex = i
			break
		}
	}
	if reviewerIndex == -1 {
		return nil, domain.NewDomainError(domain.ErrorCodeNotAssigned, "reviewer is not assigned to this PR")
	}

	oldReviewer, err := s.users.GetByID(ctx, oldReviewerID)
	if err != nil {
		return nil, fmt.Errorf("get old reviewer: %w", err)
	}

	teamMembers, err := s.users.ListByTeam(ctx, oldReviewer.TeamName)
	if err != nil {
		return nil, fmt.Errorf("list team members for reassign: %w", err)
	}

	newReviewerID, err := selectReplacementReviewer(pr.AuthorID, oldReviewerID, reviewers, teamMembers)
	if err != nil {
		return nil, err
	}

	newReviewers := make([]string, len(reviewers))
	copy(newReviewers, reviewers)
	newReviewers[reviewerIndex] = newReviewerID

	updated, updatedReviewers, err := s.prs.UpdateReviewers(ctx, prID, newReviewers)
	if err != nil {
		return nil, fmt.Errorf("update reviewers: %w", err)
	}
	updated.AssignedReviewers = updatedReviewers

	return updated, nil
}

func (s *PRService) DeactivateTeamAndReassignOpenPRs(ctx context.Context, teamName string) (domain.TeamDeactivationResult, error) {
	return s.prs.DeactivateTeamAndReassignOpenPRs(ctx, teamName)
}

func selectInitialReviewers(authorID string, members []domain.User) []string {
	candidates := make([]string, 0, len(members))
	for _, m := range members {
		if !m.IsActive {
			continue
		}
		if m.ID == authorID {
			continue
		}
		candidates = append(candidates, m.ID)
	}

	return pickRandomSubset(candidates, 2)
}

func selectReplacementReviewer(
	authorID string,
	oldReviewerID string,
	currentReviewerIDs []string,
	teamMembers []domain.User,
) (string, error) {
	currentSet := make(map[string]struct{}, len(currentReviewerIDs))
	for _, id := range currentReviewerIDs {
		currentSet[id] = struct{}{}
	}

	candidates := make([]string, 0, len(teamMembers))
	for _, m := range teamMembers {
		if !m.IsActive {
			continue
		}
		if m.ID == authorID {
			continue
		}
		if m.ID == oldReviewerID {
			continue
		}
		if _, exists := currentSet[m.ID]; exists {
			continue
		}
		candidates = append(candidates, m.ID)
	}

	if len(candidates) == 0 {
		return "", domain.NewDomainError(domain.ErrorCodeNoCandidate, "no available candidate for reassignment")
	}

	selected := pickRandomSubset(candidates, 1)
	return selected[0], nil
}

func pickRandomSubset(ids []string, max int) []string {
	if max <= 0 || len(ids) == 0 {
		return nil
	}
	if len(ids) <= max {
		out := make([]string, len(ids))
		copy(out, ids)
		return out
	}

	// #nosec G404 -- non-cryptographic random is acceptable for reviewer selection
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	pool := make([]string, len(ids))
	copy(pool, ids)

	out := make([]string, 0, max)
	for len(out) < max && len(pool) > 0 {
		idx := r.Intn(len(pool))
		out = append(out, pool[idx])
		pool[idx] = pool[len(pool)-1]
		pool = pool[:len(pool)-1]
	}
	return out
}
