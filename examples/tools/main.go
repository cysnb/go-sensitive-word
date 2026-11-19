package main

import (
	"fmt"

	sensitive "github.com/cysnb/go-sensitive-word"
)

func main() {
	fmt.Println(sensitive.HasEmail("联系 user@example.com"))
	fmt.Println(sensitive.MaskEmail("联系 user@example.com"))
	fmt.Println(sensitive.HasURL("访问 https://example.com"))
	fmt.Println(sensitive.MaskURL("访问 https://example.com"))
	fmt.Println(sensitive.HasDigit("abc123", 3))
	fmt.Println(sensitive.MaskDigit("a1b2"))
	fmt.Println(sensitive.HasWechatID("加我 abcdef123"))
	fmt.Println(sensitive.MaskWechatID("加我 abcdef123"))
}
