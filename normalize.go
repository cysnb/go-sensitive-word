package go_sensitive_word

import "github.com/cysnb/go-sensitive-word/internal/normalize"

// NormalizerConfig 定义文本归一化策略。
type NormalizerConfig struct {
	IgnoreCase         bool
	ToHalfWidth        bool
	IgnoreRepeat       bool
	IgnoreDigitType    bool
	IgnoreSimpTrad     bool
	IgnoreEnglishStyle bool
	RemoveZeroWidth    bool
	HomoglyphMap       map[rune]rune
}

func (c NormalizerConfig) toInternalConfig() normalize.Config {
	return normalize.Config{
		IgnoreCase:         c.IgnoreCase,
		ToHalfWidth:        c.ToHalfWidth,
		IgnoreRepeat:       c.IgnoreRepeat,
		IgnoreDigitType:    c.IgnoreDigitType,
		IgnoreSimpTrad:     c.IgnoreSimpTrad,
		IgnoreEnglishStyle: c.IgnoreEnglishStyle,
		RemoveZeroWidth:    c.RemoveZeroWidth,
		HomoglyphMap:       c.HomoglyphMap,
	}
}

// DefaultNormalizer 返回默认归一化配置：忽略大小写与全角转半角。
func DefaultNormalizer() NormalizerConfig {
	return NormalizerConfig{
		IgnoreCase:  true,
		ToHalfWidth: true,
	}
}

// StrictNormalizer 返回严格归一化配置，启用全部防绕过选项。
func StrictNormalizer() NormalizerConfig {
	return NormalizerConfig{
		IgnoreCase:         true,
		ToHalfWidth:        true,
		IgnoreRepeat:       true,
		IgnoreDigitType:    true,
		IgnoreSimpTrad:     true,
		IgnoreEnglishStyle: true,
		RemoveZeroWidth:    true,
		HomoglyphMap:       normalize.DefaultHomoglyphMap(),
	}
}

// NormalizeTextWithMap 归一化文本，并返回规范化索引到原始索引的映射。
func NormalizeTextWithMap(s string, cfg NormalizerConfig) (string, []int) {
	return normalize.NormalizeTextWithMap(s, cfg.toInternalConfig())
}

// NormalizeWord 归一化词条，与文本侧保持同策略。
func NormalizeWord(word string, cfg NormalizerConfig) string {
	return normalize.NormalizeWord(word, cfg.toInternalConfig())
}
