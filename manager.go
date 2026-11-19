package go_sensitive_word

import (
	"context"
	"errors"
	"os"

	"github.com/cysnb/go-sensitive-word/filter"
	"github.com/cysnb/go-sensitive-word/store"
)

type listener interface {
	Listen(addChan, delChan <-chan string, applied func())
}

type closer interface {
	Close() error
}

// Manager 整合词库存储与过滤算法。
type Manager struct {
	store.Store
	filter.Filter
	normalizer NormalizerConfig
	storeModel *store.MemoryModel
	rawFilter  filter.Filter
}

// NewFilter 初始化过滤器与词库存储。
func NewFilter(storeOption StoreOption, filterOption FilterOption, opts ...Option) (*Manager, error) {
	cfg := config{normalizer: DefaultNormalizer()}
	for _, opt := range opts {
		opt(&cfg)
	}

	var filterStore *store.MemoryModel
	switch storeOption.Type {
	case StoreMemory:
		filterStore = store.NewMemoryModel(store.WithWordNormalizer(func(word string) string {
			return NormalizeWord(word, cfg.normalizer)
		}))
	default:
		return nil, errors.New("invalid store type")
	}

	var raw filter.Filter
	switch filterOption.Type {
	case FilterDfa:
		raw = filter.NewDfaModel()
	case FilterAc:
		raw = filter.NewAcModel()
	default:
		_ = filterStore.Close()
		return nil, errors.New("invalid filter type")
	}

	if l, ok := raw.(listener); ok {
		l.Listen(filterStore.GetAddChan(), filterStore.GetDelChan(), filterStore.Done)
	}

	return &Manager{
		Store:      filterStore,
		Filter:     newNormalizedFilter(raw, cfg.normalizer),
		normalizer: cfg.normalizer,
		storeModel: filterStore,
		rawFilter:  raw,
	}, nil
}

// WaitForSync 等待词库变更应用到过滤器。
func (m *Manager) WaitForSync(ctx context.Context) error {
	if m.storeModel == nil {
		return errors.New("store is nil")
	}
	return m.storeModel.WaitForSync(ctx)
}

// Close 关闭过滤器与词库。
func (m *Manager) Close() error {
	var firstErr error
	if m.rawFilter != nil {
		if c, ok := m.rawFilter.(closer); ok {
			firstErr = c.Close()
		}
	}
	if m.Store != nil {
		if err := m.Store.Close(); firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// Shutdown 带超时关闭过滤器与词库。
func (m *Manager) Shutdown(ctx context.Context) error {
	if c, ok := m.rawFilter.(closer); ok {
		_ = c.Close()
	}
	if m.Store != nil {
		return m.Store.Shutdown(ctx)
	}
	return nil
}

// FindAllWithSource 查找敏感词及其来源。
func (m *Manager) FindAllWithSource(text string) []filter.MatchResult {
	words := m.Filter.FindAll(text)
	if len(words) == 0 {
		return []filter.MatchResult{}
	}
	result := make([]filter.MatchResult, 0, len(words))
	seen := make(map[string]bool, len(words))
	for _, word := range words {
		if seen[word] {
			continue
		}
		seen[word] = true
		result = append(result, filter.MatchResult{
			Word:   word,
			Source: m.Store.GetWordSources(word),
		})
	}
	return result
}

// FindAllCountWithSource 查找敏感词、出现次数及来源。
func (m *Manager) FindAllCountWithSource(text string) map[string]filter.MatchResult {
	counts := m.Filter.FindAllCount(text)
	result := make(map[string]filter.MatchResult, len(counts))
	for word, count := range counts {
		result[word] = filter.MatchResult{
			Word:   word,
			Source: m.Store.GetWordSources(word),
			Count:  count,
		}
	}
	return result
}

// ExportToFile 导出词库到文件。
func (m *Manager) ExportToFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	writeErr := m.Store.ExportToWriter(f)
	closeErr := f.Close()
	if writeErr != nil {
		return writeErr
	}
	return closeErr
}

// MergeFromManager 合并另一个 Manager 的词库。
func (m *Manager) MergeFromManager(other *Manager) error {
	if other == nil {
		return errors.New("nil manager")
	}
	return m.Store.Merge(other.Store)
}

// RefreshFromPath 从文件刷新词库。
func (m *Manager) RefreshFromPath(path string, replace bool) error {
	if replace {
		if err := m.Store.Clear(); err != nil {
			return err
		}
	}
	return m.Store.LoadDictPath(path)
}
