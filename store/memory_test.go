package store

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

type consumer struct {
	stop chan struct{}
	done chan struct{}
}

func startConsumer(m *MemoryModel) *consumer {
	c := &consumer{stop: make(chan struct{}), done: make(chan struct{})}
	go func() {
		defer close(c.done)
		for {
			select {
			case <-m.GetAddChan():
				m.Done()
			case <-m.GetDelChan():
				m.Done()
			case <-c.stop:
				return
			case <-m.closedCh:
				return
			}
		}
	}()
	return c
}

func (c *consumer) Close() {
	close(c.stop)
	<-c.done
}

func TestAddDeleteRead(t *testing.T) {
	m := NewMemoryModel()
	c := startConsumer(m)
	defer func() { c.Close(); _ = m.Close() }()

	if err := m.AddWord("ABC", "def"); err != nil {
		t.Fatal(err)
	}
	if err := m.WaitForSync(context.Background()); err != nil {
		t.Fatal(err)
	}
	words := m.ReadString()
	if len(words) != 2 || words[0] != "abc" || words[1] != "def" {
		t.Fatalf("words = %v", words)
	}
	if got := m.GetStats().TotalWords; got != 2 {
		t.Fatalf("total = %d, want 2", got)
	}
	if err := m.DelWord("ABC"); err != nil {
		t.Fatal(err)
	}
	if err := m.WaitForSync(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := m.ReadString(); len(got) != 1 || got[0] != "def" {
		t.Fatalf("after delete words = %v", got)
	}
}

func TestNormalizerInjection(t *testing.T) {
	m := NewMemoryModel(WithWordNormalizer(func(w string) string {
		return strings.ToUpper(strings.TrimSpace(w))
	}))
	c := startConsumer(m)
	defer func() { c.Close(); _ = m.Close() }()

	if err := m.AddWord("abc"); err != nil {
		t.Fatal(err)
	}
	if got := m.ReadString(); len(got) != 1 || got[0] != "ABC" {
		t.Fatalf("got %v", got)
	}
}

func TestLoaders(t *testing.T) {
	t.Run("embed", func(t *testing.T) {
		m := NewMemoryModel()
		c := startConsumer(m)
		defer func() { c.Close(); _ = m.Close() }()
		if err := m.LoadDictEmbed("abc\ndef\n\n"); err != nil {
			t.Fatal(err)
		}
		if got := m.GetStats().TotalWords; got != 2 {
			t.Fatalf("total = %d", got)
		}
	})

	t.Run("path", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "dict.txt")
		if err := os.WriteFile(path, []byte("abc\ndef\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		m := NewMemoryModel()
		c := startConsumer(m)
		defer func() { c.Close(); _ = m.Close() }()
		if err := m.LoadDictPath(path); err != nil {
			t.Fatal(err)
		}
		if got := m.GetStats().TotalWords; got != 2 {
			t.Fatalf("total = %d", got)
		}
	})

	t.Run("missing path", func(t *testing.T) {
		m := NewMemoryModel()
		defer func() { _ = m.Close() }()
		if err := m.LoadDictPath(filepath.Join(t.TempDir(), "missing.txt")); err == nil {
			t.Fatal("expect error for missing file")
		}
	})

	t.Run("reader", func(t *testing.T) {
		m := NewMemoryModel()
		c := startConsumer(m)
		defer func() { c.Close(); _ = m.Close() }()
		if err := m.LoadDict(strings.NewReader("abc\n")); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("callback", func(t *testing.T) {
		m := NewMemoryModel()
		c := startConsumer(m)
		defer func() { c.Close(); _ = m.Close() }()
		err := m.LoadDictCallback(func() ([]string, error) {
			return []string{"abc"}, nil
		}, "db")
		if err != nil {
			t.Fatal(err)
		}
		if got := m.GetStats().Source; len(got) != 1 || got[0] != "callback://db" {
			t.Fatalf("source = %v", got)
		}
	})

	t.Run("callback error", func(t *testing.T) {
		m := NewMemoryModel()
		defer func() { _ = m.Close() }()
		want := errors.New("boom")
		err := m.LoadDictCallback(func() ([]string, error) { return nil, want }, "db")
		if !errors.Is(err, want) {
			t.Fatalf("err = %v, want boom", err)
		}
	})

	t.Run("http", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = io.WriteString(w, "abc\ndef\n")
		}))
		defer srv.Close()
		m := NewMemoryModel()
		c := startConsumer(m)
		defer func() { c.Close(); _ = m.Close() }()
		if err := m.LoadDictHttp(srv.URL); err != nil {
			t.Fatal(err)
		}
		if got := m.GetStats().TotalWords; got != 2 {
			t.Fatalf("total = %d", got)
		}
	})

	t.Run("http non 200", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()
		m := NewMemoryModel()
		defer func() { _ = m.Close() }()
		if err := m.LoadDictHttp(srv.URL); err == nil {
			t.Fatal("expect error for non 200")
		}
	})
}

