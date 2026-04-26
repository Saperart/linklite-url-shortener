package link_service

import (
	"context"
	"errors"
	"testing"

	"github.com/Saperart/linklite-url-shortener/internal/entity"
	xerrors "github.com/Saperart/linklite-url-shortener/internal/errors"
)

const (
	validOriginalURL = "https://example.com"
	validShortCode   = "Abc123_DEF"
	secondShortCode  = "Xyz987_QWE"
)

type testRepository struct {
	byOriginalURL map[string]*entity.Link
	byShortCode   map[string]*entity.Link

	getByOriginalURLError error
	getByShortCodeError   error
	saveError             error
}

func newTestRepository() *testRepository {
	return &testRepository{
		byOriginalURL: make(map[string]*entity.Link),
		byShortCode:   make(map[string]*entity.Link),
	}
}

func (r *testRepository) GetByOriginalURL(_ context.Context, originalURL string) (*entity.Link, error) {
	if r.getByOriginalURLError != nil {
		return nil, r.getByOriginalURLError
	}

	link, ok := r.byOriginalURL[originalURL]
	if !ok {
		return nil, xerrors.ErrNotFound
	}

	return link, nil
}

func (r *testRepository) GetByShortCode(_ context.Context, shortCode string) (*entity.Link, error) {
	if r.getByShortCodeError != nil {
		return nil, r.getByShortCodeError
	}

	link, ok := r.byShortCode[shortCode]
	if !ok {
		return nil, xerrors.ErrNotFound
	}

	return link, nil
}

func (r *testRepository) Save(_ context.Context, link *entity.Link) (*entity.Link, error) {
	if r.saveError != nil {
		return nil, r.saveError
	}

	if _, ok := r.byOriginalURL[link.OriginalURL]; ok {
		return nil, xerrors.ErrOriginalURLAlreadyExists
	}

	if _, ok := r.byShortCode[link.ShortCode]; ok {
		return nil, xerrors.ErrShortCodeAlreadyExists
	}

	r.byOriginalURL[link.OriginalURL] = link
	r.byShortCode[link.ShortCode] = link

	return link, nil
}

type testGenerator struct {
	codes []string
	err   error
	calls int
}

func (g *testGenerator) Generate() (string, error) {
	g.calls++

	if g.err != nil {
		return "", g.err
	}

	if len(g.codes) == 0 {
		return validShortCode, nil
	}

	code := g.codes[0]
	if len(g.codes) > 1 {
		g.codes = g.codes[1:]
	}

	return code, nil
}

func TestLinkServiceCreateLink(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		url       string
		prepare   func(repo *testRepository, generator *testGenerator)
		wantCode  string
		wantErr   error
		wantCalls int
	}{
		{
			name:     "create new link",
			url:      validOriginalURL,
			wantCode: validShortCode,
		},
		{
			name: "return existing link for same original url",
			url:  validOriginalURL,
			prepare: func(repo *testRepository, _ *testGenerator) {
				link := entity.NewLink(validOriginalURL, secondShortCode)
				repo.byOriginalURL[link.OriginalURL] = link
				repo.byShortCode[link.ShortCode] = link
			},
			wantCode:  secondShortCode,
			wantCalls: 0,
		},
		{
			name:    "invalid url without scheme",
			url:     "example.com",
			wantErr: xerrors.ErrInvalidURL,
		},
		{
			name: "repository get error",
			url:  validOriginalURL,
			prepare: func(repo *testRepository, _ *testGenerator) {
				repo.getByOriginalURLError = xerrors.ErrStorageUnavailable
			},
			wantErr: xerrors.ErrStorageUnavailable,
		},
		{
			name: "retry after short code collision",
			url:  validOriginalURL,
			prepare: func(repo *testRepository, generator *testGenerator) {
				existing := entity.NewLink("https://other.com", validShortCode)
				repo.byOriginalURL[existing.OriginalURL] = existing
				repo.byShortCode[existing.ShortCode] = existing

				generator.codes = []string{validShortCode, secondShortCode}
			},
			wantCode:  secondShortCode,
			wantCalls: 2,
		},
		{
			name: "max retries exceeded",
			url:  validOriginalURL,
			prepare: func(repo *testRepository, generator *testGenerator) {
				existing := entity.NewLink("https://other.com", validShortCode)
				repo.byOriginalURL[existing.OriginalURL] = existing
				repo.byShortCode[existing.ShortCode] = existing

				generator.codes = []string{validShortCode}
			},
			wantErr:   xerrors.ErrMaxRetriesExceeded,
			wantCalls: maxGenerateCodeAttempts,
		},
		{
			name: "generator produced invalid short code",
			url:  validOriginalURL,
			prepare: func(_ *testRepository, generator *testGenerator) {
				generator.codes = []string{"bad"}
			},
			wantErr: xerrors.ErrInvalidShortCode,
		},
		{
			name: "save storage error",
			url:  validOriginalURL,
			prepare: func(repo *testRepository, _ *testGenerator) {
				repo.saveError = xerrors.ErrStorageUnavailable
			},
			wantErr: xerrors.ErrStorageUnavailable,
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := newTestRepository()
			generator := &testGenerator{codes: []string{validShortCode}}

			if tc.prepare != nil {
				tc.prepare(repo, generator)
			}

			service := NewLinkService(repo, generator)

			code, err := service.CreateLink(context.Background(), tc.url)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				if code != "" {
					t.Fatalf("expected empty code, got %q", code)
				}

				if tc.wantCalls != 0 && generator.calls != tc.wantCalls {
					t.Fatalf("expected generator calls %d, got %d", tc.wantCalls, generator.calls)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if code != tc.wantCode {
				t.Fatalf("expected code %q, got %q", tc.wantCode, code)
			}

			if tc.wantCalls != 0 && generator.calls != tc.wantCalls {
				t.Fatalf("expected generator calls %d, got %d", tc.wantCalls, generator.calls)
			}
		})
	}
}

