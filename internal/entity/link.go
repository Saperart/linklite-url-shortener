package entity

const (
	ShortCodeLength   = 10
	ShortCodeAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
)

type Link struct {
	OriginalURL string
	ShortCode   string
}

func NewLink(originalURL, shortCode string) *Link {
	return &Link{
		OriginalURL: originalURL,
		ShortCode:   shortCode,
	}

}
