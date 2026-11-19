// Package normalize 提供文本与词条归一化。
package normalize

import (
	"unicode"

	"github.com/cysnb/go-sensitive-word/internal/jianfan"
)

// Config 定义文本归一化策略。
type Config struct {
	IgnoreCase         bool
	ToHalfWidth        bool
	IgnoreRepeat       bool
	IgnoreDigitType    bool
	IgnoreSimpTrad     bool
	IgnoreEnglishStyle bool
	RemoveZeroWidth    bool
	HomoglyphMap       map[rune]rune
}

// DefaultConfig 返回默认归一化配置：忽略大小写与全角转半角。
func DefaultConfig() Config {
	return Config{
		IgnoreCase:  true,
		ToHalfWidth: true,
	}
}

var digitMap = map[rune]rune{
	'０': '0', '１': '1', '２': '2', '３': '3', '４': '4', '５': '5', '６': '6', '７': '7', '８': '8', '９': '9',
	'⓪': '0', '①': '1', '②': '2', '③': '3', '④': '4', '⑤': '5', '⑥': '6', '⑦': '7', '⑧': '8', '⑨': '9',
	'⓿': '0', '❶': '1', '❷': '2', '❸': '3', '❹': '4', '❺': '5', '❻': '6', '❼': '7', '❽': '8', '❾': '9',
	'⁰': '0', '¹': '1', '²': '2', '³': '3', '⁴': '4', '⁵': '5', '⁶': '6', '⁷': '7', '⁸': '8', '⁹': '9',
	'₀': '0', '₁': '1', '₂': '2', '₃': '3', '₄': '4', '₅': '5', '₆': '6', '₇': '7', '₈': '8', '₉': '9',
	'➀': '1', '➁': '2', '➂': '3', '➃': '4', '➄': '5', '➅': '6', '➆': '7', '➇': '8', '➈': '9',
	'➊': '1', '➋': '2', '➌': '3', '➍': '4', '➎': '5', '➏': '6', '➐': '7', '➑': '8', '➒': '9',
	'㈠': '1', '㈡': '2', '㈢': '3', '㈣': '4', '㈤': '5', '㈥': '6', '㈦': '7', '㈧': '8', '㈨': '9',
	'㊀': '1', '㊁': '2', '㊂': '3', '㊃': '4', '㊄': '5', '㊅': '6', '㊆': '7', '㊇': '8', '㊈': '9',
	'零': '0', '一': '1', '二': '2', '三': '3', '四': '4', '五': '5', '六': '6', '七': '7', '八': '8', '九': '9',
	'〇': '0', '壹': '1', '贰': '2', '叁': '3', '肆': '4', '伍': '5', '陆': '6', '柒': '7', '捌': '8', '玖': '9',
}

var englishVariantMap = map[rune]rune{
	'Ⓕ': 'F', 'Ⓖ': 'G',
	'Ａ': 'A', 'Ｂ': 'B', 'Ｃ': 'C', 'Ｄ': 'D', 'Ｅ': 'E', 'Ｆ': 'F',
	'Ｇ': 'G', 'Ｈ': 'H', 'Ｉ': 'I', 'Ｊ': 'J', 'Ｋ': 'K', 'Ｌ': 'L',
	'Ｍ': 'M', 'Ｎ': 'N', 'Ｏ': 'O', 'Ｐ': 'P', 'Ｑ': 'Q', 'Ｒ': 'R',
	'Ｓ': 'S', 'Ｔ': 'T', 'Ｕ': 'U', 'Ｖ': 'V', 'Ｗ': 'W', 'Ｘ': 'X',
	'Ｙ': 'Y', 'Ｚ': 'Z', 'ａ': 'a', 'ｂ': 'b', 'ｃ': 'c', 'ｄ': 'd',
	'ｅ': 'e', 'ｆ': 'f', 'ｇ': 'g', 'ｈ': 'h', 'ｉ': 'i', 'ｊ': 'j',
	'ｋ': 'k', 'ｌ': 'l', 'ｍ': 'm', 'ｎ': 'n', 'ｏ': 'o', 'ｐ': 'p',
	'ｑ': 'q', 'ｒ': 'r', 'ｓ': 's', 'ｔ': 't', 'ｕ': 'u', 'ｖ': 'v',
	'ｗ': 'w', 'ｘ': 'x', 'ｙ': 'y', 'ｚ': 'z',
}

