// Package jianfan 提供繁简中文转换。
package jianfan

import "unicode/utf8"

var t2sMapping = make(map[rune]rune)
var s2tMapping = make(map[rune]rune)

func init() {
	if len(ChT) != len(ChS) {
		panic("cht and chs data length not equal")
	}
	for index, runeValueT := range ChT {
		runeValueS, _ := utf8.DecodeRuneInString(ChS[index:])
		t2sMapping[runeValueT] = runeValueS
		s2tMapping[runeValueS] = runeValueT
	}
}

// T2S 将繁体中文转换为简体中文。
func T2S(s string) string {
	chs := make([]rune, 0, len(s))
	for _, r := range s {
		if v, ok := t2sMapping[r]; ok {
			chs = append(chs, v)
		} else {
			chs = append(chs, r)
		}
	}
	return string(chs)
}

// S2T 将简体中文转换为繁体中文。
func S2T(s string) string {
	cht := make([]rune, 0, len(s))
	for _, r := range s {
		if v, ok := s2tMapping[r]; ok {
			cht = append(cht, v)
		} else {
			cht = append(cht, r)
		}
	}
	return string(cht)
}
