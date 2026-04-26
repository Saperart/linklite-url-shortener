package generator

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"unicode/utf8"

	xerrors "github.com/Saperart/linklite-url-shortener/internal/errors"
)

type RandomCodeGenerator struct {
	length   int
	alphabet string
}

func NewRandomCodeGenerator(length int, alphabet string) (*RandomCodeGenerator, error) {
	if length <= 0 {
		return nil, xerrors.ErrInvalidCodeLength
	}

	if alphabet == "" {
		return nil, xerrors.ErrEmptyAlphabet
	}

	if len(alphabet) != utf8.RuneCountInString(alphabet) {
		return nil, xerrors.ErrInvalidAlphabet
	}

	return &RandomCodeGenerator{
		length:   length,
		alphabet: alphabet,
	}, nil
}

func (g *RandomCodeGenerator) Generate() (string, error) {
	result := make([]byte, g.length)
	maximum := big.NewInt(int64(len(g.alphabet)))

	for i := range result {
		n, err := rand.Int(rand.Reader, maximum)
		if err != nil {
			return "", fmt.Errorf("generate random index: %w", err)
		}

		result[i] = g.alphabet[n.Int64()]
	}

	return string(result), nil
}
