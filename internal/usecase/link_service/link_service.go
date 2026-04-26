package link_service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Saperart/linklite-url-shortener/internal/entity"
	xerrors "github.com/Saperart/linklite-url-shortener/internal/errors"
)

const maxGenerateCodeAttempts = 100

type LinkRepository interface {
	GetByOriginalURL(ctx context.Context, originalURL string) (*entity.Link, error)
	GetByShortCode(ctx context.Context, shortCode string) (*entity.Link, error)
	Save(ctx context.Context, link *entity.Link) (*entity.Link, error)
}

type CodeGenerator interface {
	Generate() (string, error)
}

type LinkService struct {
	repo      LinkRepository
	generator CodeGenerator
}

func NewLinkService(repo LinkRepository, generator CodeGenerator) *LinkService {
	return &LinkService{
		repo:      repo,
		generator: generator,
	}
}

func (s *LinkService) CreateLink(ctx context.Context, originalURL string) (string, error) {
	if err := validateOriginalURL(originalURL); err != nil {
		return "", err
	}

	existing, err := s.repo.GetByOriginalURL(ctx, originalURL)
	switch {
	case err == nil:
		return existing.ShortCode, nil
	case errors.Is(err, xerrors.ErrNotFound):
		// ссылки ещё нет, идём создавать новую
	default:
		return "", err
	}

	for attempt := 0; attempt < maxGenerateCodeAttempts; attempt++ {
		code, err := s.generator.Generate()
		if err != nil {
			return "", fmt.Errorf("generate short code: %w", err)
		}

		if err := validateShortCode(code); err != nil {
			return "", fmt.Errorf("generator produced invalid short code: %w", err)
		}

		saved, err := s.repo.Save(ctx, entity.NewLink(originalURL, code))
		if err == nil {
			return saved.ShortCode, nil
		}

		if errors.Is(err, xerrors.ErrShortCodeAlreadyExists) {
			continue
		}

		if errors.Is(err, xerrors.ErrOriginalURLAlreadyExists) {
			existing, getErr := s.repo.GetByOriginalURL(ctx, originalURL)
			if getErr != nil {
				return "", getErr
			}

			return existing.ShortCode, nil
		}

		return "", err
	}

	return "", xerrors.ErrMaxRetriesExceeded
}

func (s *LinkService) ResolveLink(ctx context.Context, shortCode string) (string, error) {
	if err := validateShortCode(shortCode); err != nil {
		return "", err
	}

	link, err := s.repo.GetByShortCode(ctx, shortCode)
	if err != nil {
		return "", err
	}

	return link.OriginalURL, nil
}
