package generator

import (
	"errors"
	"testing"

	xerrors "github.com/Saperart/linklite-url-shortener/internal/errors"
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
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				if g != nil {
					t.Fatalf("expected nil generator, got %#v", g)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if g == nil {
				t.Fatal("expected generator, got nil")
			}
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
			if err != nil {
				t.Fatalf("unexpected constructor error: %v", err)
			}

			code, err := g.Generate()
			if err != nil {
				t.Fatalf("unexpected generate error: %v", err)
			}

			if len(code) != tc.length {
				t.Fatalf("expected code length %d, got %d", tc.length, len(code))
			}

			for _, symbol := range code {
				if !containsRune(tc.alphabet, symbol) {
					t.Fatalf("generated code %q contains unexpected symbol %q", code, symbol)
				}
			}
		})
	}
}

func TestRandomCodeGeneratorGenerateSeveralCodes(t *testing.T) {
	t.Parallel()
	g, err := NewRandomCodeGenerator(testCodeLength, testCodeAlphabet)
	if err != nil {
		t.Fatalf("unexpected constructor error: %v", err)
	}
	for i := 0; i < 100; i++ {
		code, err := g.Generate()
		if err != nil {
			t.Fatalf("unexpected generate error: %v", err)
		}
		if len(code) != testCodeLength {
			t.Fatalf("expected code length %d, got %d", testCodeLength, len(code))
		}
		for _, symbol := range code {
			if !containsRune(testCodeAlphabet, symbol) {
				t.Fatalf("generated code %q contains unexpected symbol %q", code, symbol)
			}
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
