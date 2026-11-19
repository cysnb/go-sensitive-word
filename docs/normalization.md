# 归一化

归一化对**词库词条**与**待检测文本**使用同一套策略，防止大小写、全角、零宽字符、繁简、同形字等绕过。

## 默认配置

默认启用：

- `IgnoreCase`：忽略大小写
- `ToHalfWidth`：全角转半角

```go
cfg := sensitive.DefaultNormalizer()
// IgnoreCase=true, ToHalfWidth=true
```

## 严格配置

```go
cfg := sensitive.StrictNormalizer()
// 额外启用：
// IgnoreRepeat       连续重复字符压缩
// IgnoreDigitType    各类数字写法归一
// IgnoreSimpTrad     繁体转简体
// IgnoreEnglishStyle 英文变体归一
// RemoveZeroWidth    剔除零宽字符
// HomoglyphMap       同形字映射
```

## 自定义配置

```go
cfg := sensitive.NormalizerConfig{
    IgnoreCase:      true,
    ToHalfWidth:     true,
    RemoveZeroWidth: true,
    HomoglyphMap: map[rune]rune{
        'а': 'a', // 西里尔字母 а
    },
}
```

## 在过滤器中启用

`NewFilter` 的第三个参数为可选配置：

```go
f, err := sensitive.NewFilter(
    sensitive.StoreOption{Type: sensitive.StoreMemory},
    sensitive.FilterOption{Type: sensitive.FilterDfa},
    sensitive.WithNormalizer(cfg),
)
```

使用严格配置：

```go
f, err := sensitive.NewFilter(
    sensitive.StoreOption{Type: sensitive.StoreMemory},
    sensitive.FilterOption{Type: sensitive.FilterAc},
    sensitive.WithStrictNormalizer(),
)
```

## 结果映射

归一化只用于匹配判断，返回结果始终是**原文片段**。例如词库含 `fuck`：

```go
text := "说 ＦＵＣＫ 吧"
fmt.Println(f.FindOne(text))        // ＦＵＣＫ
fmt.Println(f.Replace(text, '*'))   // 说 **** 吧
fmt.Println(f.Remove(text))         // 说  吧
```

## 单独使用归一化函数

```go
fmt.Println(sensitive.NormalizeWord("ＦｕＣｋ", sensitive.DefaultNormalizer())) // fuck

norm, idxMap := sensitive.NormalizeTextWithMap("Ａb", sensitive.DefaultNormalizer())
// norm = "ab", idxMap = [0 1]
```
