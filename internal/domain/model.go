package domain

type User struct {
	ID       string
	Username string
	Teamnam  string
	IsActive bool
}

type Team struct {
	Name    string
	Members []User
}

type PRStatus string

const (
	PRStatusOpen   PRStatus = "open"
	PRStatusMerged PRStatus = "merged"
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
