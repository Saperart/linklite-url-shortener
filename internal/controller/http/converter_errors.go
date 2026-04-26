package httpcontroller

import (
	"errors"
	"net/http"

	xerrors "github.com/Saperart/linklite-url-shortener/internal/errors"
)

func mapError(err error) (int, errorResponse) {
	switch {
	case errors.Is(err, xerrors.ErrInvalidURL):
		return http.StatusBadRequest, errorResponse{Error: errorBody{
			Code:    "invalid_url",
			Message: "url must be absolute and use http or https scheme",
		}}
	case errors.Is(err, xerrors.ErrInvalidShortCode):
		return http.StatusBadRequest, errorResponse{Error: errorBody{
			Code:    "invalid_short_code",
			Message: "short code must be 10 chars and use [a-zA-Z0-9_]",
		}}
	case errors.Is(err, xerrors.ErrNotFound):
		return http.StatusNotFound, errorResponse{Error: errorBody{
			Code:    "not_found",
			Message: "link was not found",
		}}
	case errors.Is(err, xerrors.ErrMaxRetriesExceeded):
		return http.StatusConflict, errorResponse{Error: errorBody{
			Code:    "max_retries_exceeded",
			Message: "cannot generate unique short code",
		}}
	case errors.Is(err, xerrors.ErrStorageUnavailable):
		return http.StatusServiceUnavailable, errorResponse{Error: errorBody{
			Code:    "storage_unavailable",
			Message: "storage is unavailable",
		}}
	default:
		return http.StatusInternalServerError, errorResponse{Error: errorBody{
			Code:    "internal_error",
			Message: "internal server error",
		}}
	}
}
