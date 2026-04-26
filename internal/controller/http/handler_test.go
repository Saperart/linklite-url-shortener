package httpcontroller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	xerrors "github.com/Saperart/linklite-url-shortener/internal/errors"
	"go.uber.org/zap"
)

const (
	testBaseURL     = "http://localhost:8080"
	testOriginalURL = "https://example.com"
	testShortCode   = "Abc123_DEF"
)

type testLinkService struct {
	createLinkFunc  func(ctx context.Context, originalURL string) (string, error)
	resolveLinkFunc func(ctx context.Context, shortCode string) (string, error)
}

func (s *testLinkService) CreateLink(ctx context.Context, originalURL string) (string, error) {
	if s.createLinkFunc != nil {
		return s.createLinkFunc(ctx, originalURL)
	}

	return testShortCode, nil
}

func (s *testLinkService) ResolveLink(ctx context.Context, shortCode string) (string, error) {
	if s.resolveLinkFunc != nil {
		return s.resolveLinkFunc(ctx, shortCode)
	}

	return testOriginalURL, nil
}

func newTestHandler(service linkService) *Handler {
	return NewHandler(service, zap.NewNop(), testBaseURL)
}

func TestHandlerCreateLink(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		body           string
		service        *testLinkService
		wantStatus     int
		wantShortURL   string
		wantErrorCode  string
		wantServiceURL string
	}{
		{
			name:           "success",
			body:           `{"url":"https://example.com"}`,
			service:        &testLinkService{},
			wantStatus:     http.StatusOK,
			wantShortURL:   testBaseURL + "/" + testShortCode,
			wantServiceURL: testOriginalURL,
		},
		{
			name:          "invalid json",
			body:          `{"url":`,
			service:       &testLinkService{},
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "invalid_json",
		},
		{
			name:          "unknown field",
			body:          `{"url":"https://example.com","extra":"value"}`,
			service:       &testLinkService{},
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "invalid_json",
		},
		{
			name:          "several json objects",
			body:          `{"url":"https://example.com"} {"url":"https://other.com"}`,
			service:       &testLinkService{},
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "invalid_json",
		},
		{
			name: "service invalid url error",
			body: `{"url":"bad-url"}`,
			service: &testLinkService{
				createLinkFunc: func(_ context.Context, _ string) (string, error) {
					return "", xerrors.ErrInvalidURL
				},
			},
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "invalid_url",
		},
		{
			name: "service storage error",
			body: `{"url":"https://example.com"}`,
			service: &testLinkService{
				createLinkFunc: func(_ context.Context, _ string) (string, error) {
					return "", xerrors.ErrStorageUnavailable
				},
			},
			wantStatus:    http.StatusServiceUnavailable,
			wantErrorCode: "storage_unavailable",
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var gotServiceURL string
			if tc.service.createLinkFunc == nil {
				tc.service.createLinkFunc = func(_ context.Context, originalURL string) (string, error) {
					gotServiceURL = originalURL
					return testShortCode, nil
				}
			}

			handler := newTestHandler(tc.service)

			req := httptest.NewRequest(
				http.MethodPost,
				"/api/v1/links",
				bytes.NewBufferString(tc.body),
			)
			rec := httptest.NewRecorder()

			handler.Router().ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("expected status %d, got %d, body: %s", tc.wantStatus, rec.Code, rec.Body.String())
			}

			if tc.wantErrorCode != "" {
				var response errorResponse
				decodeResponse(t, rec, &response)

				if response.Error.Code != tc.wantErrorCode {
					t.Fatalf("expected error code %q, got %q", tc.wantErrorCode, response.Error.Code)
				}

				return
			}

			var response createLinkResponse
			decodeResponse(t, rec, &response)

			if response.ShortURL != tc.wantShortURL {
				t.Fatalf("expected short url %q, got %q", tc.wantShortURL, response.ShortURL)
			}

			if tc.wantServiceURL != "" && gotServiceURL != tc.wantServiceURL {
				t.Fatalf("expected service url %q, got %q", tc.wantServiceURL, gotServiceURL)
			}
		})
	}
}

