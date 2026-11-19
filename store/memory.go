package store

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cysnb/go-sensitive-word/fetch"
	"github.com/cysnb/go-sensitive-word/internal/normalize"
)

// ErrStoreClosed 词库已关闭。
var ErrStoreClosed = errors.New("store closed")

// Option 词库配置选项。
type Option func(*MemoryModel)

// WithWordNormalizer 注入词条归一化函数。
func WithWordNormalizer(fn func(string) string) Option {
	return func(m *MemoryModel) {
		if fn != nil {
			m.normalizeWord = fn
		}
	}
}

func defaultWordNormalizer(word string) string {
	return normalize.NormalizeWord(strings.TrimSpace(word), normalize.DefaultConfig())
}

// MemoryModel 基于标准库实现的并发安全内存词库。
type MemoryModel struct {
	store       map[string]struct{}
	wordSources map[string][]string
	storeMu     sync.RWMutex
	totalWords  atomic.Int64

	addChan  chan string
	delChan  chan string
	closedCh chan struct{}
	closing  bool

	queueMu   sync.Mutex
	queueCond *sync.Cond
	addQueue  []string
	delQueue  []string

	pending   atomic.Int64
	closeOnce sync.Once

	statsMu sync.RWMutex
	stats   Stats

	normalizeWord func(string) string
}

// NewMemoryModel 创建内存词库。
func NewMemoryModel(opts ...Option) *MemoryModel {
	m := &MemoryModel{
		store:         make(map[string]struct{}),
		wordSources:   make(map[string][]string),
		addChan:       make(chan string, 8192),
		delChan:       make(chan string, 8192),
		closedCh:      make(chan struct{}),
		normalizeWord: defaultWordNormalizer,
		stats: Stats{
			LastUpdate: time.Now(),
			Source:     make([]string, 0),
		},
	}
	m.queueCond = sync.NewCond(&m.queueMu)
	for _, opt := range opts {
		opt(m)
	}
	go m.dispatch()
	return m
}

func (m *MemoryModel) dispatch() {
	for {
		m.queueMu.Lock()
		for len(m.addQueue) == 0 && len(m.delQueue) == 0 && !m.closing {
			m.queueCond.Wait()
		}
		if m.closing {
			m.queueMu.Unlock()
			return
		}
		var add, del string
		hasAdd, hasDel := false, false
		if len(m.addQueue) > 0 {
			add = m.addQueue[0]
			m.addQueue = m.addQueue[1:]
			hasAdd = true
		}
		if len(m.delQueue) > 0 {
			del = m.delQueue[0]
			m.delQueue = m.delQueue[1:]
			hasDel = true
		}
		m.queueMu.Unlock()

		if hasAdd {
			select {
			case m.addChan <- add:
			case <-m.closedCh:
				return
			}
		}
		if hasDel {
			select {
			case m.delChan <- del:
			case <-m.closedCh:
				return
			}
		}
	}
}

func (m *MemoryModel) enqueue(queue *[]string, word string) error {
	m.queueMu.Lock()
	defer m.queueMu.Unlock()
	if m.closing {
		return ErrStoreClosed
	}
	*queue = append(*queue, word)
	m.pending.Add(1)
	m.queueCond.Signal()
	return nil
}

// Done 标记一条通知已被过滤器处理。
func (m *MemoryModel) Done() {
	for {
		p := m.pending.Load()
		if p <= 0 {
			return
		}
		if m.pending.CompareAndSwap(p, p-1) {
			return
		}
	}
}