func TestBatchExportStats(t *testing.T) {
	m := NewMemoryModel()
	c := startConsumer(m)
	defer func() { c.Close(); _ = m.Close() }()

	if err := m.AddWords([]string{"a", "b", "c"}); err != nil {
		t.Fatal(err)
	}
	if err := m.ReplaceWords([]string{"b"}, []string{"d"}); err != nil {
		t.Fatal(err)
	}
	words := m.ReadString()
	if len(words) != 3 || words[0] != "a" || words[1] != "c" || words[2] != "d" {
		t.Fatalf("words = %v", words)
	}

	out, err := m.ExportToString()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(out, "\n") != 3 {
		t.Fatalf("export = %q", out)
	}

	var buf strings.Builder
	if err := m.ExportToWriter(&buf); err != nil {
		t.Fatal(err)
	}
	if buf.String() != out {
		t.Fatalf("writer export mismatch")
	}

	stats := m.GetStats()
	if stats.TotalWords != 3 {
		t.Fatalf("total = %d", stats.TotalWords)
	}
	stats.Source = append(stats.Source, "x")
	if len(m.GetStats().Source) != 0 {
		t.Fatal("stats must return a copy")
	}
}

func TestClearAndMerge(t *testing.T) {
	a := NewMemoryModel()
	ca := startConsumer(a)
	defer func() { ca.Close(); _ = a.Close() }()
	b := NewMemoryModel()
	cb := startConsumer(b)
	defer func() { cb.Close(); _ = b.Close() }()

	if err := a.AddWords([]string{"a", "b"}); err != nil {
		t.Fatal(err)
	}
	if err := b.AddWords([]string{"b", "c"}); err != nil {
		t.Fatal(err)
	}
	if err := a.Merge(b); err != nil {
		t.Fatal(err)
	}
	if got := a.GetStats().TotalWords; got != 3 {
		t.Fatalf("merged total = %d, want 3", got)
	}
	if err := a.Clear(); err != nil {
		t.Fatal(err)
	}
	if got := a.GetStats().TotalWords; got != 0 {
		t.Fatalf("after clear total = %d", got)
	}
}

func TestSourcesAndCopy(t *testing.T) {
	m := NewMemoryModel()
	c := startConsumer(m)
	defer func() { c.Close(); _ = m.Close() }()

	if err := m.AddWordsWithSource([]string{"Kill"}, "custom"); err != nil {
		t.Fatal(err)
	}
	got := m.GetWordSources("kill")
	if len(got) != 1 || got[0] != "custom" {
		t.Fatalf("sources = %v", got)
	}
	got[0] = "changed"
	if again := m.GetWordSources("kill"); again[0] != "custom" {
		t.Fatal("sources must return a copy")
	}
	all := m.GetAllWordSources()
	if len(all) != 1 {
		t.Fatalf("all sources = %v", all)
	}
}

func TestReadChan(t *testing.T) {
	m := NewMemoryModel()
	c := startConsumer(m)
	defer func() { c.Close(); _ = m.Close() }()
	if err := m.AddWords([]string{"a", "b"}); err != nil {
		t.Fatal(err)
	}
	var got []string
	for word := range m.ReadChan() {
		got = append(got, word)
	}
	if len(got) != 2 {
		t.Fatalf("got %v", got)
	}
}

func TestWaitForSync(t *testing.T) {
	m := NewMemoryModel()
	c := startConsumer(m)
	defer func() { c.Close(); _ = m.Close() }()
	if err := m.AddWord("x"); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := m.WaitForSync(ctx); err != nil {
		t.Fatal(err)
	}
	if m.pending.Load() != 0 {
		t.Fatalf("pending = %d", m.pending.Load())
	}
}

func TestLifecycle(t *testing.T) {
	before := runtime.NumGoroutine()
	m := NewMemoryModel()
	c := startConsumer(m)
	if err := m.AddWord("a"); err != nil {
		t.Fatal(err)
	}
	if err := m.Close(); err != nil {
		t.Fatal(err)
	}
	if err := m.Close(); err != nil {
		t.Fatal(err)
	}
	if err := m.AddWord("b"); !errors.Is(err, ErrStoreClosed) {
		t.Fatalf("err = %v, want ErrStoreClosed", err)
	}
	c.Close()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if runtime.NumGoroutine() <= before+2 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("goroutines before=%d after=%d", before, runtime.NumGoroutine())
}
