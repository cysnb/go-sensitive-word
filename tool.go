package go_sensitive_word

import "regexp"

var (
	emailRegex  = regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	urlRegex    = regexp.MustCompile(`https?://[^\s]+`)
	digitRegex  = regexp.MustCompile(`\d`)
	wechatRegex = regexp.MustCompile(`[a-zA-Z][a-zA-Z0-9_-]{5,19}`)
)

// HasEmail 判断字符串中是否存在邮箱地址。
func HasEmail(s string) bool {
	return emailRegex.MatchString(s)
}

// MaskEmail 将字符串中存在的邮箱地址替换为固定掩码。
func MaskEmail(s string) string {
	return emailRegex.ReplaceAllStringFunc(s, func(string) string { return "***" })
}

// HasURL 判断字符串中是否存在网址。
func HasURL(s string) bool {
	return urlRegex.MatchString(s)
}

// MaskURL 将字符串中存在的网址替换为固定掩码。
func MaskURL(s string) string {
	return urlRegex.ReplaceAllStringFunc(s, func(string) string { return "***" })
}

// HasDigit 判断字符串中是否存在至少 count 个数字。
func HasDigit(s string, count int) bool {
	return len(digitRegex.FindAllString(s, -1)) >= count
}

// MaskDigit 将字符串中的每个数字替换为 *。
func MaskDigit(s string) string {
	return digitRegex.ReplaceAllStringFunc(s, func(string) string { return "*" })
}

// HasWechatID 判断字符串中是否存在微信号。
func HasWechatID(s string) bool {
	return wechatRegex.MatchString(s)
}

// MaskWechatID 将字符串中存在的微信号替换为固定掩码。
func MaskWechatID(s string) string {
	return wechatRegex.ReplaceAllStringFunc(s, func(string) string { return "***" })
}
