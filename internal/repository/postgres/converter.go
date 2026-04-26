package postgres

import (
	"github.com/Saperart/linklite-url-shortener/internal/entity"
	"github.com/Saperart/linklite-url-shortener/internal/repository/postgres/sqlc"
)

func toEntityLink(link sqlc.Link) *entity.Link {
	return &entity.Link{
		OriginalURL: link.OriginalUrl,
		ShortCode:   link.ShortCode,
	}
}
