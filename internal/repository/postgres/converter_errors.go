package postgres

import (
	"errors"

	xerrors "github.com/Saperart/linklite-url-shortener/internal/errors"
	"github.com/jackc/pgx/v5/pgconn"
)

func mapSaveError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.ConstraintName {
		case "links_original_url_unique":
			return xerrors.ErrOriginalURLAlreadyExists
		case "links_short_code_unique":
			return xerrors.ErrShortCodeAlreadyExists
		}
	}

	return xerrors.ErrStorageUnavailable
}
