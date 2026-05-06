package httpcontroller

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

//go:generate go run github.com/golang/mock/mockgen@v1.6.0 -source=handler.go -destination=./mocks/mock_link_service.go -package=mocks

type linkService interface {
	CreateLink(ctx context.Context, originalURL string) (string, error)
	ResolveLink(ctx context.Context, shortCode string) (string, error)
}

type Handler struct {
	service linkService
	logger  *zap.Logger
	baseURL string
}

func NewHandler(service linkService, logger *zap.Logger, baseURL string) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
		baseURL: strings.TrimRight(baseURL, "/"),
	}
}
func (h *Handler) CreateLink(w http.ResponseWriter, r *http.Request) {
	defer closeRequestBody(h.logger, r.Body)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var req createLinkRequest
	if err := decoder.Decode(&req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: errorBody{
				Code:    "invalid_json",
				Message: "invalid JSON request body",
			},
		})
		return
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		h.writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: errorBody{
				Code:    "invalid_json",
				Message: "request body must contain a single JSON object",
			},
		})
		return
	}

	shortCode, err := h.service.CreateLink(r.Context(), req.URL)
	if err != nil {
		status, payload := mapError(err)
		h.writeJSON(w, status, payload)
		return
	}

	h.writeJSON(w, http.StatusOK, createLinkResponse{
		ShortURL: h.baseURL + "/" + shortCode,
	})
}
func (h *Handler) ResolveLink(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")

	originalURL, err := h.service.ResolveLink(r.Context(), code)
	if err != nil {
		status, payload := mapError(err)
		h.writeJSON(w, status, payload)
		return
	}

	h.writeJSON(w, http.StatusOK, resolveLinkResponse{
		OriginalURL: originalURL,
	})
}

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/" {
		http.NotFound(w, r)
		return
	}

	code := strings.TrimPrefix(r.URL.Path, "/")

	originalURL, err := h.service.ResolveLink(r.Context(), code)
	if err != nil {
		status, payload := mapError(err)
		h.writeJSON(w, status, payload)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		h.logger.Error("write json response", zap.Error(err))
	}
}

func closeRequestBody(logger *zap.Logger, body io.ReadCloser) {
	if body == nil {
		return
	}
	if err := body.Close(); err != nil {
		logger.Warn("close request body", zap.Error(err))
	}
}
