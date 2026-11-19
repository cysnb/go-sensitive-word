# Unicode 相似字符攻击

攻击者可能使用形近字符替换敏感词中的字符，例如用西里尔字母 `а` 代替拉丁字母 `a`，或用全角字母、数学字母等变体绕过检测。

## 默认能力

- 全角转半角：`ｆｕｃｋ` → `fuck`
- 忽略大小写：`FuCk` → `fuck`

## 严格能力

`StrictNormalizer` 额外启用：

- `IgnoreEnglishStyle`：花体、全角字母等英文变体归一
- `HomoglyphMap`：默认同形字映射（西里尔、希腊字母等）

```go
f, _ := sensitive.NewFilter(
    sensitive.StoreOption{Type: sensitive.StoreMemory},
    sensitive.FilterOption{Type: sensitive.FilterDfa},
    sensitive.WithStrictNormalizer(),
)
```

## 自定义同形字映射

```go
cfg := sensitive.DefaultNormalizer()
cfg.HomoglyphMap = map[rune]rune{
    'а': 'a',
    'е': 'e',
    'о': 'o',
}
f, _ := sensitive.NewFilter(
    sensitive.StoreOption{Type: sensitive.StoreMemory},
    sensitive.FilterOption{Type: sensitive.FilterDfa},
    sensitive.WithNormalizer(cfg),
)
```

映射表的扩展取决于业务面临的攻击面；建议结合实际样本逐步补充。
