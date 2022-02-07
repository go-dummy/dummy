package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/neotoolkit/dummy/internal/api"
	"github.com/neotoolkit/dummy/internal/logger"
	"github.com/neotoolkit/dummy/internal/model"
)

type NoOpResponse struct {
}

func (n NoOpResponse) StatusCode() int {
	return http.StatusInternalServerError
}

func (n NoOpResponse) ExampleValue(_ string) interface{} {
	return struct{}{}
}

// Handlers -.
type Handlers struct {
	API    model.API
	Logger *logger.Logger
}

// NewHandlers returns a new instance of Handlers
func NewHandlers(api model.API, l *logger.Logger) Handlers {
	return Handlers{
		API:    api,
		Logger: l,
	}
}

// Handler -.
func (s *Server) Handler(w http.ResponseWriter, r *http.Request) {
	if setStatusCode(w, r.Header.Get("X-Set-Status-Code")) {
		return
	}

	w.Header().Set("Content-Type", "application/json")

	path := RemoveFragment(r.URL.Path)

	response, ok, err := s.Handlers.Get(path, r.Method, r.Body)
	if ok {
		if _, ok := err.(*json.SyntaxError); ok || errors.Is(err, api.ErrEmptyRequireField) {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		w.WriteHeader(response.StatusCode())
		resp := response.ExampleValue(r.Header.Get("X-Example"))

		if nil == resp {
			return
		}

		bytes, err := json.Marshal(resp)
		if err != nil {
			s.Logger.Error().Err(err).Msg("serialize response")
		}

		if _, err := w.Write(bytes); err != nil {
			s.Logger.Error().Err(err).Msg("write response")
		}

		return
	}

	w.WriteHeader(http.StatusNotFound)
}

// Get -.
func (h Handlers) Get(path, method string, body io.ReadCloser) (model.Response, bool, error) {
	response, err := h.API.FindResponse(model.FindResponseParams{
		Path:   path,
		Method: method,
		Body:   body,
	})
	if err != nil {
		if errors.Is(err, api.ErrEmptyRequireField) {
			return NoOpResponse{}, true, err
		}

		if _, ok := err.(*json.SyntaxError); ok {
			return NoOpResponse{}, true, err
		}

		return NoOpResponse{}, false, err
	}

	return response, true, nil
}

func setStatusCode(w http.ResponseWriter, statusCode string) bool {
	switch statusCode {
	case "500":
		w.WriteHeader(http.StatusInternalServerError)

		return true
	default:
		return false
	}
}

// RemoveTrailingSlash returns path without trailing slash
func RemoveTrailingSlash(path string) string {
	if len(path) > 0 && path[len(path)-1] == '/' {
		return path[0 : len(path)-1]
	}

	return path
}

// RemoveFragment - clear url from reference in path
func RemoveFragment(path string) string {
	return RemoveTrailingSlash(strings.Split(path, "#")[0])
}