var zeroWidthChars = map[rune]bool{
	'\u200B': true,
	'\u200C': true,
	'\u200D': true,
	'\uFEFF': true,
	'\u200E': true,
	'\u200F': true,
	'\u202A': true,
	'\u202B': true,
	'\u202C': true,
	'\u202D': true,
	'\u202E': true,
	'\u2060': true,
	'\u2061': true,
	'\u2062': true,
	'\u2063': true,
	'\u2064': true,
}

var defaultHomoglyphMap = map[rune]rune{
	'Ａ': 'A', 'Ｂ': 'B', 'Ｃ': 'C', 'Ｄ': 'D', 'Ｅ': 'E', 'Ｆ': 'F',
	'Ｇ': 'G', 'Ｈ': 'H', 'Ｉ': 'I', 'Ｊ': 'J', 'Ｋ': 'K', 'Ｌ': 'L',
	'Ｍ': 'M', 'Ｎ': 'N', 'Ｏ': 'O', 'Ｐ': 'P', 'Ｑ': 'Q', 'Ｒ': 'R',
	'Ｓ': 'S', 'Ｔ': 'T', 'Ｕ': 'U', 'Ｖ': 'V', 'Ｗ': 'W', 'Ｘ': 'X',
	'Ｙ': 'Y', 'Ｚ': 'Z',
	'ａ': 'a', 'ｂ': 'b', 'ｃ': 'c', 'ｄ': 'd', 'ｅ': 'e', 'ｆ': 'f',
	'ｇ': 'g', 'ｈ': 'h', 'ｉ': 'i', 'ｊ': 'j', 'ｋ': 'k', 'ｌ': 'l',
	'ｍ': 'm', 'ｎ': 'n', 'ｏ': 'o', 'ｐ': 'p', 'ｑ': 'q', 'ｒ': 'r',
	'ｓ': 's', 'ｔ': 't', 'ｕ': 'u', 'ｖ': 'v', 'ｗ': 'w', 'ｘ': 'x',
	'ｙ': 'y', 'ｚ': 'z',
	'０': '0', '１': '1', '２': '2', '３': '3', '４': '4', '５': '5',
	'６': '6', '７': '7', '８': '8', '９': '9',
	'ο': 'o', 'а': 'a', 'е': 'e', 'о': 'o', 'р': 'p', 'с': 'c',
	'у': 'y', 'х': 'x', 'ι': 'i', 'τ': 't',
}

// DefaultHomoglyphMap 返回默认同形字映射表的副本。
func DefaultHomoglyphMap() map[rune]rune {
	m := make(map[rune]rune, len(defaultHomoglyphMap))
	for k, v := range defaultHomoglyphMap {
		m[k] = v
	}
	return m
}

func toHalfWidth(r rune) rune {
	if r == '\u3000' {
		return ' '
	}
	if r >= '\uff01' && r <= '\uff5e' {
		return r - 0xFEE0
	}
	return r
}

func normalizeRune(r rune, cfg Config) rune {
	if cfg.RemoveZeroWidth && zeroWidthChars[r] {
		return rune(-1)
	}
	if cfg.HomoglyphMap != nil {
		if mapped, ok := cfg.HomoglyphMap[r]; ok {
			r = mapped
		}
	}
	if cfg.IgnoreDigitType {
		if digit, ok := digitMap[r]; ok {
			r = digit
		}
	}
	if cfg.IgnoreSimpTrad {
		s := string(r)
		cn := jianfan.T2S(s)
		if runes := []rune(cn); len(runes) == 1 {
			r = runes[0]
		}
	}
	if cfg.IgnoreEnglishStyle {
		if eng, ok := englishVariantMap[r]; ok {
			r = eng
		}
	}
	if cfg.ToHalfWidth {
		r = toHalfWidth(r)
	}
	if cfg.IgnoreCase {
		r = unicode.ToLower(r)
	}
	return r
}

// NormalizeTextWithMap 归一化文本，并返回规范化索引到原始索引的映射。
func NormalizeTextWithMap(s string, cfg Config) (string, []int) {
	runes := []rune(s)
	norm := make([]rune, 0, len(runes))
	idxMap := make([]int, 0, len(runes))

	var last rune
	var hasLast bool
	for i, r := range runes {
		nr := normalizeRune(r, cfg)
		if nr == rune(-1) {
			continue
		}
		if cfg.IgnoreRepeat {
			if hasLast && nr == last {
				continue
			}
			last = nr
			hasLast = true
		}
		norm = append(norm, nr)
		idxMap = append(idxMap, i)
	}
	return string(norm), idxMap
}

// NormalizeWord 归一化词条，与文本侧保持同策略。
func NormalizeWord(word string, cfg Config) string {
	n, _ := NormalizeTextWithMap(word, cfg)
	return n
}
