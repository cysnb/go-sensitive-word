// Package store 提供敏感词词库的存储与维护能力。
package store

import (
	"context"
	"io"
	"time"
)

// Stats 词库统计信息。
type Stats struct {
	TotalWords  int
	LastUpdate  time.Time
	UpdateCount int
	Source      []string
}

// DictLoader 词库加载回调。
type DictLoader func() ([]string, error)

// Store 定义词库存储接口。
type Store interface {
	LoadDictPath(path ...string) error
	LoadDictEmbed(contents ...string) error
	LoadDictHttp(url ...string) error
	LoadDict(reader io.Reader) error
	LoadDictCallback(loader DictLoader, source string) error
	LoadDictEmbedWithSource(content string, source string) error

	ReadChan() <-chan string
	ReadString() []string

	GetAddChan() <-chan string
	GetDelChan() <-chan string
	AddWord(words ...string) error
	DelWord(words ...string) error
	AddWords(words []string) error
	DelWords(words []string) error
	ReplaceWords(oldWords, newWords []string) error

	AddWordsWithSource(words []string, source string) error
	GetWordSources(word string) []string
	GetAllWordSources() map[string][]string

	ExportToWriter(w io.Writer) error
	ExportToString() (string, error)
	GetStats() Stats
	Clear() error
	Merge(other Store) error

	WaitForSync(ctx context.Context) error
	Close() error
	Shutdown(ctx context.Context) error
}
