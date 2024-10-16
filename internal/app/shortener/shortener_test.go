package shortener

import (
	"strings"
	"testing"
	"unicode"
)

func TestShortener(t *testing.T) {
	tests := []struct {
		name      string
		inputURL  string
		wantLen   int
		wantErr   bool
		checkFunc func(t *testing.T, shortURL string)
	}{
		{
			name:     "valid short URL length",
			inputURL: "https://example.com",
			wantLen:  urlLength,
			wantErr:  false,
			checkFunc: func(t *testing.T, shortURL string) {
				if len(shortURL) != urlLength {
					t.Errorf("expected shortened URL length of %d, got %d", urlLength, len(shortURL))
				}
			},
		},
		{
			name:     "valid charset characters",
			inputURL: "https://example.com",
			wantErr:  false,
			checkFunc: func(t *testing.T, shortURL string) {
				for _, ch := range shortURL {
					if !strings.ContainsRune(charset, ch) {
						t.Errorf("unexpected character in shortened URL: %c", ch)
					}
				}
			},
		},
		{
			name:     "only alphanumeric characters",
			inputURL: "https://example.com",
			wantErr:  false,
			checkFunc: func(t *testing.T, shortURL string) {
				for _, r := range shortURL {
					if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
						t.Errorf("unexpected non-alphanumeric character in shortened URL: %c", r)
					}
				}
			},
		},
		{
			name:     "multiple unique short URLs",
			inputURL: "https://example.com",
			wantErr:  false,
			checkFunc: func(t *testing.T, shortURL string) {
				urls := make(map[string]bool)
				for i := 0; i < 100; i++ {
					shortURL, err := NewShortener().Shorten("https://example.com")
					if err != nil {
						t.Fatalf("unexpected error: %v", err)
					}
					if urls[shortURL] {
						t.Errorf("duplicate shortened URL generated: %s", shortURL)
					}
					urls[shortURL] = true
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shortener := NewShortener()
			shortURL, err := shortener.Shorten(tt.inputURL)

			if (err != nil) != tt.wantErr {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, shortURL)
			}
		})
	}
}
