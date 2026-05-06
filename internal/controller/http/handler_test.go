package httpcontroller

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	mockhttp "github.com/Saperart/linklite-url-shortener/internal/controller/http/mocks"
	xerrors "github.com/Saperart/linklite-url-shortener/internal/errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const (
	testBaseURL     = "http://localhost:8080"
	testOriginalURL = "https://example.com"
	testShortCode   = "Abc123_DEF"
)

func newTestHandler(service linkService) *Handler {
	return NewHandler(service, zap.NewNop(), testBaseURL)
}

func TestHandlerCreateLink(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		body          string
		prepare       func(mockService *mockhttp.MocklinkService)
		wantStatus    int
		wantShortURL  string
		wantErrorCode string
	}{
		{
			name: "success",
			body: `{"url":"https://example.com"}`,
			prepare: func(mockService *mockhttp.MocklinkService) {
				mockService.EXPECT().
					CreateLink(gomock.Any(), testOriginalURL).
					Return(testShortCode, nil)
			},
			wantStatus:   http.StatusOK,
			wantShortURL: testBaseURL + "/" + testShortCode,
		},
		{
			name:          "invalid json",
			body:          `{"url":`,
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "invalid_json",
		},
		{
			name:          "unknown field",
			body:          `{"url":"https://example.com","extra":"value"}`,
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "invalid_json",
		},
		{
			name:          "several json objects",
			body:          `{"url":"https://example.com"} {"url":"https://other.com"}`,
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "invalid_json",
		},
		{
			name: "service invalid url error",
			body: `{"url":"bad-url"}`,
			prepare: func(mockService *mockhttp.MocklinkService) {
				mockService.EXPECT().
					CreateLink(gomock.Any(), "bad-url").
					Return("", xerrors.ErrInvalidURL)
			},
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "invalid_url",
		},
		{
			name: "service storage error",
			body: `{"url":"https://example.com"}`,
			prepare: func(mockService *mockhttp.MocklinkService) {
				mockService.EXPECT().
					CreateLink(gomock.Any(), testOriginalURL).
					Return("", xerrors.ErrStorageUnavailable)
			},
			wantStatus:    http.StatusServiceUnavailable,
			wantErrorCode: "storage_unavailable",
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockService := mockhttp.NewMocklinkService(ctrl)
			if tc.prepare != nil {
				tc.prepare(mockService)
			}

			handler := newTestHandler(mockService)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/links", bytes.NewBufferString(tc.body))
			rec := httptest.NewRecorder()

			handler.Router().ServeHTTP(rec, req)

			require.Equal(t, tc.wantStatus, rec.Code, "body: %s", rec.Body.String())

			if tc.wantErrorCode != "" {
				var response errorResponse
				decodeResponse(t, rec, &response)
				assert.Equal(t, tc.wantErrorCode, response.Error.Code)
				return
			}

			var response createLinkResponse
			decodeResponse(t, rec, &response)
			assert.Equal(t, tc.wantShortURL, response.ShortURL)
		})
	}
}

func TestHandlerResolveLink(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		code          string
		prepare       func(mockService *mockhttp.MocklinkService)
		wantStatus    int
		wantURL       string
		wantErrorCode string
	}{
		{
			name: "success",
			code: testShortCode,
			prepare: func(mockService *mockhttp.MocklinkService) {
				mockService.EXPECT().
					ResolveLink(gomock.Any(), testShortCode).
					Return(testOriginalURL, nil)
			},
			wantStatus: http.StatusOK,
			wantURL:    testOriginalURL,
		},
		{
			name: "not found",
			code: testShortCode,
			prepare: func(mockService *mockhttp.MocklinkService) {
				mockService.EXPECT().
					ResolveLink(gomock.Any(), testShortCode).
					Return("", xerrors.ErrNotFound)
			},
			wantStatus:    http.StatusNotFound,
			wantErrorCode: "not_found",
		},
		{
			name: "invalid short code",
			code: "bad",
			prepare: func(mockService *mockhttp.MocklinkService) {
				mockService.EXPECT().
					ResolveLink(gomock.Any(), "bad").
					Return("", xerrors.ErrInvalidShortCode)
			},
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "invalid_short_code",
		},
		{
			name: "storage unavailable",
			code: testShortCode,
			prepare: func(mockService *mockhttp.MocklinkService) {
				mockService.EXPECT().
					ResolveLink(gomock.Any(), testShortCode).
					Return("", xerrors.ErrStorageUnavailable)
			},
			wantStatus:    http.StatusServiceUnavailable,
			wantErrorCode: "storage_unavailable",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockService := mockhttp.NewMocklinkService(ctrl)
			if tc.prepare != nil {
				tc.prepare(mockService)
			}

			handler := newTestHandler(mockService)
			req := httptest.NewRequest(http.MethodGet, "/api/v1/links/"+tc.code, nil)
			rec := httptest.NewRecorder()

			handler.Router().ServeHTTP(rec, req)
			require.Equal(t, tc.wantStatus, rec.Code, "body: %s", rec.Body.String())

			if tc.wantErrorCode != "" {
				var response errorResponse
				decodeResponse(t, rec, &response)
				assert.Equal(t, tc.wantErrorCode, response.Error.Code)
				return
			}

			var response resolveLinkResponse
			decodeResponse(t, rec, &response)
			assert.Equal(t, tc.wantURL, response.OriginalURL)
		})
	}
}

