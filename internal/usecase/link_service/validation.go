package link_service

import (
	"net/url"

	"github.com/Saperart/linklite-url-shortener/internal/entity"
	xerrors "github.com/Saperart/linklite-url-shortener/internal/errors"
)

func validateOriginalURL(raw string) error {
	if raw == "" {
		return xerrors.ErrInvalidURL
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return xerrors.ErrInvalidURL
	}

	if parsed.Host == "" {
		return xerrors.ErrInvalidURL
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return xerrors.ErrInvalidURL
	}

	return nil
}

func validateShortCode(code string) error {
	if len(code) != entity.ShortCodeLength {
		return xerrors.ErrInvalidShortCode
	}

	for _, symbol := range code {
		if !isAllowedShortCodeChar(symbol) {
			return xerrors.ErrInvalidShortCode
		}
	}

	return nil
}

func isAllowedShortCodeChar(symbol rune) bool {
	for _, allowed := range entity.ShortCodeAlphabet {
		if symbol == allowed {
			return true
		}
	}

	return false
}
