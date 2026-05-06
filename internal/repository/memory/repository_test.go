package memory

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/Saperart/linklite-url-shortener/internal/entity"
	xerrors "github.com/Saperart/linklite-url-shortener/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testOriginalURL = "https://example.com"
	testShortCode   = "Abc123_DEF"

	otherOriginalURL = "https://other.com"
	otherShortCode   = "Xyz987_QWE"
)

func TestRepositorySave(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		prepare func(repo *Repository)
		link    *entity.Link
		wantErr error
	}{
		{
			name: "success",
			link: entity.NewLink(testOriginalURL, testShortCode),
		},
		{
			name: "duplicate original url",
			prepare: func(repo *Repository) {
				_, _ = repo.Save(context.Background(), entity.NewLink(testOriginalURL, testShortCode))
			},
			link:    entity.NewLink(testOriginalURL, otherShortCode),
			wantErr: xerrors.ErrOriginalURLAlreadyExists,
		},
		{
			name: "duplicate short code",
			prepare: func(repo *Repository) {
				_, _ = repo.Save(context.Background(), entity.NewLink(testOriginalURL, testShortCode))
			},
			link:    entity.NewLink(otherOriginalURL, testShortCode),
			wantErr: xerrors.ErrShortCodeAlreadyExists,
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := NewRepository()
			if tc.prepare != nil {
				tc.prepare(repo)
			}

			saved, err := repo.Save(context.Background(), tc.link)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Nil(t, saved)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, saved)
			assert.Equal(t, tc.link.OriginalURL, saved.OriginalURL)
			assert.Equal(t, tc.link.ShortCode, saved.ShortCode)
		})
	}
}

func TestRepositoryGetByOriginalURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		prepare     func(repo *Repository)
		originalURL string
		wantCode    string
		wantErr     error
	}{
		{
			name:        "found",
			originalURL: testOriginalURL,
			prepare: func(repo *Repository) {
				_, _ = repo.Save(context.Background(), entity.NewLink(testOriginalURL, testShortCode))
			},
			wantCode: testShortCode,
		},
		{
			name:        "not found",
			originalURL: testOriginalURL,
			wantErr:     xerrors.ErrNotFound,
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := NewRepository()
			if tc.prepare != nil {
				tc.prepare(repo)
			}

			link, err := repo.GetByOriginalURL(context.Background(), tc.originalURL)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Nil(t, link)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, link)
			assert.Equal(t, tc.originalURL, link.OriginalURL)
			assert.Equal(t, tc.wantCode, link.ShortCode)
		})
	}
}

func TestRepositoryGetByShortCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		prepare   func(repo *Repository)
		shortCode string
		wantURL   string
		wantErr   error
	}{
		{
			name:      "found",
			shortCode: testShortCode,
			prepare: func(repo *Repository) {
				_, _ = repo.Save(context.Background(), entity.NewLink(testOriginalURL, testShortCode))
			},
			wantURL: testOriginalURL,
		},
		{
			name:      "not found",
			shortCode: testShortCode,
			wantErr:   xerrors.ErrNotFound,
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := NewRepository()
			if tc.prepare != nil {
				tc.prepare(repo)
			}

			link, err := repo.GetByShortCode(context.Background(), tc.shortCode)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Nil(t, link)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, link)
			assert.Equal(t, tc.shortCode, link.ShortCode)
			assert.Equal(t, tc.wantURL, link.OriginalURL)
		})
	}
}

func TestRepositorySaveAndReadBothIndexes(t *testing.T) {
	t.Parallel()

	repo := NewRepository()
	link := entity.NewLink(testOriginalURL, testShortCode)

	saved, err := repo.Save(context.Background(), link)
	require.NoError(t, err)

	byOriginalURL, err := repo.GetByOriginalURL(context.Background(), testOriginalURL)
	require.NoError(t, err)

	byShortCode, err := repo.GetByShortCode(context.Background(), testShortCode)
	require.NoError(t, err)

	assert.Same(t, saved, byOriginalURL)
	assert.Same(t, saved, byShortCode)
}

func TestRepositoryConcurrentSaveDifferentLinks(t *testing.T) {
	t.Parallel()

	const workers = 100

	repo := NewRepository()
	var wg sync.WaitGroup
	errs := make(chan error, workers)

	for i := 0; i < workers; i++ {
		i := i

		wg.Add(1)
		go func() {
			defer wg.Done()

			link := entity.NewLink(
				fmt.Sprintf("https://example.com/page/%d", i),
				makeShortCode(i),
			)

			_, err := repo.Save(context.Background(), link)
			errs <- err
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		require.NoError(t, err)
	}
}

func TestRepositoryConcurrentDuplicateShortCode(t *testing.T) {
	t.Parallel()

	const workers = 20

	repo := NewRepository()
	var wg sync.WaitGroup
	errs := make(chan error, workers)

	for i := 0; i < workers; i++ {
		i := i

		wg.Add(1)
		go func() {
			defer wg.Done()

			link := entity.NewLink(
				fmt.Sprintf("https://example.com/page/%d", i),
				testShortCode,
			)

			_, err := repo.Save(context.Background(), link)
			errs <- err
		}()
	}

	wg.Wait()
	close(errs)

	successCount := 0
	duplicateCount := 0

	for err := range errs {
		switch {
		case err == nil:
			successCount++
		case errors.Is(err, xerrors.ErrShortCodeAlreadyExists):
			duplicateCount++
		default:
			t.Fatalf("unexpected error: %v", err)
		}
	}

	assert.Equal(t, 1, successCount)
	assert.Equal(t, workers-1, duplicateCount)
}
func makeShortCode(index int) string {
	return fmt.Sprintf("Code%06d", index)
}
