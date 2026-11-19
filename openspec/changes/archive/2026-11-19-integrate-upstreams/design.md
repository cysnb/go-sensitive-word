# Design

## Context

参见 `proposal.md - Why`。本设计基于以下已确认约束：

- 仓库当前为空骨架，无历史代码与 API 兼容包袱。
- 两个上游均已确认：`zmexing` 以词库与加载方式见长，`LuYongwang` 以算法实现、归一化与词库管理见长。
- 已确认采用单一超集库、zmexing 公开包风格、零第三方依赖、内置词库取 zmexing 16 库超集并并入 LuYongwang 独有词条。

## Goals / Non-Goals

**Goals:**

- 形成一套对外一致、可独立发布的 API，覆盖两方全部对外能力。
- 在零第三方依赖前提下实现并发安全、可关闭、可确定性等待更新的运行时。
- 词库数据可追溯：保留分类、去重、并入两方并集。

**Non-Goals:**

- 不实现 Redis/数据库等外部存储后端（仅保留回调加载扩展点）。
- 不做与两个上游的源码级 import 路径兼容；本变更是新模块。
- 不引入 Web 服务、CLI 或管理后台。
- 不在运行时从上游仓库自动同步词库。

## Decisions

### D1: 单一超集库 + zmexing 包布局

采用根包 + 公开 `filter/`、`store/`、`fetch/`，`normalize` 与 `jianfan` 置于 `internal/`。

备选：以 LuYongwang 的 `internal/filter`、`internal/store` 全收敛布局为主。放弃原因是用户选择 zmexing 风格，且 zmexing 用户可能直接引用 `store.Store`、`filter.Filter`。

目录：

```
go-sensitive-word/
├── go.mod
├── options.go
├── manager.go
├── normalize.go
├── wrapper.go
├── tool.go
├── fetch/fetch.go
├── filter/{filter,dfa,ac}.go
├── store/{store,memory}.go
├── internal/normalize/normalize.go
├── internal/jianfan/{jianfan,charsets}.go
├── wordlists/*.txt
├── examples/...
├── docs/...
├── Makefile
├── .golangci.yml
└── .github/workflows/go.yml
```

### D2: 零第三方依赖

用标准库 `sync.RWMutex` 实现内存词库，移除 `concurrent-map`；测试改用标准库断言，移除 `testify`。

备选：保留 `concurrent-map` 与 `testify` 以减少移植量。放弃原因是用户明确要求零依赖，且 LuYongwang 已证明标准库方案可行。

### D3: 算法内核采用 LuYongwang 版本

- DFA：采用 LuYongwang 实现，保留 `FindAllRanges` 扩展接口，删除其内部 `ToLower`。
- AC：采用 `atomic.Pointer` + copy-on-write + 100ms/1000 条窗口合并，保证读操作无锁、写操作不阻塞读。
- `FilterAc` 与 `FilterAC` 作为同一值的两个别名，兼顾两方调用习惯。

备选：采用 zmexing 的可变树 + `built` 标志。放弃原因是并发读写不安全，且与归一化包装叠加会重复转小写。

### D4: 归一化双侧注入

```
                NormalizerConfig
                       │
        ┌──────────────┴──────────────┐
        ▼                             ▼
store.WithWordNormalizer(fn)   newNormalizedFilter(rawFilter, cfg)
  词条写入前归一化                查询文本归一化 + 区间映射
```

- 根包暴露 `NormalizerConfig`、`DefaultNormalizer()`、`StrictNormalizer()`。
- `store` 通过 `store.Option` 接收 `func(string) string`，避免 `store` 反向依赖根包造成循环导入。
- `store.NewMemoryModel()` 无参调用时使用内部默认配置，保证直接使用者与 `NewFilter` 行为一致。
- 内核 filter 不做任何大小写/全角处理。

备选：仅在 Manager 层归一化词条。放弃原因是 `filter.Store.AddWord` 是公开入口，绕过 Manager 会导致词库与查询策略不一致。

