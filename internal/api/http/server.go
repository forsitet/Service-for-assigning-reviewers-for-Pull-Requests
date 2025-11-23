package http

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/api/openapi"
	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/internal/domain"
	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/internal/service"
	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/internal/service/converter"
)

type Server struct {
	app    *service.App
	logger *slog.Logger
}

func NewServer(app *service.App, logger *slog.Logger) *Server {
	return &Server{
		app:    app,
		logger: logger,
	}
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type errorResponse struct {
	Error apiError `json:"error"`
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if v == nil {
		return
	}

	if err := json.NewEncoder(w).Encode(v); err != nil {
		s.logger.Error("failed to encode JSON response", "error", err)
	}
}

func (s *Server) writeDomainError(w http.ResponseWriter, status int, code domain.ErrorCode, message string) {
	resp := errorResponse{
		Error: apiError{
			Code:    string(code),
			Message: message,
		},
	}
	s.writeJSON(w, status, resp)
}

func (s *Server) writeUnknownError(w http.ResponseWriter, message string) {
	resp := errorResponse{
		Error: apiError{
			Code:    string(domain.ErrorCodeNotFound),
			Message: message,
		},
	}
	s.writeJSON(w, http.StatusInternalServerError, resp)
}

func (s *Server) handleError(w http.ResponseWriter, err error, defaultStatus int) {
	if err == nil {
		return
	}

	var de *domain.DomainError
	if errors.As(err, &de) {
		status := defaultStatus

		switch de.Code {
		case domain.ErrorCodeTeamExists:
			status = http.StatusBadRequest
		case domain.ErrorCodePRExists:
			status = http.StatusConflict
		case domain.ErrorCodePRMerged,
			domain.ErrorCodeNotAssigned,
			domain.ErrorCodeNoCandidate:
			status = http.StatusConflict
		case domain.ErrorCodeNotFound:
			status = http.StatusNotFound
		}

		s.writeDomainError(w, status, de.Code, de.Message)
		return
	}

	if errors.Is(err, sql.ErrNoRows) {
		s.writeDomainError(w, http.StatusNotFound, domain.ErrorCodeNotFound, "resource not found")
		return
	}

	s.logger.Error("unexpected error", "error", err)
	s.writeUnknownError(w, "internal server error")
}

type createTeamResponse struct {
	Team openapi.Team `json:"team"`
}

func (s *Server) HandleTeamAdd(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req openapi.Team
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeDomainError(w, http.StatusBadRequest, domain.ErrorCodeNotFound, "invalid JSON body")
		return
	}

	domainTeam := converter.TeamFromOpenAPI(&req)

	created, err := s.app.Team.CreateTeam(r.Context(), domainTeam.Name, domainTeam.Members)
	if err != nil {
		s.handleError(w, err, http.StatusInternalServerError)
		return
	}

	resp := createTeamResponse{
		Team: converter.TeamToOpenAPI(created),
	}
	s.writeJSON(w, http.StatusCreated, resp)
}

func (s *Server) HandleTeamGet(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		s.writeDomainError(w, http.StatusBadRequest, domain.ErrorCodeNotFound, "team_name is required")
		return
	}

	team, err := s.app.Team.GetTeam(r.Context(), teamName)
	if err != nil {
		s.handleError(w, err, http.StatusInternalServerError)
		return
	}

	resp := converter.TeamToOpenAPI(team)
	s.writeJSON(w, http.StatusOK, resp)
}

type setUserActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

type userResponse struct {
	User openapi.User `json:"user"`
}

func (s *Server) HandleUserSetIsActive(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req setUserActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeDomainError(w, http.StatusBadRequest, domain.ErrorCodeNotFound, "invalid JSON body")
		return
	}
	user, err := s.app.User.SetActive(r.Context(), req.UserID, req.IsActive)
	if err != nil {
		s.handleError(w, err, http.StatusInternalServerError)
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
		s.handleError(w, err, http.StatusInternalServerError)
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

func (s *Server) HealthCheck(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

const openAPISpecPath = "api/openapi/openapi.yml"

func (s *Server) ServeOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, openAPISpecPath)
}

func (s *Server) SwaggerUI(w http.ResponseWriter, r *http.Request) {
	const html = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
    <title>Swagger UI - Reviewer Service</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist/swagger-ui.css" />
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist/swagger-ui-standalone-preset.js"></script>
    <script>
      window.onload = function() {
        window.ui = SwaggerUIBundle({
          url: '/openapi.yaml',
          dom_id: '#swagger-ui',
          presets: [
            SwaggerUIBundle.presets.apis,
            SwaggerUIStandalonePreset
          ],
          layout: 'StandaloneLayout'
        });
      };
    </script>
  </body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := w.Write([]byte(html)); err != nil {
		s.logger.Error("failed to write swagger ui html", "error", err)
	}
}