func TestHandlerRedirect(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		path          string
		prepare       func(mockService *mockhttp.MocklinkService)
		wantStatus    int
		wantLocation  string
		wantErrorCode string
	}{
		{
			name: "success",
			path: "/" + testShortCode,
			prepare: func(mockService *mockhttp.MocklinkService) {
				mockService.EXPECT().
					ResolveLink(gomock.Any(), testShortCode).
					Return(testOriginalURL, nil)
			},
			wantStatus:   http.StatusFound,
			wantLocation: testOriginalURL,
		},
		{
			name:       "root is not found",
			path:       "/",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "api path is not redirected",
			path:       "/api/something",
			wantStatus: http.StatusNotFound,
		},
		{
			name: "link not found",
			path: "/" + testShortCode,
			prepare: func(mockService *mockhttp.MocklinkService) {
				mockService.EXPECT().
					ResolveLink(gomock.Any(), testShortCode).
					Return("", xerrors.ErrNotFound)
			},
			wantStatus:    http.StatusNotFound,
			wantErrorCode: "not_found",
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockService := mockhttp.NewMocklinkService(ctrl)
			if tc.prepare != nil {
				tc.prepare(mockService)
			}

			handler := newTestHandler(mockService)
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()

			handler.Router().ServeHTTP(rec, req)
			require.Equal(t, tc.wantStatus, rec.Code, "body: %s", rec.Body.String())

			if tc.wantLocation != "" {
				assert.Equal(t, tc.wantLocation, rec.Header().Get("Location"))
			}

			if tc.wantErrorCode != "" {
				var response errorResponse
				decodeResponse(t, rec, &response)
				assert.Equal(t, tc.wantErrorCode, response.Error.Code)
			}
		})
	}
}

func TestHandlerWriteJSON(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	mockService := mockhttp.NewMocklinkService(ctrl)

	handler := newTestHandler(mockService)
	rec := httptest.NewRecorder()

	handler.writeJSON(rec, http.StatusCreated, createLinkResponse{
		ShortURL: testBaseURL + "/" + testShortCode,
	})

	require.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var response createLinkResponse
	decodeResponse(t, rec, &response)
	assert.Equal(t, testBaseURL+"/"+testShortCode, response.ShortURL)
}

func TestCloseRequestBody(t *testing.T) {
	t.Parallel()

	t.Run("nil body", func(t *testing.T) {
		t.Parallel()
		closeRequestBody(zap.NewNop(), nil)
	})

	t.Run("body close success", func(t *testing.T) {
		t.Parallel()

		body := ioReadCloser{
			closeFunc: func() error {
				return nil
			},
		}

		closeRequestBody(zap.NewNop(), body)
	})

	t.Run("body close error", func(t *testing.T) {
		t.Parallel()

		body := ioReadCloser{
			closeFunc: func() error {
				return errors.New("close failed")
			},
		}

		closeRequestBody(zap.NewNop(), body)
	})
}

type ioReadCloser struct {
	closeFunc func() error
}

func (c ioReadCloser) Read(_ []byte) (int, error) {
	return 0, io.EOF
}

func (c ioReadCloser) Close() error {
	return c.closeFunc()
}

func decodeResponse(t *testing.T, rec *httptest.ResponseRecorder, target any) {
	t.Helper()

	err := json.NewDecoder(rec.Body).Decode(target)
	require.NoError(t, err, "body: %s", rec.Body.String())
}
