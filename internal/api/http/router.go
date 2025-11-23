package http

import (
	"net/http"

	"log/slog"

	"github.com/go-chi/chi/v5"
)

func NewRouter(server *Server, logger *slog.Logger) http.Handler {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger", http.StatusTemporaryRedirect)
	})

	r.Get("/healthz", server.HealthCheck)

	r.Post("/team/add", server.HandleTeamAdd)
	r.Get("/team/get", server.HandleTeamGet)

	r.Post("/users/setIsActive", server.HandleUserSetIsActive)
	r.Get("/users/getReview", server.HandleUserGetReview)

	r.Post("/pullRequest/create", server.HandlePullRequestCreate)
	r.Post("/pullRequest/merge", server.HandlePullRequestMerge)
	r.Post("/pullRequest/reassign", server.HandlePullRequestReassign)

	r.Get("/openapi.yaml", server.ServeOpenAPISpec)
	r.Get("/swagger", server.SwaggerUI)

	r.Get("/stats/assignments", server.HandleStatsAssignments)

	return r
}
