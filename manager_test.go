package go_sensitive_word

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cysnb/go-sensitive-word/store"
)

func mustManager(t *testing.T, filterType uint32, opts ...Option) *Manager {
	t.Helper()
	m, err := NewFilter(
		StoreOption{Type: StoreMemory},
		FilterOption{Type: filterType},
		opts...,
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = m.Close() })
	return m
}

func wait(t *testing.T, m *Manager) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := m.WaitForSync(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestNewFilter(t *testing.T) {
	for _, typ := range []uint32{FilterDfa, FilterAc} {
		m := mustManager(t, typ)
		if m.Store == nil || m.Filter == nil {
			t.Fatalf("type %d: manager not initialized", typ)
		}
	}
	if _, err := NewFilter(StoreOption{Type: 99}, FilterOption{Type: FilterDfa}); err == nil {
		t.Fatal("expect invalid store error")
	}
	if _, err := NewFilter(StoreOption{Type: StoreMemory}, FilterOption{Type: 99}); err == nil {
		t.Fatal("expect invalid filter error")
	}
}

func TestDynamicAddDelete(t *testing.T) {
	for _, typ := range []uint32{FilterDfa, FilterAc} {
		m := mustManager(t, typ)
		if err := m.Store.AddWord("毒品"); err != nil {
			t.Fatal(err)
		}
		wait(t, m)
		if !m.IsSensitive("这里有毒品") {
			t.Fatalf("type %d: should detect word", typ)
		}
		if err := m.Store.DelWord("毒品"); err != nil {
			t.Fatal(err)
		}
		wait(t, m)
		if m.IsSensitive("这里有毒品") {
			t.Fatalf("type %d: should not detect deleted word", typ)
		}
	}
}

func TestNormalizationConsistency(t *testing.T) {
	m := mustManager(t, FilterDfa)
	if err := m.Store.AddWord("Ｆｕｃｋ"); err != nil {
		t.Fatal(err)
	}
	wait(t, m)
	if !m.IsSensitive("fuck") {
		t.Fatal("word should be normalized to half-width lower case")
	}
	if got := m.FindOne("说fuck吧"); got != "fuck" {
		t.Fatalf("FindOne = %q", got)
	}
	all := m.FindAll("ＦＵＣＫ")
	if len(all) != 1 || all[0] != "ＦＵＣＫ" {
		t.Fatalf("FindAll = %v", all)
	}
	if got := m.Replace("ＦＵＣＫ", '*'); got != "****" {
		t.Fatalf("Replace = %q", got)
	}
}

func TestStrictNormalizerZeroWidth(t *testing.T) {
	m := mustManager(t, FilterDfa, WithStrictNormalizer())
	if err := m.Store.AddWord("fuck"); err != nil {
		t.Fatal(err)
	}
	wait(t, m)
	text := "f\u200Buck"
	if !m.IsSensitive(text) {
		t.Fatal("strict normalizer should detect zero-width bypass")
	}
	if got := m.FindAll(text); len(got) != 1 || got[0] != text {
		t.Fatalf("FindAll = %v", got)
	}
}

func TestCustomNormalizer(t *testing.T) {
	cfg := DefaultNormalizer()
	cfg.HomoglyphMap = map[rune]rune{'а': 'a'}
	m := mustManager(t, FilterDfa, WithNormalizer(cfg))
	if err := m.Store.AddWord("abc"); err != nil {
		t.Fatal(err)
	}
	wait(t, m)
	if !m.IsSensitive("аbc") {
		t.Fatal("custom homoglyph map should apply")
	}
}

func TestSourceTracking(t *testing.T) {
	m := mustManager(t, FilterDfa)
	if err := m.Store.AddWordsWithSource([]string{"违禁词A"}, "custom"); err != nil {
		t.Fatal(err)
	}
	wait(t, m)

	results := m.FindAllWithSource("包含违禁词A")
	if len(results) != 1 || results[0].Word != "违禁词A" {
		t.Fatalf("results = %v", results)
	}
	if len(results[0].Source) != 1 || results[0].Source[0] != "custom" {
		t.Fatalf("source = %v", results[0].Source)
	}

	counts := m.FindAllCountWithSource("违禁词A违禁词A")
	got, ok := counts["违禁词A"]
	if !ok || got.Count != 2 {
		t.Fatalf("counts = %v", counts)
	}
}

func TestExportToFile(t *testing.T) {
	m := mustManager(t, FilterDfa)
	if err := m.Store.AddWords([]string{"a", "b"}); err != nil {
		t.Fatal(err)
	}
	wait(t, m)
	path := filepath.Join(t.TempDir(), "export.txt")
	if err := m.ExportToFile(path); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(data), "\n") != 2 {
		t.Fatalf("file = %q", data)
	}
}

func TestMergeAndRefresh(t *testing.T) {
	m1 := mustManager(t, FilterDfa)
	m2 := mustManager(t, FilterDfa)
	if err := m1.Store.AddWord("a"); err != nil {
		t.Fatal(err)
	}
	if err := m2.Store.AddWord("b"); err != nil {
		t.Fatal(err)
	}
	if err := m1.MergeFromManager(m2); err != nil {
		t.Fatal(err)
	}
	wait(t, m1)
	if got := m1.GetStats().TotalWords; got != 2 {
		t.Fatalf("merged total = %d", got)
	}

	path := filepath.Join(t.TempDir(), "dict.txt")
	if err := os.WriteFile(path, []byte("x\ny\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := m1.RefreshFromPath(path, true); err != nil {
		t.Fatal(err)
	}
	wait(t, m1)
	if got := m1.GetStats().TotalWords; got != 2 {
		t.Fatalf("refreshed total = %d", got)
	}
	if !m1.IsSensitive("x") || !m1.IsSensitive("y") {
		t.Fatal("refreshed words should be detected")
	}
}

func TestLifecycle(t *testing.T) {
	m := mustManager(t, FilterAc)
	if err := m.Store.AddWord("a"); err != nil {
		t.Fatal(err)
	}
	wait(t, m)
	if err := m.Close(); err != nil {
		t.Fatal(err)
	}
	if err := m.Close(); err != nil {
		t.Fatal(err)
	}
	if err := m.Store.AddWord("b"); !errors.Is(err, store.ErrStoreClosed) {
		t.Fatalf("err = %v, want ErrStoreClosed", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := m.WaitForSync(ctx); !errors.Is(err, store.ErrStoreClosed) {
		t.Fatalf("WaitForSync err = %v", err)
	}
}

func TestShutdown(t *testing.T) {
	m, err := NewFilter(StoreOption{Type: StoreMemory}, FilterOption{Type: FilterAc})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := m.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
}
