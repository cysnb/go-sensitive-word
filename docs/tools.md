# 字符串工具

用于敏感词过滤之外的常见敏感信息检测与脱敏。

## 邮箱

```go
sensitive.HasEmail("联系 user@example.com")   // true
sensitive.MaskEmail("联系 user@example.com")  // 联系 ***
```

## 网址

```go
sensitive.HasURL("访问 https://example.com")  // true
sensitive.MaskURL("访问 https://example.com") // 访问 ***
```

## 数字

```go
sensitive.HasDigit("abc123", 3) // true
sensitive.HasDigit("abc12", 3)  // false
sensitive.MaskDigit("a1b2")     // a*b*
```

## 微信号

```go
sensitive.HasWechatID("加我 abcdef123")   // true
sensitive.MaskWechatID("加我 abcdef123")  // 加我 ***
```

## 说明

- 邮箱、网址、微信号按整体替换为 `***`。
- 数字按字符逐个替换为 `*`。
- 这些函数基于正则表达式，不依赖过滤器实例。
