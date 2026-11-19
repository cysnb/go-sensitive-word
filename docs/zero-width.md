# 零宽字符攻击

零宽字符不可见但会插入到文本中，例如：

- `U+200B` Zero Width Space
- `U+200C` Zero Width Non-Joiner
- `U+200D` Zero Width Joiner
- `U+FEFF` BOM

攻击者可能在敏感词字符之间插入零宽字符，使字面匹配失败。

## 开启零宽剔除

```go
f, _ := sensitive.NewFilter(
    sensitive.StoreOption{Type: sensitive.StoreMemory},
    sensitive.FilterOption{Type: sensitive.FilterDfa},
    sensitive.WithStrictNormalizer(),
)
```

或自定义：

```go
cfg := sensitive.DefaultNormalizer()
cfg.RemoveZeroWidth = true
```

## 行为

文本 `f\u200Buck` 在开启剔除后归一化为 `fuck`：

- `FindOne` 返回原文片段 `f\u200Buck`
- `Replace` 将命中覆盖的原文区间全部替换
- `Remove` 移除命中区间（包括夹带的零宽字符）

## 注意

零宽字符剔除会改变索引映射。本项目通过规范化索引到原始索引的映射保证结果仍作用于原文。
