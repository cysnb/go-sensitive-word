# AC 自动机

AC（Aho-Corasick）通过失败指针实现单次线性扫描，适合大词库与长文本。

## 并发模型

- 读操作通过 `atomic.Pointer` 读取当前根节点，无锁。
- 写操作在互斥锁内深拷贝整棵树，修改后重建失败指针，再原子切换根指针（Copy-on-Write）。
- 通过通道接收词库变更，100ms 窗口或累计 1000 条时批量重建，降低重建频率。

## 使用

```go
f, err := sensitive.NewFilter(
    sensitive.StoreOption{Type: sensitive.StoreMemory},
    sensitive.FilterOption{Type: sensitive.FilterAc},
)
if err != nil {
    log.Fatal(err)
}

_ = f.LoadDictEmbed(sensitive.DictPolitical, sensitive.DictViolence)
f.Store.AddWord("自定义词")
_ = f.WaitForSync(ctx)

fmt.Println(f.FindAll("包含自定义词的文本"))
```

## 注意

- 词库变更存在最大 100ms 的窗口延迟；需要立即生效时调用 `WaitForSync`。
- 动态更新会重建整棵树，超大规模词库且高频更新时应合并批量写入。

## 基准

```bash
go test -bench=BenchmarkAC -benchmem ./filter/
```
