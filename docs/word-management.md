# 词库管理

## 动态增删

```go
f.Store.AddWord("秦始皇", "李世民")
f.Store.DelWord("武汉")
f.WaitForSync(ctx) // 等待词库变更应用完成
```

批量接口：

```go
f.Store.AddWords([]string{"a", "b"})
f.Store.DelWords([]string{"a"})
f.Store.ReplaceWords([]string{"old"}, []string{"new"})
```

## 读取与导出

```go
words := f.Store.ReadString()

for word := range f.Store.ReadChan() {
    fmt.Println(word)
}

s, _ := f.Store.ExportToString()
_ = f.Store.ExportToFile("export.txt") // Manager 方法
```

`ExportToWriter` 可用于写入任意 `io.Writer`。

## 统计信息

```go
stats := f.GetStats()
fmt.Println(stats.TotalWords, stats.LastUpdate, stats.UpdateCount, stats.Source)
```

返回值为副本，修改不会影响内部状态。

## 合并与清空

```go
_ = f.Merge(otherFilter.Store) // 去重合并
_ = f.Clear()                  // 清空词库
```

## 更新可见性

词库写入通过后台通道异步应用到过滤器。调用 `WaitForSync` 可确保后续查询看到最新词库：

```go
f.Store.AddWord("新词")
if err := f.WaitForSync(ctx); err != nil {
    log.Fatal(err)
}
fmt.Println(f.IsSensitive("包含新词"))
```
