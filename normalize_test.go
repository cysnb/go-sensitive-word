package go_sensitive_word

import "testing"

func TestDefaultNormalizer(t *testing.T) {
	cfg := DefaultNormalizer()
	if got := NormalizeWord("ＦｕＣｋ", cfg); got != "fuck" {
		t.Fatalf("got %q, want fuck", got)
	}
}

func TestStrictNormalizer(t *testing.T) {
	cfg := StrictNormalizer()
	if got := NormalizeWord("漢字", cfg); got != "汉字" {
		t.Fatalf("simp trad normalize got %q, want 汉字", got)
	}
	if got := NormalizeWord("a\u200Bb", cfg); got != "ab" {
		t.Fatalf("zero width normalize got %q, want ab", got)
	}
	if got := NormalizeWord("①", cfg); got != "1" {
		t.Fatalf("digit normalize got %q, want 1", got)
	}
}

func TestCustomHomoglyphConfig(t *testing.T) {
	cfg := NormalizerConfig{
		HomoglyphMap: map[rune]rune{'а': 'a'},
	}
	if got := NormalizeWord("а", cfg); got != "a" {
		t.Fatalf("got %q, want a", got)
	}
}
