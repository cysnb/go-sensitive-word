# 算法选择

项目支持 DFA 与 AC 自动机两种算法。

## 对比

| 维度 | DFA | AC |
|------|-----|-----|
| 实现 | 前缀树逐起点回退 | 失败指针 + 单次扫描 |
| 时间复杂度 | 最坏 O(n*m) | O(n) |
| 并发读 | RWMutex 保护 | atomic.Pointer + COW，无锁读 |
| 动态更新 | 即时生效 | 100ms 窗口合并或 1000 条批量重建 |
| 适用场景 | 小词库、更新频繁 | 大词库、长文本、生产环境 |

## 选择建议

- 词库小于 1000 且需要频繁增删：DFA。
- 词库大、文本长、并发高：AC。
- 无法确定时：AC。

## 切换算法

```go
dfa, _ := sensitive.NewFilter(
    sensitive.StoreOption{Type: sensitive.StoreMemory},
    sensitive.FilterOption{Type: sensitive.FilterDfa},
)

ac, _ := sensitive.NewFilter(
    sensitive.StoreOption{Type: sensitive.StoreMemory},
    sensitive.FilterOption{Type: sensitive.FilterAc},
)
```

`FilterAc` 与 `FilterAC` 等价。

## 基准测试

```bash
make bench
go test -bench=BenchmarkAC_IsSensitive -benchmem ./filter/
go test -bench=BenchmarkDFA_IsSensitive -benchmem ./filter/
```
