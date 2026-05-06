package generator

import (
	"testing"

	xerrors "github.com/Saperart/linklite-url-shortener/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testCodeLength   = 10
	testCodeAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
)

func TestNewRandomCodeGenerator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		length   int
		alphabet string
		wantErr  error
	}{
		{
			name:     "zero length",
			length:   0,
			alphabet: testCodeAlphabet,
			wantErr:  xerrors.ErrInvalidCodeLength,
		},
		{
			name:     "negative length",
			length:   -1,
			alphabet: testCodeAlphabet,
			wantErr:  xerrors.ErrInvalidCodeLength,
		},
		{
			name:     "empty alphabet",
			length:   testCodeLength,
			alphabet: "",
			wantErr:  xerrors.ErrEmptyAlphabet,
		},
		{
			name:     "multibyte alphabet symbols",
			length:   testCodeLength,
			alphabet: "abcд",
			wantErr:  xerrors.ErrInvalidAlphabet,
		},
		{
			name:     "valid config",
			length:   testCodeLength,
			alphabet: testCodeAlphabet,
			wantErr:  nil,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			g, err := NewRandomCodeGenerator(tc.length, tc.alphabet)
			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Nil(t, g)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, g)
		})
	}
}

func TestRandomCodeGeneratorGenerate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		length   int
		alphabet string
	}{
		{
			name:     "default alphabet",
			length:   10,
			alphabet: testCodeAlphabet,
		},
		{
			name:     "small alphabet",
			length:   20,
			alphabet: "abc123_",
		},
		{
			name:     "single symbol alphabet",
			length:   10,
			alphabet: "a",
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			g, err := NewRandomCodeGenerator(tc.length, tc.alphabet)
			require.NoError(t, err)

			code, err := g.Generate()
			require.NoError(t, err)

			require.Len(t, code, tc.length)

			for _, symbol := range code {
				assert.Truef(t, containsRune(tc.alphabet, symbol), "generated code %q contains unexpected symbol %q", code, symbol)
			}
		})
	}
}

func TestRandomCodeGeneratorGenerateSeveralCodes(t *testing.T) {
	t.Parallel()
	g, err := NewRandomCodeGenerator(testCodeLength, testCodeAlphabet)
	require.NoError(t, err)
	for i := 0; i < 100; i++ {
		code, err := g.Generate()
		require.NoError(t, err)
		require.Len(t, code, testCodeLength)
		for _, symbol := range code {
			assert.Truef(t, containsRune(testCodeAlphabet, symbol), "generated code %q contains unexpected symbol %q", code, symbol)
		}
	}
}

func containsRune(value string, target rune) bool {
	for _, symbol := range value {
		if symbol == target {
			return true
		}
	}

	return false
}