func TestHandlerResolveLink(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		code          string
		service       *testLinkService
		wantStatus    int
		wantURL       string
		wantErrorCode string
	}{
		{
			name:       "success",
			code:       testShortCode,
			service:    &testLinkService{},
			wantStatus: http.StatusOK,
			wantURL:    testOriginalURL,
		},
		{
			name: "not found",
			code: testShortCode,
			service: &testLinkService{
				resolveLinkFunc: func(_ context.Context, _ string) (string, error) {
					return "", xerrors.ErrNotFound
				},
			},
			wantStatus:    http.StatusNotFound,
			wantErrorCode: "not_found",
		},
		{
			name: "invalid short code",
			code: "bad",
			service: &testLinkService{
				resolveLinkFunc: func(_ context.Context, _ string) (string, error) {
					return "", xerrors.ErrInvalidShortCode
				},
			},
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "invalid_short_code",
		},
		{
			name: "storage unavailable",
			code: testShortCode,
			service: &testLinkService{
				resolveLinkFunc: func(_ context.Context, _ string) (string, error) {
					return "", xerrors.ErrStorageUnavailable
				},
			},
			wantStatus:    http.StatusServiceUnavailable,
			wantErrorCode: "storage_unavailable",
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			handler := newTestHandler(tc.service)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/links/"+tc.code, nil)
			rec := httptest.NewRecorder()

			handler.Router().ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("expected status %d, got %d, body: %s", tc.wantStatus, rec.Code, rec.Body.String())
			}

			if tc.wantErrorCode != "" {
				var response errorResponse
				decodeResponse(t, rec, &response)

				if response.Error.Code != tc.wantErrorCode {
					t.Fatalf("expected error code %q, got %q", tc.wantErrorCode, response.Error.Code)
				}

				return
			}

			var response resolveLinkResponse
			decodeResponse(t, rec, &response)

			if response.OriginalURL != tc.wantURL {
				t.Fatalf("expected original url %q, got %q", tc.wantURL, response.OriginalURL)
			}
		})
	}
}

func TestHandlerRedirect(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		path          string
		service       *testLinkService
		wantStatus    int
		wantLocation  string
		wantErrorCode string
	}{
		{
			name:         "success",
			path:         "/" + testShortCode,
			service:      &testLinkService{},
			wantStatus:   http.StatusFound,
			wantLocation: testOriginalURL,
		},
		{
			name:       "root is not found",
			path:       "/",
			service:    &testLinkService{},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "api path is not redirected",
			path:       "/api/something",
			service:    &testLinkService{},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "link not found",
			path: "/" + testShortCode,
			service: &testLinkService{
				resolveLinkFunc: func(_ context.Context, _ string) (string, error) {
					return "", xerrors.ErrNotFound
				},
			},
			wantStatus:    http.StatusNotFound,
			wantErrorCode: "not_found",
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			handler := newTestHandler(tc.service)

			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()

			handler.Router().ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("expected status %d, got %d, body: %s", tc.wantStatus, rec.Code, rec.Body.String())
			}

			if tc.wantLocation != "" {
				location := rec.Header().Get("Location")
				if location != tc.wantLocation {
					t.Fatalf("expected location %q, got %q", tc.wantLocation, location)
				}
			}

			if tc.wantErrorCode != "" {
				var response errorResponse
				decodeResponse(t, rec, &response)

				if response.Error.Code != tc.wantErrorCode {
					t.Fatalf("expected error code %q, got %q", tc.wantErrorCode, response.Error.Code)
				}
			}
		})
	}
}

func TestHandlerWriteJSON(t *testing.T) {
	t.Parallel()

	handler := newTestHandler(&testLinkService{})
	rec := httptest.NewRecorder()

	handler.writeJSON(rec, http.StatusCreated, createLinkResponse{
		ShortURL: testBaseURL + "/" + testShortCode,
	})

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Fatalf("expected content type %q, got %q", "application/json", contentType)
	}

	var response createLinkResponse
	decodeResponse(t, rec, &response)

	if response.ShortURL != testBaseURL+"/"+testShortCode {
		t.Fatalf("unexpected short url: %q", response.ShortURL)
	}
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

	if err := json.NewDecoder(rec.Body).Decode(target); err != nil {
		t.Fatalf("decode response body: %v, body: %s", err, rec.Body.String())
	}
}
