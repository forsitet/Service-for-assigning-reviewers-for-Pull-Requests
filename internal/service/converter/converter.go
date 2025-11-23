package converter

import (
	"time"

	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/api/openapi"
	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/internal/domain"
)

func TeamFromOpenAPI(t *openapi.Team) domain.Team {
	if t == nil {
		return domain.Team{}
	}

	members := make([]domain.User, 0, len(t.Members))
	for _, m := range t.Members {
		members = append(members, UserFromTeamMember(&m, t.TeamName))
	}

	return domain.Team{
		Name:    t.TeamName,
		Members: members,
	}
}

func TeamToOpenAPI(t *domain.Team) openapi.Team {
	if t == nil {
		return openapi.Team{}
	}

	members := make([]openapi.TeamMember, 0, len(t.Members))
	for _, u := range t.Members {
		members = append(members, TeamMemberFromDomain(&u))
	}

	return openapi.Team{
		TeamName: t.Name,
		Members:  members,
	}
}

func TeamMemberFromDomain(u *domain.User) openapi.TeamMember {
	if u == nil {
		return openapi.TeamMember{}
	}

	return openapi.TeamMember{
		UserId:   u.ID,
		Username: u.Username,
		IsActive: u.IsActive,
	}
}

func UserFromTeamMember(m *openapi.TeamMember, teamName string) domain.User {
	if m == nil {
		return domain.User{}
	}

	return domain.User{
		ID:       m.UserId,
		Username: m.Username,
		TeamName: teamName,
		IsActive: m.IsActive,
	}
}

func UserFromOpenAPI(u *openapi.User) domain.User {
	if u == nil {
		return domain.User{}
	}

	return domain.User{
		ID:       u.UserId,
		Username: u.Username,
		TeamName: u.TeamName,
		IsActive: u.IsActive,
	}
}

func UserToOpenAPI(u *domain.User) openapi.User {
	if u == nil {
		return openapi.User{}
	}

	return openapi.User{
		UserId:   u.ID,
		Username: u.Username,
		TeamName: u.TeamName,
		IsActive: u.IsActive,
	}
}

func PullRequestFromOpenAPI(p *openapi.PullRequest) domain.PullRequest {
	if p == nil {
		return domain.PullRequest{}
	}

	return domain.PullRequest{
		ID:                p.PullRequestId,
		Name:              p.PullRequestName,
		AuthorID:          p.AuthorId,
		Status:            domain.PRStatus(p.Status),
		AssignedReviewers: append([]string(nil), p.AssignedReviewers...),
		CreatedAt:         timePtrToUnix(p.CreatedAt),
		MergedAt:          timePtrToUnix(p.MergedAt),
	}
}

func PullRequestToOpenAPI(p *domain.PullRequest) openapi.PullRequest {
	if p == nil {
		return openapi.PullRequest{}
	}

	assigned := append([]string(nil), p.AssignedReviewers...)

	return openapi.PullRequest{
		PullRequestId:     p.ID,
		PullRequestName:   p.Name,
		AuthorId:          p.AuthorID,
		Status:            openapi.PullRequestStatus(p.Status),
		AssignedReviewers: assigned,
		CreatedAt:         unixToTimePtr(p.CreatedAt),
		MergedAt:          unixToTimePtr(p.MergedAt),
	}
}

func PullRequestShortFromDomain(p *domain.PullRequest) openapi.PullRequestShort {
	if p == nil {
		return openapi.PullRequestShort{}
	}

	return openapi.PullRequestShort{
		PullRequestId:   p.ID,
		PullRequestName: p.Name,
		AuthorId:        p.AuthorID,
		Status:          openapi.PullRequestShortStatus(p.Status),
	}
}

func unixToTimePtr(v int64) *time.Time {
	if v == 0 {
		return nil
	}
	t := time.Unix(v, 0).UTC()
	return &t
}

func timePtrToUnix(t *time.Time) int64 {
	if t == nil {
		return 0
	}
	return t.Unix()
}
