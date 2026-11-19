// Package filter 提供 DFA 与 AC 敏感词匹配算法。
package filter

// Range 表示匹配区间 [Start, End]，闭区间。
type Range struct {
	Start int
	End   int
}

// MatchResult 包含匹配词、来源与出现次数。
type MatchResult struct {
	Word   string
	Source []string
	Count  int
}

// RangedFilter 返回匹配区间而非字符串，包装器优先使用以提高性能。
type RangedFilter interface {
	FindAllRanges(text string) []Range
}

// Filter 定义敏感词匹配器。
type Filter interface {
	FindAll(text string) []string
	FindAllCount(text string) map[string]int
	FindOne(text string) string
	IsSensitive(text string) bool
	Replace(text string, repl rune) string
	Remove(text string) string
}
