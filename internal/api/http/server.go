package http

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/internal/domain"
	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/internal/service"
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

func (s *Server) handleError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	var de *domain.DomainError
	if errors.As(err, &de) {
		status := http.StatusInternalServerError

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
	s.writeDomainError(w, http.StatusInternalServerError, domain.ErrorCodeNotFound, "internal server error")
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
