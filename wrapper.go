package go_sensitive_word

import (
	"github.com/cysnb/go-sensitive-word/filter"
)

type normalizedFilter struct {
	cfg   NormalizerConfig
	inner filter.Filter
}

func newNormalizedFilter(inner filter.Filter, cfg NormalizerConfig) *normalizedFilter {
	return &normalizedFilter{cfg: cfg, inner: inner}
}

func findFirst(haystack string, needle string) (int, int, bool) {
	rText := []rune(haystack)
	rNeedle := []rune(needle)
	if len(rNeedle) == 0 {
		return 0, 0, false
	}
	for i := 0; i+len(rNeedle) <= len(rText); i++ {
		ok := true
		for j := 0; j < len(rNeedle); j++ {
			if rText[i+j] != rNeedle[j] {
				ok = false
				break
			}
		}
		if ok {
			return i, i + len(rNeedle) - 1, true
		}
	}
	return 0, 0, false
}

func (nf *normalizedFilter) FindOne(text string) string {
	normText, idxMap := NormalizeTextWithMap(text, nf.cfg)
	hit := nf.inner.FindOne(normText)
	if hit == "" {
		return ""
	}
	start, end, ok := findFirst(normText, hit)
	if !ok {
		return ""
	}
	origRunes := []rune(text)
	lo, hi := idxMap[start], idxMap[end]
	if lo >= len(origRunes) || hi >= len(origRunes) {
		return ""
	}
	return string(origRunes[lo : hi+1])
}

func (nf *normalizedFilter) FindAll(text string) []string {
	normText, idxMap := NormalizeTextWithMap(text, nf.cfg)

	if rf, ok := nf.inner.(filter.RangedFilter); ok {
		ranges := rf.FindAllRanges(normText)
		if len(ranges) == 0 {
			return []string{}
		}
		origRunes := []rune(text)
		res := make([]string, 0, len(ranges))
		for _, r := range ranges {
			lo, hi := idxMap[r.Start], idxMap[r.End]
			if lo < len(origRunes) && hi < len(origRunes) {
				res = append(res, string(origRunes[lo:hi+1]))
			}
		}
		return res
	}

	hits := nf.inner.FindAll(normText)
	if len(hits) == 0 {
		return hits
	}
	origRunes := []rune(text)
	res := make([]string, 0, len(hits))
	for _, h := range hits {
		start, end, ok := findFirst(normText, h)
		if !ok {
			continue
		}
		res = append(res, string(origRunes[idxMap[start]:idxMap[end]+1]))
	}
	return res
}

func (nf *normalizedFilter) FindAllCount(text string) map[string]int {
	normText, idxMap := NormalizeTextWithMap(text, nf.cfg)
	hits := nf.inner.FindAll(normText)
	res := make(map[string]int, len(hits))
	for _, h := range hits {
		start, end, ok := findFirst(normText, h)
		if !ok {
			continue
		}
		count := 0
		rNorm := []rune(normText)
		rHit := []rune(h)
		for i := 0; i+len(rHit) <= len(rNorm); i++ {
			match := true
			for j := 0; j < len(rHit); j++ {
				if rNorm[i+j] != rHit[j] {
					match = false
					break
				}
			}
			if match {
				count++
				i += len(rHit) - 1
			}
		}
		if count == 0 {
			continue
		}
		origRunes := []rune(text)
		res[string(origRunes[idxMap[start]:idxMap[end]+1])] = count
	}
	return res
}

func (nf *normalizedFilter) IsSensitive(text string) bool {
	return nf.FindOne(text) != ""
}

func (nf *normalizedFilter) Replace(text string, repl rune) string {
	normText, idxMap := NormalizeTextWithMap(text, nf.cfg)
	rOrig := []rune(text)
	marked := make([]bool, len(rOrig))

	mark := func(start, end int) {
		lo, hi := idxMap[start], idxMap[end]
		for k := lo; k <= hi && k < len(marked); k++ {
			marked[k] = true
		}
	}

	if rf, ok := nf.inner.(filter.RangedFilter); ok {
		for _, r := range rf.FindAllRanges(normText) {
			mark(r.Start, r.End)
		}
	} else {
		applyAllOccurrences(normText, nf.inner.FindAll(normText), mark)
	}

	for i := range marked {
		if marked[i] {
			rOrig[i] = repl
		}
	}
	return string(rOrig)
}

func (nf *normalizedFilter) Remove(text string) string {
	normText, idxMap := NormalizeTextWithMap(text, nf.cfg)
	rOrig := []rune(text)
	del := make([]bool, len(rOrig))

	mark := func(start, end int) {
		lo, hi := idxMap[start], idxMap[end]
		for k := lo; k <= hi && k < len(del); k++ {
			del[k] = true
		}
	}

	if rf, ok := nf.inner.(filter.RangedFilter); ok {
		for _, r := range rf.FindAllRanges(normText) {
			mark(r.Start, r.End)
		}
	} else {
		applyAllOccurrences(normText, nf.inner.FindAll(normText), mark)
	}

	out := make([]rune, 0, len(rOrig))
	for i, r := range rOrig {
		if !del[i] {
			out = append(out, r)
		}
	}
	return string(out)
}

func applyAllOccurrences(normText string, hits []string, fn func(start, end int)) {
	rNorm := []rune(normText)
	for _, h := range hits {
		rHit := []rune(h)
		for i := 0; i+len(rHit) <= len(rNorm); i++ {
			match := true
			for j := 0; j < len(rHit); j++ {
				if rNorm[i+j] != rHit[j] {
					match = false
					break
				}
			}
			if match {
				fn(i, i+len(rHit)-1)
				i += len(rHit) - 1
			}
		}
	}
}
