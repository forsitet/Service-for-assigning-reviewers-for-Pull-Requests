package http

import (
	"net/http"
	"sort"
)

type userAssignmentStatsDTO struct {
	UserID      string `json:"user_id"`
	Assignments int64  `json:"assignments"`
}

type prAssignmentStatsDTO struct {
	PullRequestID string `json:"pull_request_id"`
	Assignments   int64  `json:"assignments"`
}

type assignmentStatsResponse struct {
	ByUser        []userAssignmentStatsDTO `json:"by_user"`
	ByPullRequest []prAssignmentStatsDTO   `json:"by_pull_request"`
}

func (s *Server) HandleStatsAssignments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	byUser, byPR, err := s.app.Stats.GetAssignmentStats(ctx)
	if err != nil {
		s.handleError(w, err, http.StatusInternalServerError)
		return
	}

	userIDs := make([]string, 0, len(byUser))
	for id := range byUser {
		userIDs = append(userIDs, id)
	}
	sort.Strings(userIDs)

	prIDs := make([]string, 0, len(byPR))
	for id := range byPR {
		prIDs = append(prIDs, id)
	}
	sort.Strings(prIDs)

	resp := assignmentStatsResponse{
		ByUser:        make([]userAssignmentStatsDTO, 0, len(userIDs)),
		ByPullRequest: make([]prAssignmentStatsDTO, 0, len(prIDs)),
	}

	for _, id := range userIDs {
		resp.ByUser = append(resp.ByUser, userAssignmentStatsDTO{
			UserID:      id,
			Assignments: byUser[id],
		})
	}

	for _, id := range prIDs {
		resp.ByPullRequest = append(resp.ByPullRequest, prAssignmentStatsDTO{
			PullRequestID: id,
			Assignments:   byPR[id],
		})
	}

	s.writeJSON(w, http.StatusOK, resp)
}
