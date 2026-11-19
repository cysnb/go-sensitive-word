package filter

import (
	"sort"
	"sync"
	"testing"
	"time"
)

func sortedCopy(in []string) []string {
	out := append([]string{}, in...)
	sort.Strings(out)
	return out
}

func sameSet(a, b []string) bool {
	sa := sortedCopy(a)
	sb := sortedCopy(b)
	if len(sa) != len(sb) {
		return false
	}
	for i := range sa {
		if sa[i] != sb[i] {
			return false
		}
	}
	return true
}

func newFilters(words ...string) (Filter, Filter) {
	dfa := NewDfaModel()
	dfa.AddWords(words...)
	ac := NewAcModel()
	ac.AddWords(words...)
	return dfa, ac
}

func TestDetection(t *testing.T) {
	dfa, ac := newFilters("毒品", "台湾国", "fuck")
	defer func() { _ = dfa.(*DfaModel).Close() }()
	defer func() { _ = ac.(*AcModel).Close() }()

	text := "成小王对毒品销售说，台湾国的人 fuck 了"
	for name, f := range map[string]Filter{"dfa": dfa, "ac": ac} {
		if !f.IsSensitive(text) {
			t.Fatalf("%s: IsSensitive = false", name)
		}
		all := f.FindAll(text)
		for _, want := range []string{"毒品", "台湾国", "fuck"} {
			found := false
			for _, w := range all {
				if w == want {
					found = true
				}
			}
			if !found {
				t.Fatalf("%s: FindAll = %v, missing %s", name, all, want)
			}
		}
		count := f.FindAllCount("毒品毒品")
		if count["毒品"] != 2 {
			t.Fatalf("%s: count = %v", name, count)
		}
	}
}

func TestReplaceRemove(t *testing.T) {
	dfa, ac := newFilters("毒品", "台湾国")
	text := "毒品和台湾国"
	for name, f := range map[string]Filter{"dfa": dfa, "ac": ac} {
		rep := f.Replace(text, '*')
		if rep != "**和***" {
			t.Fatalf("%s: Replace = %q", name, rep)
		}
		rem := f.Remove(text)
		if rem != "和" {
			t.Fatalf("%s: Remove = %q", name, rem)
		}
	}
}

func TestEquivalence(t *testing.T) {
	words := []string{"毒品", "台湾国", "fuck", "sb", "政治"}
	dfa, ac := newFilters(words...)
	texts := []string{
		"",
		"没有敏感词",
		"毒品和台湾国",
		"FuCk sb 政治",
		"毒品毒品毒品",
		"台湾国台湾国",
	}
	for _, text := range texts {
		if dfa.IsSensitive(text) != ac.IsSensitive(text) {
			t.Fatalf("IsSensitive mismatch on %q", text)
		}
		if !sameSet(dfa.FindAll(text), ac.FindAll(text)) {
			t.Fatalf("FindAll mismatch on %q: dfa=%v ac=%v", text, dfa.FindAll(text), ac.FindAll(text))
		}
		dc := dfa.FindAllCount(text)
		acc := ac.FindAllCount(text)
		if len(dc) != len(acc) {
			t.Fatalf("FindAllCount mismatch on %q: dfa=%v ac=%v", text, dc, acc)
		}
		for k, v := range dc {
			if acc[k] != v {
				t.Fatalf("FindAllCount mismatch on %q: dfa=%v ac=%v", text, dc, acc)
			}
		}
		if dfa.Replace(text, '*') != ac.Replace(text, '*') {
			t.Fatalf("Replace mismatch on %q: dfa=%q ac=%q", text, dfa.Replace(text, '*'), ac.Replace(text, '*'))
		}
		if dfa.Remove(text) != ac.Remove(text) {
			t.Fatalf("Remove mismatch on %q: dfa=%q ac=%q", text, dfa.Remove(text), ac.Remove(text))
		}
	}
}

func TestRanges(t *testing.T) {
	dfa, ac := newFilters("毒品")
	for name, f := range map[string]RangedFilter{"dfa": dfa.(RangedFilter), "ac": ac.(RangedFilter)} {
		ranges := f.FindAllRanges("x毒品y")
		if len(ranges) != 1 || ranges[0].Start != 1 || ranges[0].End != 2 {
			t.Fatalf("%s: ranges = %v", name, ranges)
		}
	}
}

func TestConcurrentAccess(t *testing.T) {
	models := map[string]Filter{
		"dfa": NewDfaModel(),
		"ac":  NewAcModel(),
	}
	for name, f := range models {
		t.Run(name, func(t *testing.T) {
			var wg sync.WaitGroup
			stop := make(chan struct{})
			for i := 0; i < 8; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for {
						select {
						case <-stop:
							return
						default:
							_ = f.FindAll("毒品和台湾国")
						}
					}
				}()
			}
			for i := 0; i < 50; i++ {
				switch v := f.(type) {
				case *DfaModel:
					v.AddWord("词")
					v.DelWord("词")
				case *AcModel:
					v.AddWord("词")
					v.DelWord("词")
				}
			}
			close(stop)
			wg.Wait()
		})
	}
}

func TestListenAndApply(t *testing.T) {
	for name, f := range map[string]Filter{
		"dfa": NewDfaModel(),
		"ac":  NewAcModel(),
	} {
		t.Run(name, func(t *testing.T) {
			add := make(chan string, 1)
			del := make(chan string, 1)
			appliedCh := make(chan struct{}, 4)
			applied := func() { appliedCh <- struct{}{} }
			switch v := f.(type) {
			case *DfaModel:
				v.Listen(add, del, applied)
			case *AcModel:
				v.Listen(add, del, applied)
			}
			add <- "毒品"
			select {
			case <-appliedCh:
			case <-time.After(2 * time.Second):
				t.Fatal("timeout waiting applied")
			}
			if !f.IsSensitive("毒品") {
				t.Fatal("word not applied")
			}
			switch v := f.(type) {
			case *DfaModel:
				_ = v.Close()
				_ = v.Close()
			case *AcModel:
				_ = v.Close()
				_ = v.Close()
			}
		})
	}
}
