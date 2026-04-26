package memory

import (
	"context"
	"sync"

	"github.com/Saperart/linklite-url-shortener/internal/entity"
	xerrors "github.com/Saperart/linklite-url-shortener/internal/errors"
)

type Repository struct {
	mx            sync.RWMutex
	byOriginalURL map[string]*entity.Link
	byShortCode   map[string]*entity.Link
}

func NewRepository() *Repository {
	return &Repository{
		byOriginalURL: make(map[string]*entity.Link),
		byShortCode:   make(map[string]*entity.Link),
	}
}

func (r *Repository) GetByOriginalURL(_ context.Context, originalURL string) (*entity.Link, error) {
	r.mx.RLock()
	defer r.mx.RUnlock()

	link, ok := r.byOriginalURL[originalURL]
	if !ok {
		return nil, xerrors.ErrNotFound
	}
	return link, nil
}

func (r *Repository) GetByShortCode(_ context.Context, shortCode string) (*entity.Link, error) {
	r.mx.RLock()
	defer r.mx.RUnlock()

	link, ok := r.byShortCode[shortCode]
	if !ok {
		return nil, xerrors.ErrNotFound
	}
	return link, nil
}

func (r *Repository) Save(_ context.Context, link *entity.Link) (*entity.Link, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

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
