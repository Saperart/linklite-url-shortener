package postgres

import (
	"context"
	"errors"

	"github.com/Saperart/linklite-url-shortener/internal/entity"
	xerrors "github.com/Saperart/linklite-url-shortener/internal/errors"
	"github.com/Saperart/linklite-url-shortener/internal/repository/postgres/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		queries: sqlc.New(pool),
	}
}

func (r *Repository) GetByOriginalURL(ctx context.Context, originalURL string) (*entity.Link, error) {
	link, err := r.queries.GetLinkByOriginalURL(ctx, originalURL)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, xerrors.ErrNotFound
	}
	if err != nil {
		return nil, xerrors.ErrStorageUnavailable
	}

	return toEntityLink(link), nil
}

func (r *Repository) GetByShortCode(ctx context.Context, shortCode string) (*entity.Link, error) {
	link, err := r.queries.GetLinkByShortCode(ctx, shortCode)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, xerrors.ErrNotFound
	}
	if err != nil {
		return nil, xerrors.ErrStorageUnavailable
	}

	return toEntityLink(link), nil
}

func (r *Repository) Save(ctx context.Context, link *entity.Link) (*entity.Link, error) {
	created, err := r.queries.CreateLink(ctx, sqlc.CreateLinkParams{
		OriginalUrl: link.OriginalURL,
		ShortCode:   link.ShortCode,
	})
	if err != nil {
		return nil, mapSaveError(err)
	}

	return toEntityLink(created), nil
}
