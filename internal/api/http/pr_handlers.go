package http

import (
	"encoding/json"
	"net/http"

	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/api/openapi"
	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/internal/domain"
	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/internal/service/converter"
)

type createPRRequest struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
}

type prResponse struct {
	PR openapi.PullRequest `json:"pr"`
}

func (s *Server) HandlePullRequestCreate(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req createPRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeDomainError(w, http.StatusBadRequest, domain.ErrorCodeNotFound, "invalid JSON body")
		return
	}
	pr, err := s.app.PR.CreatePullRequest(r.Context(), req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		s.handleError(w, err, http.StatusInternalServerError)
		return
	}
	resp := prResponse{
		PR: converter.PullRequestToOpenAPI(pr),
	}
	s.writeJSON(w, http.StatusCreated, resp)
}

type mergePRRequest struct {
	PullRequestID string `json:"pull_request_id"`
}

func (s *Server) HandlePullRequestMerge(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req mergePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeDomainError(w, http.StatusBadRequest, domain.ErrorCodeNotFound, "invalid JSON body")
		return
	}

	pr, err := s.app.PR.MergePullRequest(r.Context(), req.PullRequestID)
	if err != nil {
		s.handleError(w, err, http.StatusInternalServerError)
		return
	}
	resp := prResponse{
		PR: converter.PullRequestToOpenAPI(pr),
	}
	s.writeJSON(w, http.StatusOK, resp)
}

type reassignPRRequest struct {
	PullRequestID string `json:"pull_request_id"`
	OldUserID     string `json:"old_reviewer_id"`
}

type reassignPRResponse struct {
	PR         openapi.PullRequest `json:"pr"`
	ReplacedBy string              `json:"replaced_by"`
}

func (s *Server) HandlePullRequestReassign(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req reassignPRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeDomainError(w, http.StatusBadRequest, domain.ErrorCodeNotFound, "invalid JSON body")
		return
	}
	updated, err := s.app.PR.ReassignReviewer(r.Context(), req.PullRequestID, req.OldUserID)
	if err != nil {
		s.handleError(w, err, http.StatusInternalServerError)
		return
	}
	replacedBy := ""
	for _, id := range updated.AssignedReviewers {
		if id != req.OldUserID {
			replacedBy = id
			break
		}
	}
	resp := reassignPRResponse{
		PR:         converter.PullRequestToOpenAPI(updated),
		ReplacedBy: replacedBy,
	}
	s.writeJSON(w, http.StatusOK, resp)
}
