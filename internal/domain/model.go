package domain

type User struct {
	ID       string
	Username string
	TeamName string
	IsActive bool
}

type Team struct {
	Name    string
	Members []User
}

type PRStatus string

const (
	PRStatusOpen   PRStatus = "OPEN"
	PRStatusMerged PRStatus = "MERGED"
)

type PullRequest struct {
	ID                string
	Name              string
	AuthorID          string
	Status            PRStatus
	AssignedReviewers []string
	CreatedAt         int64
	MergedAt          int64
}

func (p PullRequest) IsMerged() bool {
	return p.Status == PRStatusMerged
}
