# 最佳实践

## 按需加载词库

全量 16 个词库约 5.1 万词，其中 `零时-Tencent`、`非法网址`、`网易前端过滤敏感词库` 体积较大。按业务场景选择分类，避免不必要的内存与加载时间。

## 生产环境优先 AC

- 长文本、大词库、高并发场景选择 `FilterAc`。
- 小词库且高频增删可选择 `FilterDfa`。

## 更新后等待同步

词库写入是异步应用。写入后需要立即查询时调用 `WaitForSync`：

```go
f.Store.AddWord("新词")
if err := f.WaitForSync(ctx); err != nil {
    return err
}
```

## 批量写入

大量词条优先使用 `AddWords`、`LoadDictEmbed`，避免逐条 `AddWord`。

## 选择归一化策略

- 默认配置覆盖大小写与全角，满足大多数场景。
- 面向用户输入的内容审核建议使用 `WithStrictNormalizer`，可防御零宽字符、繁简、同形字绕过。
- 严格模式会增加处理开销，按需启用。

## 资源释放

长生命周期服务在退出时调用 `Close` 或 `Shutdown(ctx)`，避免 goroutine 与 ticker 泄漏。

## 并发

所有公开操作并发安全。查询可任意并发；写入会通过锁或原子切换保证一致性。

## 测试

```bash
make test
make test-race
```
