package errors

import "errors"

var (
	ErrNotFound                 = errors.New("not found")
	ErrInvalidURL               = errors.New("invalid url")
	ErrInvalidShortCode         = errors.New("invalid short code")
	ErrShortCodeAlreadyExists   = errors.New("short code already exists")
	ErrOriginalURLAlreadyExists = errors.New("original url already exists")
	ErrStorageUnavailable       = errors.New("storage unavailable")
	ErrMaxRetriesExceeded       = errors.New("max retries exceeded")

	ErrInvalidCodeLength = errors.New("code length must be positive")
	ErrEmptyAlphabet     = errors.New("alphabet must not be empty")
	ErrInvalidAlphabet   = errors.New("invalid alphabet")
)
