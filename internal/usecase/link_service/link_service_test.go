package link_service

import (
	"context"
	"errors"
	"testing"

	"github.com/Saperart/linklite-url-shortener/internal/entity"
	xerrors "github.com/Saperart/linklite-url-shortener/internal/errors"
	"github.com/Saperart/linklite-url-shortener/internal/usecase/link_service/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	validOriginalURL = "https://example.com"
	validShortCode   = "Abc123_DEF"
	secondShortCode  = "Xyz987_QWE"
)

var errRandomSourceFailed = errors.New("random source failed")

func TestLinkServiceCreateLink(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		url     string
		prepare func(repo *mocks.MockLinkRepository, generator *mocks.MockCodeGenerator)
		want    string
		wantErr error
	}{
		{
			name: "create new link",
			url:  validOriginalURL,
			prepare: func(repo *mocks.MockLinkRepository, generator *mocks.MockCodeGenerator) {
				repo.EXPECT().
					GetByOriginalURL(gomock.Any(), validOriginalURL).
					Return(nil, xerrors.ErrNotFound)
				generator.EXPECT().Generate().Return(validShortCode, nil)
				repo.EXPECT().
					Save(gomock.Any(), gomock.AssignableToTypeOf(&entity.Link{})).
					DoAndReturn(func(_ context.Context, link *entity.Link) (*entity.Link, error) {
						return link, nil
					})
			},
			want: validShortCode,
		},
		{
			name: "return existing link for same original url",
			url:  validOriginalURL,
			prepare: func(repo *mocks.MockLinkRepository, _ *mocks.MockCodeGenerator) {
				repo.EXPECT().
					GetByOriginalURL(gomock.Any(), validOriginalURL).
					Return(entity.NewLink(validOriginalURL, secondShortCode), nil)
			},
			want: secondShortCode,
		},
		{
			name:    "invalid url without scheme",
			url:     "example.com",
			wantErr: xerrors.ErrInvalidURL,
		},
		{
			name: "repository get error",
			url:  validOriginalURL,
			prepare: func(repo *mocks.MockLinkRepository, _ *mocks.MockCodeGenerator) {
				repo.EXPECT().
					GetByOriginalURL(gomock.Any(), validOriginalURL).
					Return(nil, xerrors.ErrStorageUnavailable)
			},
			wantErr: xerrors.ErrStorageUnavailable,
		},
		{
			name: "retry after short code collision",
			url:  validOriginalURL,
			prepare: func(repo *mocks.MockLinkRepository, generator *mocks.MockCodeGenerator) {
				repo.EXPECT().
					GetByOriginalURL(gomock.Any(), validOriginalURL).
					Return(nil, xerrors.ErrNotFound)
				generator.EXPECT().Generate().Return(validShortCode, nil)
				repo.EXPECT().
					Save(gomock.Any(), gomock.AssignableToTypeOf(&entity.Link{})).
					Return(nil, xerrors.ErrShortCodeAlreadyExists)
				generator.EXPECT().Generate().Return(secondShortCode, nil)
				repo.EXPECT().
					Save(gomock.Any(), gomock.AssignableToTypeOf(&entity.Link{})).
					DoAndReturn(func(_ context.Context, link *entity.Link) (*entity.Link, error) {
						return link, nil
					})
			},
			want: secondShortCode,
		},
		{
			name: "max retries exceeded",
			url:  validOriginalURL,
			prepare: func(repo *mocks.MockLinkRepository, generator *mocks.MockCodeGenerator) {
				repo.EXPECT().
					GetByOriginalURL(gomock.Any(), validOriginalURL).
					Return(nil, xerrors.ErrNotFound)
				generator.EXPECT().
					Generate().
					Return(validShortCode, nil).
					Times(maxGenerateCodeAttempts)
				repo.EXPECT().
					Save(gomock.Any(), gomock.AssignableToTypeOf(&entity.Link{})).
					Return(nil, xerrors.ErrShortCodeAlreadyExists).
					Times(maxGenerateCodeAttempts)
			},
			wantErr: xerrors.ErrMaxRetriesExceeded,
		},
		{
			name: "generator produced invalid short code",
			url:  validOriginalURL,
			prepare: func(repo *mocks.MockLinkRepository, generator *mocks.MockCodeGenerator) {
				repo.EXPECT().
					GetByOriginalURL(gomock.Any(), validOriginalURL).
					Return(nil, xerrors.ErrNotFound)
				generator.EXPECT().Generate().Return("bad", nil)
			},
			wantErr: xerrors.ErrInvalidShortCode,
		},
		{
			name: "save storage error",
			url:  validOriginalURL,
			prepare: func(repo *mocks.MockLinkRepository, generator *mocks.MockCodeGenerator) {
				repo.EXPECT().
					GetByOriginalURL(gomock.Any(), validOriginalURL).
					Return(nil, xerrors.ErrNotFound)
				generator.EXPECT().Generate().Return(validShortCode, nil)
				repo.EXPECT().
					Save(gomock.Any(), gomock.AssignableToTypeOf(&entity.Link{})).
					Return(nil, xerrors.ErrStorageUnavailable)
			},
			wantErr: xerrors.ErrStorageUnavailable,
		},
		{
			name: "generator error",
			url:  validOriginalURL,
			prepare: func(repo *mocks.MockLinkRepository, generator *mocks.MockCodeGenerator) {
				repo.EXPECT().
					GetByOriginalURL(gomock.Any(), validOriginalURL).
					Return(nil, xerrors.ErrNotFound)
				generator.EXPECT().Generate().Return("", errRandomSourceFailed)
			},
			wantErr: errRandomSourceFailed,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			repo := mocks.NewMockLinkRepository(ctrl)
			generator := mocks.NewMockCodeGenerator(ctrl)

			if tc.prepare != nil {
				tc.prepare(repo, generator)
			}

			service := NewLinkService(repo, generator)

			code, err := service.CreateLink(context.Background(), tc.url)
			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Empty(t, code)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, code)
		})
	}
}

func TestLinkServiceResolveLink(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		shortCode string
		prepare   func(repo *mocks.MockLinkRepository)
		wantURL   string
		wantErr   error
	}{
		{
			name:      "success",
			shortCode: validShortCode,
			prepare: func(repo *mocks.MockLinkRepository) {
				repo.EXPECT().
					GetByShortCode(gomock.Any(), validShortCode).
					Return(entity.NewLink(validOriginalURL, validShortCode), nil)
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
			prepare: func(repo *mocks.MockLinkRepository) {
				repo.EXPECT().
					GetByShortCode(gomock.Any(), validShortCode).
					Return(nil, xerrors.ErrNotFound)
			},
			wantErr: xerrors.ErrNotFound,
		},
		{
			name:      "repository error",
			shortCode: validShortCode,
			prepare: func(repo *mocks.MockLinkRepository) {
				repo.EXPECT().
					GetByShortCode(gomock.Any(), validShortCode).
					Return(nil, xerrors.ErrStorageUnavailable)
			},
			wantErr: xerrors.ErrStorageUnavailable,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			repo := mocks.NewMockLinkRepository(ctrl)
			generator := mocks.NewMockCodeGenerator(ctrl)
			if tc.prepare != nil {
				tc.prepare(repo)
			}

			service := NewLinkService(repo, generator)
			originalURL, err := service.ResolveLink(context.Background(), tc.shortCode)
			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Empty(t, originalURL)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantURL, originalURL)
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
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
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
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