### D5: Store 接口取两方并集，Manager 仅保留跨组件能力

`store.Store` 包含：加载（Embed/Path/HTTP/Reader/Callback）、读取（Chan/String）、增删（单个/批量/替换）、来源追踪、导出、统计、合并、清空、生命周期。

`Manager` 仅保留必须同时访问 Store 与 Filter 的方法：`FindAllWithSource`、`FindAllCountWithSource`、`ExportToFile`、`MergeFromManager`、`RefreshFromPath`、`Close`、`Shutdown`、`WaitForSync`。

备选：把所有 LuYongwang Manager 方法原样保留。放弃原因是与 zmexing 调用风格不一致，会让用户面对两套增删入口。

### D6: 更新可见性与同步点

Store 入队时递增 `pending`，Filter 应用完一条后回调递减；`WaitForSync(ctx)` 等待 `pending == 0`。

```
AddWord ──► store.map 写入 ──► addChan ──► filter.Apply ──► applied() 递减 pending
   │                                                              │
   └──────────────── WaitForSync 等待 pending==0 ◄────────────────┘
```

- 不改变 `AddWord` 的异步语义，避免 AC 窗口合并时阻塞调用方。
- 测试与示例在写入后调用 `WaitForSync`，替代上游的 `time.Sleep`。

备选：让 `AddWord` 同步返回。放弃原因是 AC 窗口合并最多需 100ms，会显著拖慢批量加载。

### D7: 关闭链路

`Manager.Close` 依次关闭 Filter 监听（停止 ticker/goroutine）与 Store 通道；`Shutdown(ctx)` 带超时。所有 Close 幂等。

备选：沿用 LuYongwang 仅关闭 Store。放弃原因是 Filter 的 ticker 与监听 goroutine 会泄漏。

### D8: 词库合并与去重

- 以 zmexing 的 16 个分类文件为基础。
- 每个文件按行 `TrimSpace`、去空行、按小写去重。
- 将 LuYongwang `色情词库` 中不在 zmexing 版本里的 35 个词并入。
- 结果一次性生成并提交到 `wordlists/`，运行时不再处理上游文件。
- 全量去重并集约 51,103 词；README 明确建议按需加载。

备选：运行时合并两方词库。放弃原因是增加运行时复杂度与不确定性，且词库属于数据而非逻辑。

### D9: 测试与工具链

- 单元测试使用表驱动 + 标准库，覆盖两方原有测试场景。
- HTTP 加载测试使用 `httptest.Server`，移除外部 URL。
- 增加 `go test -race` 的并发场景。
- Makefile 提供 `build/test/lint/bench`；CI 执行 `lint + test`。

## Risks / Trade-offs

- [异步可见性增加调用方心智负担] → 提供 `WaitForSync` 并在文档/示例中默认使用。
- [51k 词全量加载内存与延迟上升] → README 推荐按分类加载；基准测试使用精选子集。
- [归一化后区间映射在重复字符/零宽删除时易错] → 以区间标记实现 Replace/Remove，并补充针对性测试。
- [公开 `filter`/`store` 包增加 API 稳定性负担] → 仅在包内暴露必要类型，内部实现放 `internal`。
- [移除第三方依赖需重写测试] → 优先改写为表驱动，保留原测试语义。
- [词库内容合规风险] → 保留两方 README 中的声明与使用限制。
- [上游无明确许可证] → 见 Open Questions。

## Migration Plan

全新工程，无既有用户迁移。上线步骤：

1. 建立模块与目录骨架。
2. 移植内核与存储，补齐归一化与生命周期。
3. 生成并提交词库。
4. 合并文档、示例、测试与 CI。
5. 通过 `make lint test` 与 `go test -race` 后打 tag。

回滚：删除新增文件并回退到仅含 README 的初始提交。

## Open Questions

- 许可证选择：两个上游均未提供明确 LICENSE（LuYongwang 标注“待定”）。发布前需确认采用何种许可证；该问题不影响 specs 与任务拆分。
