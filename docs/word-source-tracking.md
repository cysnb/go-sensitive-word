# 词库来源追踪

来源追踪记录每个词来自哪个词库，便于审计与按来源治理。

## 带来源添加

```go
f.Store.AddWordsWithSource([]string{"违禁词A"}, "custom")
f.Store.AddWordsWithSource([]string{"违禁词A"}, "political")
f.WaitForSync(ctx)
```

同一个词可以关联多个来源。

## 带来源加载内置词库

```go
f.LoadDictEmbedWithSource(sensitive.DictViolence, "violence")
```

## 查询来源

```go
fmt.Println(f.GetWordSources("违禁词A")) // [custom political]
fmt.Println(f.GetAllWordSources())
```

未知词返回空结果。返回值为副本，修改不影响内部状态。

## 带来源的匹配

```go
for _, r := range f.FindAllWithSource("包含违禁词A") {
    fmt.Println(r.Word, r.Source)
}

counts := f.FindAllCountWithSource("违禁词A违禁词A")
fmt.Println(counts["违禁词A"].Count)  // 2
fmt.Println(counts["违禁词A"].Source) // [custom political]
```

## 与归一化一致

查询来源时会应用与添加词条相同的归一化策略，因此大小写差异不影响查询：

```go
f.Store.AddWordsWithSource([]string{"ABC"}, "custom")
f.WaitForSync(ctx)
fmt.Println(f.GetWordSources("abc")) // [custom]
```
