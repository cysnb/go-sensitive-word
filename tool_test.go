package go_sensitive_word

import (
	"strings"
	"testing"
)

func TestStringTools(t *testing.T) {
	t.Run("email", func(t *testing.T) {
		if !HasEmail("联系 user@example.com") {
			t.Fatal("HasEmail should be true")
		}
		if HasEmail("没有邮箱") {
			t.Fatal("HasEmail should be false")
		}
		got := MaskEmail("联系 user@example.com 谢谢")
		if strings.Contains(got, "user@example.com") || !strings.Contains(got, "***") {
			t.Fatalf("MaskEmail = %q", got)
		}
	})

	t.Run("url", func(t *testing.T) {
		if !HasURL("访问 https://example.com/path") {
			t.Fatal("HasURL should be true")
		}
		got := MaskURL("https://example.com/path 结束")
		if strings.Contains(got, "example.com") {
			t.Fatalf("MaskURL = %q", got)
		}
	})

	t.Run("digit", func(t *testing.T) {
		if !HasDigit("abc123", 3) {
			t.Fatal("HasDigit should be true")
		}
		if HasDigit("abc12", 3) {
			t.Fatal("HasDigit should be false")
		}
		if got := MaskDigit("a1b2"); got != "a*b*" {
			t.Fatalf("MaskDigit = %q", got)
		}
	})

	t.Run("wechat", func(t *testing.T) {
		if !HasWechatID("加我 abcdef123") {
			t.Fatal("HasWechatID should be true")
		}
		got := MaskWechatID("加我 abcdef123")
		if strings.Contains(got, "abcdef123") {
			t.Fatalf("MaskWechatID = %q", got)
		}
	})
}
