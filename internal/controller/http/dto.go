package httpcontroller

type createLinkRequest struct {
	URL string `json:"url"`
}

type createLinkResponse struct {
	ShortURL string `json:"short_url"`
}

type resolveLinkResponse struct {
	OriginalURL string `json:"original_url"`
}

type errorResponse struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
