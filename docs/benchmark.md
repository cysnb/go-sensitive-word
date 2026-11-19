# 性能基准

## 运行

```bash
make bench
go test -bench=BenchmarkAC -benchmem ./filter/
go test -bench=BenchmarkDFA -benchmem ./filter/
go test -bench=BenchmarkAC_IsSensitive_Parallel -benchmem ./filter/
```

## 参考数据

环境：Intel Core i9-9980HK，词库 1000 词，Go 1.27。

| 操作 | DFA | AC |
|------|-----|-----|
| IsSensitive | 558 ns/op | 187 ns/op |
| FindAll | 1510 ns/op | 839 ns/op |
| IsSensitive（并发） | 213 ns/op | 48 ns/op |

## 说明

- AC 通过失败指针实现单次扫描，在长文本与大词库下优势明显。
- DFA 实现简单、更新即时，适合小词库。
- 具体数值随机器、词库规模与文本长度变化。
