package go_sensitive_word

import (
	"strings"
	"testing"
)

func TestBuiltinDictionaries(t *testing.T) {
	dicts := map[string]string{
		"DictCovid19":          DictCovid19,
		"DictGFWAdditional":    DictGFWAdditional,
		"DictOther":            DictOther,
		"DictReactionary":      DictReactionary,
		"DictAdvertisement":    DictAdvertisement,
		"DictPolitical":        DictPolitical,
		"DictViolence":         DictViolence,
		"DictPeopleLife":       DictPeopleLife,
		"DictGunExplosion":     DictGunExplosion,
		"DictNeteaseFE":        DictNeteaseFE,
		"DictSexual":           DictSexual,
		"DictPornography":      DictPornography,
		"DictAdditional":       DictAdditional,
		"DictCorruption":       DictCorruption,
		"DictTemporaryTencent": DictTemporaryTencent,
		"DictIllegalURL":       DictIllegalURL,
	}
	if len(dicts) != 16 {
		t.Fatalf("expect 16 dictionaries, got %d", len(dicts))
	}
	union := make(map[string]struct{})
	for name, content := range dicts {
		if content == "" {
			t.Fatalf("%s is empty", name)
		}
		seen := make(map[string]struct{})
		for _, line := range strings.Split(content, "\n") {
			word := strings.TrimSpace(line)
			if word == "" {
				continue
			}
			key := strings.ToLower(word)
			if _, ok := seen[key]; ok {
				t.Fatalf("%s contains duplicate word %q", name, word)
			}
			seen[key] = struct{}{}
			union[key] = struct{}{}
		}
	}
	if len(union) != 51103 {
		t.Fatalf("expect 51103 unique words, got %d", len(union))
	}
}

func TestPornographyDictionarySize(t *testing.T) {
	seen := make(map[string]struct{})
	lines := 0
	for _, line := range strings.Split(DictPornography, "\n") {
		word := strings.TrimSpace(line)
		if word == "" {
			continue
		}
		lines++
		seen[strings.ToLower(word)] = struct{}{}
	}
	if lines != len(seen) {
		t.Fatalf("porn dictionary has duplicates: %d lines, %d unique", lines, len(seen))
	}
	if len(seen) != 587 {
		t.Fatalf("expect 587 porn words, got %d", len(seen))
	}
}

func TestFilterAcAlias(t *testing.T) {
	if FilterAc != FilterAC {
		t.Fatalf("FilterAc and FilterAC must be equal")
	}
}
