package memory

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/Saperart/linklite-url-shortener/internal/entity"
	xerrors "github.com/Saperart/linklite-url-shortener/internal/errors"
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
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				if saved != nil {
					t.Fatalf("expected nil saved link, got %#v", saved)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if saved == nil {
				t.Fatal("expected saved link, got nil")
			}
			if saved.OriginalURL != tc.link.OriginalURL {
				t.Fatalf("expected original url %q, got %q", tc.link.OriginalURL, saved.OriginalURL)
			}
			if saved.ShortCode != tc.link.ShortCode {
				t.Fatalf("expected short code %q, got %q", tc.link.ShortCode, saved.ShortCode)
			}
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
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				if link != nil {
					t.Fatalf("expected nil link, got %#v", link)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if link == nil {
				t.Fatal("expected link, got nil")
			}
			if link.OriginalURL != tc.originalURL {
				t.Fatalf("expected original url %q, got %q", tc.originalURL, link.OriginalURL)
			}
			if link.ShortCode != tc.wantCode {
				t.Fatalf("expected short code %q, got %q", tc.wantCode, link.ShortCode)
			}
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
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				if link != nil {
					t.Fatalf("expected nil link, got %#v", link)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if link == nil {
				t.Fatal("expected link, got nil")
			}
			if link.ShortCode != tc.shortCode {
				t.Fatalf("expected short code %q, got %q", tc.shortCode, link.ShortCode)
			}
			if link.OriginalURL != tc.wantURL {
				t.Fatalf("expected original url %q, got %q", tc.wantURL, link.OriginalURL)
			}
		})
	}
}

func TestRepositorySaveAndReadBothIndexes(t *testing.T) {
	t.Parallel()

	repo := NewRepository()
	link := entity.NewLink(testOriginalURL, testShortCode)

	saved, err := repo.Save(context.Background(), link)
	if err != nil {
		t.Fatalf("save link: %v", err)
	}

	byOriginalURL, err := repo.GetByOriginalURL(context.Background(), testOriginalURL)
	if err != nil {
		t.Fatalf("get by original url: %v", err)
	}

	byShortCode, err := repo.GetByShortCode(context.Background(), testShortCode)
	if err != nil {
		t.Fatalf("get by short code: %v", err)
	}

	if saved != byOriginalURL {
		t.Fatal("expected GetByOriginalURL to return the same link pointer as Save")
	}

	if saved != byShortCode {
		t.Fatal("expected GetByShortCode to return the same link pointer as Save")
	}
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
		if err != nil {
			t.Fatalf("unexpected save error: %v", err)
		}
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

	if successCount != 1 {
		t.Fatalf("expected exactly one successful save, got %d", successCount)
	}

	if duplicateCount != workers-1 {
		t.Fatalf("expected %d duplicate errors, got %d", workers-1, duplicateCount)
	}
}
func makeShortCode(index int) string {
	return fmt.Sprintf("Code%06d", index)
}
