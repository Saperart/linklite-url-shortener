//go:build integration

package postgres

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/Saperart/linklite-url-shortener/internal/entity"
	xerrors "github.com/Saperart/linklite-url-shortener/internal/errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testOriginalURLPrefix = "https://integration.test/"
	testCodeAlphabet      = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
	testCodeLength        = 10
)

func TestRepositoryIntegrationSaveAndGetByOriginalURL(t *testing.T) {
	pool := newTestPool(t)
	repo := NewRepository(pool)

	originalURL, shortCode := newTestLinkData(t)

	saved, err := repo.Save(context.Background(), entity.NewLink(originalURL, shortCode))
	require.NoError(t, err)
	assert.Equal(t, originalURL, saved.OriginalURL)
	assert.Equal(t, shortCode, saved.ShortCode)

	found, err := repo.GetByOriginalURL(context.Background(), originalURL)
	require.NoError(t, err)
	assert.Equal(t, originalURL, found.OriginalURL)
	assert.Equal(t, shortCode, found.ShortCode)
}

func TestRepositoryIntegrationSaveAndGetByShortCode(t *testing.T) {
	pool := newTestPool(t)
	repo := NewRepository(pool)

	originalURL, shortCode := newTestLinkData(t)

	_, err := repo.Save(context.Background(), entity.NewLink(originalURL, shortCode))
	require.NoError(t, err)

	found, err := repo.GetByShortCode(context.Background(), shortCode)
	require.NoError(t, err)
	assert.Equal(t, originalURL, found.OriginalURL)
	assert.Equal(t, shortCode, found.ShortCode)
}

func TestRepositoryIntegrationGetNotFound(t *testing.T) {
	pool := newTestPool(t)
	repo := NewRepository(pool)

	t.Run("get by original url", func(t *testing.T) {
		_, err := repo.GetByOriginalURL(context.Background(), testOriginalURLPrefix+"not-found")

		assert.ErrorIs(t, err, xerrors.ErrNotFound)
	})

	t.Run("get by short code", func(t *testing.T) {
		_, err := repo.GetByShortCode(context.Background(), "NotFound1_")

		assert.ErrorIs(t, err, xerrors.ErrNotFound)
	})
}

func TestRepositoryIntegrationSaveDuplicateOriginalURL(t *testing.T) {
	pool := newTestPool(t)
	repo := NewRepository(pool)

	originalURL, shortCode := newTestLinkData(t)
	_, anotherShortCode := newTestLinkData(t)

	_, err := repo.Save(context.Background(), entity.NewLink(originalURL, shortCode))
	require.NoError(t, err)

	_, err = repo.Save(context.Background(), entity.NewLink(originalURL, anotherShortCode))
	assert.ErrorIs(t, err, xerrors.ErrOriginalURLAlreadyExists)
}

func TestRepositoryIntegrationSaveDuplicateShortCode(t *testing.T) {
	pool := newTestPool(t)
	repo := NewRepository(pool)

	originalURL, shortCode := newTestLinkData(t)
	anotherOriginalURL := testOriginalURLPrefix + "another-" + shortCode

	_, err := repo.Save(context.Background(), entity.NewLink(originalURL, shortCode))
	require.NoError(t, err)

	_, err = repo.Save(context.Background(), entity.NewLink(anotherOriginalURL, shortCode))
	assert.ErrorIs(t, err, xerrors.ErrShortCodeAlreadyExists)
}

func newTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("create pgx pool: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("ping postgres: %v", err)
	}

	cleanupTestLinks(t, pool)

	t.Cleanup(func() {
		cleanupTestLinks(t, pool)
		pool.Close()
	})

	return pool
}

func cleanupTestLinks(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := pool.Exec(ctx, `
		DELETE FROM links
		WHERE original_url LIKE $1
	`, testOriginalURLPrefix+"%")
	if err != nil {
		t.Fatalf("cleanup test links: %v", err)
	}
}

func newTestLinkData(t *testing.T) (string, string) {
	t.Helper()

	shortCode := randomShortCode(t)
	originalURL := fmt.Sprintf("%s%s", testOriginalURLPrefix, shortCode)

	return originalURL, shortCode
}

func randomShortCode(t *testing.T) string {
	t.Helper()

	result := make([]byte, testCodeLength)
	maximum := big.NewInt(int64(len(testCodeAlphabet)))

	for i := range result {
		n, err := rand.Int(rand.Reader, maximum)
		if err != nil {
			t.Fatalf("generate random short code: %v", err)
		}

		result[i] = testCodeAlphabet[n.Int64()]
	}

	return string(result)
}
