package http

import (
	"encoding/json"
	"net/http"

	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/api/openapi"
	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/internal/domain"
	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/internal/service/converter"
)

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

type deactivateTeamRequest struct {
	TeamName string `json:"team_name"`
}

type deactivateTeamResponse struct {
	TeamName            string `json:"team_name"`
	DeactivatedUsers    int    `json:"deactivated_users"`
	UpdatedPullRequests int    `json:"updated_pull_requests"`
}

func (s *Server) HandleTeamDeactivate(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req deactivateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeDomainError(w, http.StatusBadRequest, domain.ErrorCodeNotFound, "invalid JSON body")
		return
	}
	if req.TeamName == "" {
		s.writeDomainError(w, http.StatusBadRequest, domain.ErrorCodeNotFound, "team_name is required")
		return
	}

	res, err := s.app.PR.DeactivateTeamAndReassignOpenPRs(r.Context(), req.TeamName)
	if err != nil {
		s.handleError(w, err, http.StatusInternalServerError)
		return
	}

	resp := deactivateTeamResponse{
		TeamName:            res.TeamName,
		DeactivatedUsers:    res.DeactivatedUsers,
		UpdatedPullRequests: res.UpdatedPullRequests,
	}

	s.writeJSON(w, http.StatusOK, resp)
}
