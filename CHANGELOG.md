# Changelog

本项目为 `zmexing/go-sensitive-word` 与 `LuYongwang/go-sensitive-word` 的整合维护版。

## [1.2.0] - 2026-11-19

### 整合

- 以两个上游最新版本为基础，合并为单一超集库。
- 采用 zmexing 风格的公开包布局：`filter`、`store`、`fetch`。
- 保留 LuYongwang 的增强能力并下沉到 `store.Store` 接口。

### 新增

- DFA 与 AC 双算法，`FilterAc` 与 `FilterAC` 同时可用。
- AC 自动机采用 `atomic.Pointer` + Copy-on-Write + 100ms/1000 条窗口合并。
- 16 个内置分类词库，去重并集 51,103 词。
- 默认归一化（大小写 + 全角）与严格归一化（重复、数字、繁简、英文变体、零宽、同形字）。
- 词库来源追踪：`AddWordsWithSource`、`GetWordSources`、`FindAllWithSource` 等。
- 词库管理：批量增删、替换、统计、导出、合并、清空。
- 五种加载方式：内置、文件、HTTP、Reader、回调。
- 字符串工具：邮箱、URL、数字、微信号检测与脱敏。
- `WaitForSync` 确定性等待词库变更应用。
- 幂等 `Close` / `Shutdown`，修复过滤器 goroutine 与 ticker 泄漏。

### 修复

- 修复 zmexing `LoadDictPath` 文件打开失败时的空指针风险。
- 修复 `LoadDictEmbed` 绕过归一化的问题。
- 修复 AC `FindAllRanges` 按词去重导致重复出现只替换一处的问题。
- 移除对外部网络的测试依赖。

### 变更

- **BREAKING**：模块路径为 `github.com/cysnb/go-sensitive-word`，与两个上游均不同。
- 移除 `concurrent-map` 与 `testify`，保持零第三方依赖。
- 最低 Go 版本提升到 1.21。
