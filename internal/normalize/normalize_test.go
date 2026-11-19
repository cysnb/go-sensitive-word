package normalize

import "testing"

func TestNormalizeRuneOptions(t *testing.T) {
	cases := []struct {
		name string
		in   string
		cfg  Config
		want string
	}{
		{"ignore case", "FuCk", Config{IgnoreCase: true}, "fuck"},
		{"half width", "ｆｕｃｋ", Config{ToHalfWidth: true}, "fuck"},
		{"ignore repeat", "fuuuck", Config{IgnoreRepeat: true}, "fuck"},
		{"digit type", "①②③", Config{IgnoreDigitType: true}, "123"},
		{"simp trad", "漢字", Config{IgnoreSimpTrad: true}, "汉字"},
		{"english style", "Ⓕ", Config{IgnoreEnglishStyle: true, IgnoreCase: true}, "f"},
		{"zero width", "a\u200Bb", Config{RemoveZeroWidth: true}, "ab"},
		{"homoglyph", "а", Config{HomoglyphMap: map[rune]rune{'а': 'a'}}, "a"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := NormalizeWord(c.in, c.cfg); got != c.want {
				t.Fatalf("NormalizeWord(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestNormalizeTextWithMap(t *testing.T) {
	in := "Ａbc"
	cfg := Config{ToHalfWidth: true, IgnoreCase: true}
	got, idx := NormalizeTextWithMap(in, cfg)
	if got != "abc" {
		t.Fatalf("got %q, want abc", got)
	}
	want := []int{0, 1, 2}
	if len(idx) != len(want) {
		t.Fatalf("idx len %d, want %d", len(idx), len(want))
	}
	for i := range want {
		if idx[i] != want[i] {
			t.Fatalf("idx[%d] = %d, want %d", i, idx[i], want[i])
		}
	}
}

func TestNormalizeTextWithMapSkipsZeroWidth(t *testing.T) {
	got, idx := NormalizeTextWithMap("a\u200Bb", Config{RemoveZeroWidth: true})
	if got != "ab" {
		t.Fatalf("got %q, want ab", got)
	}
	if len(idx) != 2 || idx[0] != 0 || idx[1] != 2 {
		t.Fatalf("unexpected idx map: %v", idx)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if !cfg.IgnoreCase || !cfg.ToHalfWidth {
		t.Fatalf("default config must enable case and half width")
	}
	if cfg.IgnoreRepeat || cfg.RemoveZeroWidth {
		t.Fatalf("default config must not enable strict options")
	}
}

func TestDefaultHomoglyphMapCopy(t *testing.T) {
	a := DefaultHomoglyphMap()
	a['а'] = 'z'
	b := DefaultHomoglyphMap()
	if b['а'] != 'a' {
		t.Fatalf("DefaultHomoglyphMap must return a copy")
	}
}
