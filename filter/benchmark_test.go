package filter

import (
	"fmt"
	"testing"
)

func benchmarkWords(n int) []string {
	words := make([]string, 0, n)
	for i := 0; i < n; i++ {
		words = append(words, fmt.Sprintf("敏感词%d", i))
	}
	return words
}

func benchmarkText() string {
	return "这是一个用于基准测试的文本，包含敏感词10和敏感词500，也包含一些普通内容用于模拟真实场景。"
}

func BenchmarkDFA_IsSensitive(b *testing.B) {
	m := NewDfaModel()
	m.AddWords(benchmarkWords(1000)...)
	text := benchmarkText()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.IsSensitive(text)
	}
}

func BenchmarkAC_IsSensitive(b *testing.B) {
	m := NewAcModel()
	m.AddWords(benchmarkWords(1000)...)
	text := benchmarkText()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.IsSensitive(text)
	}
}

func BenchmarkDFA_FindAll(b *testing.B) {
	m := NewDfaModel()
	m.AddWords(benchmarkWords(1000)...)
	text := benchmarkText()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.FindAll(text)
	}
}

func BenchmarkAC_FindAll(b *testing.B) {
	m := NewAcModel()
	m.AddWords(benchmarkWords(1000)...)
	text := benchmarkText()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.FindAll(text)
	}
}

func BenchmarkDFA_IsSensitive_Parallel(b *testing.B) {
	m := NewDfaModel()
	m.AddWords(benchmarkWords(1000)...)
	text := benchmarkText()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = m.IsSensitive(text)
		}
	})
}

func BenchmarkAC_IsSensitive_Parallel(b *testing.B) {
	m := NewAcModel()
	m.AddWords(benchmarkWords(1000)...)
	text := benchmarkText()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = m.IsSensitive(text)
		}
	})
}
