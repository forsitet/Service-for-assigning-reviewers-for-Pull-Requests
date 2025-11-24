package http

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/api/openapi"
	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/internal/domain"
	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/internal/service/converter"
)

type setUserActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

type userResponse struct {
	User openapi.User `json:"user"`
}

func (s *Server) HandleUserSetIsActive(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := r.Body.Close(); err != nil {
			slog.Debug("error closing body in HandleUserSetIsActive", "error", err)
		}
	}()
	var req setUserActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeDomainError(w, http.StatusBadRequest, domain.ErrorCodeNotFound, "invalid JSON body")
		return
	}
	user, err := s.app.User.SetActive(r.Context(), req.UserID, req.IsActive)
	if err != nil {
		s.handleError(w, err)
		return
	}
	resp := userResponse{
		User: converter.UserToOpenAPI(user),
	}
	s.writeJSON(w, http.StatusOK, resp)
}

type userReviewsResponse struct {
	UserID       string                     `json:"user_id"`
	PullRequests []openapi.PullRequestShort `json:"pull_requests"`
}

func (s *Server) HandleUserGetReview(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		s.writeDomainError(w, http.StatusBadRequest, domain.ErrorCodeNotFound, "user_id is required")
		return
	}

	prs, err := s.app.User.ListAssignedPullRequests(r.Context(), userID)
	if err != nil {
		s.handleError(w, err)
		return
	}

	resp := userReviewsResponse{
		UserID:       userID,
		PullRequests: make([]openapi.PullRequestShort, 0, len(prs)),
	}

	for i := range prs {
		resp.PullRequests = append(resp.PullRequests, converter.PullRequestShortFromDomain(&prs[i]))
	}

	s.writeJSON(w, http.StatusOK, resp)
}