// WaitForSync 等待全部词库通知被过滤器处理。
func (m *MemoryModel) WaitForSync(ctx context.Context) error {
	select {
	case <-m.closedCh:
		return ErrStoreClosed
	default:
	}
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	for {
		if m.pending.Load() == 0 {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (m *MemoryModel) touch(count int) {
	m.statsMu.Lock()
	m.stats.TotalWords = int(m.totalWords.Load())
	m.stats.LastUpdate = time.Now()
	m.stats.UpdateCount += count
	m.statsMu.Unlock()
}

func (m *MemoryModel) recordSource(source string) {
	if source == "" {
		return
	}
	m.statsMu.Lock()
	m.stats.Source = append(m.stats.Source, source)
	m.statsMu.Unlock()
}

// LoadDictPath 从本地文件加载词库。
func (m *MemoryModel) LoadDictPath(paths ...string) error {
	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		err = m.LoadDict(f)
		closeErr := f.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
		m.recordSource("file://" + path)
	}
	return nil
}

// LoadDictEmbed 加载嵌入式词库。
func (m *MemoryModel) LoadDictEmbed(contents ...string) error {
	for _, content := range contents {
		if err := m.LoadDict(strings.NewReader(content)); err != nil {
			return err
		}
	}
	return nil
}

// LoadDictHttp 从远程 URL 加载词库。
func (m *MemoryModel) LoadDictHttp(urls ...string) error {
	for _, url := range urls {
		resp, err := fetch.FetchUrlSimple(url)
		if err != nil {
			return err
		}
		if resp == nil {
			return errors.New("nil http response")
		}
		if resp.StatusCode != http.StatusOK {
			_ = resp.Body.Close()
			return fmt.Errorf("load dict %s: %s", url, resp.Status)
		}
		err = m.LoadDict(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return err
		}
		m.recordSource("http://" + url)
	}
	return nil
}

// LoadDict 从 Reader 按行加载词库。
func (m *MemoryModel) LoadDict(reader io.Reader) error {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		word := m.normalizeWord(scanner.Text())
		if word == "" {
			continue
		}
		if err := m.addNormalized(word, ""); err != nil {
			return err
		}
	}
	return scanner.Err()
}

// LoadDictCallback 通过回调加载词库。
func (m *MemoryModel) LoadDictCallback(loader DictLoader, source string) error {
	if loader == nil {
		return errors.New("loader callback is nil")
	}
	words, err := loader()
	if err != nil {
		return err
	}
	if err := m.AddWords(words); err != nil {
		return err
	}
	m.recordSource("callback://" + source)
	return nil
}

// LoadDictEmbedWithSource 加载嵌入词库并记录来源。
func (m *MemoryModel) LoadDictEmbedWithSource(content string, source string) error {
	words := make([]string, 0)
	for _, line := range strings.Split(content, "\n") {
		word := strings.TrimSpace(line)
		if word == "" {
			continue
		}
		words = append(words, word)
	}
	return m.AddWordsWithSource(words, source)
}

func (m *MemoryModel) addNormalized(word, source string) error {
	m.storeMu.Lock()
	if _, exists := m.store[word]; !exists {
		m.store[word] = struct{}{}
		m.totalWords.Add(1)
	}
	if source != "" {
		sources := m.wordSources[word]
		exists := false
		for _, s := range sources {
			if s == source {
				exists = true
				break
			}
		}
		if !exists {
			m.wordSources[word] = append(sources, source)
		}
	}
	m.storeMu.Unlock()
	if err := m.enqueue(&m.addQueue, word); err != nil {
		return err
	}
	m.touch(0)
	return nil
}

func (m *MemoryModel) addWordsWithSource(words []string, source string) error {
	count := 0
	for _, raw := range words {
		word := m.normalizeWord(raw)
		if word == "" {
			continue
		}
		m.storeMu.Lock()
		if _, exists := m.store[word]; !exists {
			m.store[word] = struct{}{}
			m.totalWords.Add(1)
			count++
		}
		if source != "" {
			sources := m.wordSources[word]
			exists := false
			for _, s := range sources {
				if s == source {
					exists = true
					break
				}
			}
			if !exists {
				m.wordSources[word] = append(sources, source)
			}
		}
		m.storeMu.Unlock()
		if err := m.enqueue(&m.addQueue, word); err != nil {
			return err
		}
	}
	m.touch(count)
	return nil
}

// AddWord 添加一个或多个词。
func (m *MemoryModel) AddWord(words ...string) error {
	return m.addWordsWithSource(words, "")
}

// AddWords 批量添加词。
func (m *MemoryModel) AddWords(words []string) error {
	return m.addWordsWithSource(words, "")
}

// AddWordsWithSource 批量添加词并记录来源。
func (m *MemoryModel) AddWordsWithSource(words []string, source string) error {
	return m.addWordsWithSource(words, source)
}

// DelWord 删除一个或多个词。
func (m *MemoryModel) DelWord(words ...string) error {
	return m.DelWords(words)
}

// DelWords 批量删除词。
func (m *MemoryModel) DelWords(words []string) error {
	count := 0
	for _, raw := range words {
		word := m.normalizeWord(raw)
		if word == "" {
			continue
		}
		m.storeMu.Lock()
		if _, exists := m.store[word]; exists {
			delete(m.store, word)
			delete(m.wordSources, word)
			m.totalWords.Add(-1)
			count++
		}
		m.storeMu.Unlock()
		if err := m.enqueue(&m.delQueue, word); err != nil {
			return err
		}
	}
	m.touch(count)
	return nil
}

// ReplaceWords 批量替换：先删除旧词，再添加新词。
func (m *MemoryModel) ReplaceWords(oldWords, newWords []string) error {
	if err := m.DelWords(oldWords); err != nil {
		return err
	}
	return m.AddWords(newWords)
}

// ReadChan 通过通道逐个读取词条。
func (m *MemoryModel) ReadChan() <-chan string {
	ch := make(chan string)
	words := m.ReadString()
	go func() {
		defer close(ch)
		for _, word := range words {
			select {
			case ch <- word:
			case <-m.closedCh:
				return
			}
		}
	}()
	return ch
}

// ReadString 返回全部词条。
func (m *MemoryModel) ReadString() []string {
	m.storeMu.RLock()
	res := make([]string, 0, len(m.store))
	for word := range m.store {
		res = append(res, word)
	}
	m.storeMu.RUnlock()
	sort.Strings(res)
	return res
}

// GetAddChan 返回新增词通知通道。
func (m *MemoryModel) GetAddChan() <-chan string { return m.addChan }

// GetDelChan 返回删除词通知通道。
func (m *MemoryModel) GetDelChan() <-chan string { return m.delChan }

// GetWordSources 获取指定词的来源列表。
func (m *MemoryModel) GetWordSources(word string) []string {
	word = m.normalizeWord(word)
	m.storeMu.RLock()
	defer m.storeMu.RUnlock()
	sources := m.wordSources[word]
	if len(sources) == 0 {
		return nil
	}
	res := make([]string, len(sources))
	copy(res, sources)
	return res
}

// GetAllWordSources 获取全部词的来源映射。
func (m *MemoryModel) GetAllWordSources() map[string][]string {
	m.storeMu.RLock()
	defer m.storeMu.RUnlock()
	res := make(map[string][]string, len(m.wordSources))
	for word, sources := range m.wordSources {
		res[word] = append([]string{}, sources...)
	}
	return res
}

// ExportToWriter 导出词库到 Writer，每行一个词。
func (m *MemoryModel) ExportToWriter(w io.Writer) error {
	writer := bufio.NewWriter(w)
	for _, word := range m.ReadString() {
		if _, err := writer.WriteString(word + "\n"); err != nil {
			return err
		}
	}
	return writer.Flush()
}

// ExportToString 导出词库为字符串。
func (m *MemoryModel) ExportToString() (string, error) {
	var builder strings.Builder
	for _, word := range m.ReadString() {
		builder.WriteString(word)
		builder.WriteString("\n")
	}
	return builder.String(), nil
}

// GetStats 获取统计信息。
func (m *MemoryModel) GetStats() Stats {
	m.statsMu.RLock()
	defer m.statsMu.RUnlock()
	return Stats{
		TotalWords:  int(m.totalWords.Load()),
		LastUpdate:  m.stats.LastUpdate,
		UpdateCount: m.stats.UpdateCount,
		Source:      append([]string{}, m.stats.Source...),
	}
}

// Clear 清空词库。
func (m *MemoryModel) Clear() error {
	if err := m.DelWords(m.ReadString()); err != nil {
		return err
	}
	m.storeMu.Lock()
	m.totalWords.Store(0)
	m.wordSources = make(map[string][]string)
	m.storeMu.Unlock()
	m.statsMu.Lock()
	m.stats.TotalWords = 0
	m.stats.UpdateCount = 0
	m.stats.Source = make([]string, 0)
	m.stats.LastUpdate = time.Now()
	m.statsMu.Unlock()
	return nil
}

// Merge 合并另一个词库，重复词只保留一份。
func (m *MemoryModel) Merge(other Store) error {
	if other == nil {
		return errors.New("nil store")
	}
	return m.AddWords(other.ReadString())
}

// Close 关闭词库。
func (m *MemoryModel) Close() error {
	m.closeOnce.Do(func() {
		m.queueMu.Lock()
		m.closing = true
		m.queueCond.Broadcast()
		m.queueMu.Unlock()
		close(m.closedCh)
	})
	return nil
}

// Shutdown 带超时关闭词库。
func (m *MemoryModel) Shutdown(ctx context.Context) error {
	done := make(chan error, 1)
	go func() { done <- m.Close() }()
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

var _ Store = (*MemoryModel)(nil)
