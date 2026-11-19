# Proposal

## Why

当前仓库只有一个占位 README，而两个主要上游工程 `zmexing/go-sensitive-word` 与 `LuYongwang/go-sensitive-word` 已明显分叉：前者拥有更完整的词库与 HTTP 加载方式，后者拥有并发安全的 AC 算法、文本归一化、来源追踪与完善的词库管理。使用方被迫在“词库全”和“能力强”之间二选一。本变更将两者最新版本整合为一个可直接依赖的 Go 库。

## What Changes

- 建立模块 `github.com/cysnb/go-sensitive-word`，要求 Go 1.21，保持**零第三方依赖**。
- 采用 zmexing 风格对外 API：公开 `filter`、`store`、`fetch` 包，`Manager` 嵌入 `Store` 与 `Filter`，用户通过 `filter.Store.AddWord` 等入口操作词库。
- 同时提供 DFA 与 AC 两种算法；AC 采用 LuYongwang 的 copy-on-write + 原子切换实现；`FilterAc` 与 `FilterAC` 两个常量名同时可用。
- 合并内置词库为 16 个分类：以 zmexing 词库为超集逐库去重，并并入 LuYongwang 独有的 35 个色情词。
- 默认启用 LuYongwang 的归一化配置（忽略大小写 + 全角转半角），额外提供 Strict 配置；词库写入与文本查询使用同一套归一化策略。
- 保留 LuYongwang 的增强能力并下沉为 `store.Store` 接口方法：批量增删、替换、统计、导出、合并、回调加载、来源追踪、优雅关闭。
- 保留两方共有的字符串检测/脱敏工具函数。
- 合并两方文档与示例，README 按两方 README 组织，并明确标注 zmexing 与 LuYongwang 为主要上游工程。
- 修复上游已知问题：`LoadDictPath` nil 解引用、过滤器监听协程泄漏、`LoadDictEmbed` 绕过归一化、外部网络测试。
- **BREAKING**：import 路径与两个上游均不同；移除 `concurrent-map` 与 `testify` 依赖，相关测试改用标准库。

## Capabilities

### New Capabilities

- `sensitive-word-filter`: 敏感词检测、查找、替换、删除，以及 DFA/AC 算法选择与并发安全行为。
- `dictionary-management`: 词库加载、动态维护、批量操作、统计、导出、合并与生命周期管理。
- `builtin-dictionaries`: 16 个内置分类词库的嵌入、常量暴露与去重后的词条规模。
- `text-normalization`: 文本与词条的归一化策略，以及归一化对匹配/替换/删除结果的影响。
- `word-source-tracking`: 词库来源标识的记录、查询，以及带来源的匹配结果。
- `string-tools`: 邮箱、URL、数字、微信号的检测与脱敏工具函数。
- `project-distribution`: 模块路径、零第三方依赖、README 上游声明、文档与示例的组织要求。

### Modified Capabilities

无。仓库当前没有任何既有能力。

## Impact

- 代码：全新工程，包含根包、`filter/`、`store/`、`fetch/`、`internal/normalize/`、`internal/jianfan/`、`wordlists/`。
- API：根包导出 `NewFilter`、`StoreOption`、`FilterOption`、`NormalizerConfig`、16 个 `DictXxx` 常量与字符串工具函数。
- 依赖：仅 Go 标准库。
- 文档：新增合并版 `README.md`、`docs/` 与 `examples/`。
- 构建：新增 `Makefile`、`.golangci.yml`、`.github/workflows/go.yml`。
- 测试：合并两方测试，改为表驱动 + 标准库断言，移除外网依赖。
