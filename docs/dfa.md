# DFA

DFA 以前缀树存储词库，从文本每个位置尝试匹配，命中叶子节点即视为敏感词。

## 特点

- 实现简单、内存占用低。
- 通过 `sync.RWMutex` 保证并发安全，写操作会短暂阻塞读。
- 词库变更即时生效，无需重建。
- 最坏情况下复杂度高于 AC，适合小词库。

## 使用

```go
f, err := sensitive.NewFilter(
    sensitive.StoreOption{Type: sensitive.StoreMemory},
    sensitive.FilterOption{Type: sensitive.FilterDfa},
)
if err != nil {
    log.Fatal(err)
}
_ = f.LoadDictEmbed(sensitive.DictViolence)
f.Store.AddWord("自定义词")
_ = f.WaitForSync(ctx)

fmt.Println(f.Replace("包含自定义词的文本", '*'))
```

## 区间接口

DFA 实现 `FindAllRanges`，返回匹配的 rune 区间，供归一化包装器映射回原文并用于替换/删除。

## 基准

```bash
go test -bench=BenchmarkDFA -benchmem ./filter/
```
