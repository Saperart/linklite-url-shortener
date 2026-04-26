package httpcontroller

import "net/http"

func (h *Handler) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/links", h.CreateLink)
	mux.HandleFunc("GET /api/v1/links/{code}", h.ResolveLink)
	mux.HandleFunc("GET /{code}", h.Redirect)

	return mux
}