func TestLinkServiceResolveLink(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		shortCode string
		prepare   func(repo *testRepository)
		wantURL   string
		wantErr   error
	}{
		{
			name:      "success",
			shortCode: validShortCode,
			prepare: func(repo *testRepository) {
				link := entity.NewLink(validOriginalURL, validShortCode)
				repo.byOriginalURL[link.OriginalURL] = link
				repo.byShortCode[link.ShortCode] = link
			},
			wantURL: validOriginalURL,
		},
		{
			name:      "invalid short code length",
			shortCode: "short",
			wantErr:   xerrors.ErrInvalidShortCode,
		},
		{
			name:      "invalid short code alphabet",
			shortCode: "Abc123-DE!",
			wantErr:   xerrors.ErrInvalidShortCode,
		},
		{
			name:      "not found",
			shortCode: validShortCode,
			wantErr:   xerrors.ErrNotFound,
		},
		{
			name:      "repository error",
			shortCode: validShortCode,
			prepare: func(repo *testRepository) {
				repo.getByShortCodeError = xerrors.ErrStorageUnavailable
			},
			wantErr: xerrors.ErrStorageUnavailable,
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			repo := newTestRepository()
			generator := &testGenerator{codes: []string{validShortCode}}
			if tc.prepare != nil {
				tc.prepare(repo)
			}

			service := NewLinkService(repo, generator)
			originalURL, err := service.ResolveLink(context.Background(), tc.shortCode)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				if originalURL != "" {
					t.Fatalf("expected empty original url, got %q", originalURL)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if originalURL != tc.wantURL {
				t.Fatalf("expected original url %q, got %q", tc.wantURL, originalURL)
			}
		})
	}
}

func TestValidateOriginalURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		rawURL  string
		wantErr error
	}{
		{
			name:   "valid https url",
			rawURL: "https://example.com",
		},
		{
			name:   "valid http url with path",
			rawURL: "http://example.com/path?q=1",
		},
		{
			name:    "empty url",
			rawURL:  "",
			wantErr: xerrors.ErrInvalidURL,
		},
		{
			name:    "url without scheme",
			rawURL:  "example.com",
			wantErr: xerrors.ErrInvalidURL,
		},
		{
			name:    "unsupported scheme",
			rawURL:  "ftp://example.com",
			wantErr: xerrors.ErrInvalidURL,
		},
		{
			name:    "empty host",
			rawURL:  "https://",
			wantErr: xerrors.ErrInvalidURL,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := validateOriginalURL(tc.rawURL)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateShortCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		shortCode string
		wantErr   error
	}{
		{
			name:      "valid short code",
			shortCode: validShortCode,
		},
		{
			name:      "too short",
			shortCode: "short",
			wantErr:   xerrors.ErrInvalidShortCode,
		},
		{
			name:      "too long",
			shortCode: "Abc123_DEFx",
			wantErr:   xerrors.ErrInvalidShortCode,
		},
		{
			name:      "contains dash",
			shortCode: "Abc123-DE!",
			wantErr:   xerrors.ErrInvalidShortCode,
		},
		{
			name:      "contains space",
			shortCode: "Abc123 DE!",
			wantErr:   xerrors.ErrInvalidShortCode,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := validateShortCode(tc.shortCode)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestLinkServiceCreateLinkGeneratorError(t *testing.T) {
	t.Parallel()

	repo := newTestRepository()
	generatorErr := errors.New("random source failed")
	generator := &testGenerator{err: generatorErr}
	service := NewLinkService(repo, generator)

	code, err := service.CreateLink(context.Background(), validOriginalURL)

	if !errors.Is(err, generatorErr) {
		t.Fatalf("expected generator error %v, got %v", generatorErr, err)
	}

	if code != "" {
		t.Fatalf("expected empty code, got %q", code)
	}
}
