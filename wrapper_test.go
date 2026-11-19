package go_sensitive_word

import (
	"strings"
	"testing"

	"github.com/cysnb/go-sensitive-word/filter"
)

type fakeInner struct {
	word string
}

func (f *fakeInner) FindOne(text string) string {
	if strings.Contains(text, f.word) {
		return f.word
	}
	return ""
}

func (f *fakeInner) FindAll(text string) []string {
	if strings.Contains(text, f.word) {
		return []string{f.word}
	}
	return nil
}

func (f *fakeInner) FindAllCount(text string) map[string]int {
	return map[string]int{f.word: strings.Count(text, f.word)}
}

func (f *fakeInner) IsSensitive(text string) bool {
	return strings.Contains(text, f.word)
}

func (f *fakeInner) Replace(text string, repl rune) string {
	return strings.ReplaceAll(text, f.word, string(repl))
}

func (f *fakeInner) Remove(text string) string {
	return strings.ReplaceAll(text, f.word, "")
}

func (f *fakeInner) FindAllRanges(text string) []filter.Range {
	var res []filter.Range
	rText := []rune(text)
	rWord := []rune(f.word)
	for i := 0; i+len(rWord) <= len(rText); i++ {
		ok := true
		for j := range rWord {
			if rText[i+j] != rWord[j] {
				ok = false
				break
			}
		}
		if ok {
			res = append(res, filter.Range{Start: i, End: i + len(rWord) - 1})
		}
	}
	return res
}

func TestWrapperReturnsOriginalFragments(t *testing.T) {
	nf := newNormalizedFilter(&fakeInner{word: "fuck"}, DefaultNormalizer())
	if got := nf.FindOne("说 ＦＵＣＫ 吧"); got != "ＦＵＣＫ" {
		t.Fatalf("FindOne = %q, want ＦＵＣＫ", got)
	}
	all := nf.FindAll("ＦＵＣＫ and Ｆｕｃｋ")
	if len(all) == 0 {
		t.Fatalf("FindAll should not be empty")
	}
	for _, w := range all {
		if strings.ToLower(w) != "ｆｕｃｋ" {
			t.Fatalf("unexpected fragment %q", w)
		}
	}
}

func TestWrapperReplaceAndRemove(t *testing.T) {
	nf := newNormalizedFilter(&fakeInner{word: "fuck"}, DefaultNormalizer())
	if got := nf.Replace("ＦＵＣＫ!", '*'); got != "****!" {
		t.Fatalf("Replace = %q, want ****!", got)
	}
	if got := nf.Remove("ＦＵＣＫ!"); got != "!" {
		t.Fatalf("Remove = %q, want !", got)
	}
}

func TestWrapperStrictZeroWidth(t *testing.T) {
	nf := newNormalizedFilter(&fakeInner{word: "fuck"}, StrictNormalizer())
	text := "f\u200Buck"
	if got := nf.FindOne(text); got != text {
		t.Fatalf("FindOne = %q, want %q", got, text)
	}
	if got := nf.Replace(text, '*'); got != "*****" {
		t.Fatalf("Replace = %q, want *****", got)
	}
	if got := nf.Remove(text); got != "" {
		t.Fatalf("Remove = %q, want empty", got)
	}
}
