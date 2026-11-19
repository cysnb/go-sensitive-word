# Tasks

## 1. 工程骨架

- [x] 1.1 初始化 `go.mod`：module `github.com/cysnb/go-sensitive-word`、`go 1.21`、根包名 `go_sensitive_word`；验证：`go mod tidy` 后 `go.mod` 无任何 `require`
- [x] 1.2 创建 `filter/`、`store/`、`fetch/`、`internal/normalize/`、`internal/jianfan/`、`wordlists/`、`docs/`、`examples/` 及占位包；验证：`go build ./...` 通过
- [x] 1.3 添加 `.gitignore`（忽略构建产物、覆盖率文件、临时目录）；验证：`git status --short` 不出现构建产物
- [x] 1.4 编写 `Makefile`（`build/test/lint/bench/help`）与 `.golangci.yml`；验证：`make help` 列出全部目标
- [x] 1.5 添加 `.github/workflows/go.yml`（lint + test）；验证：本地执行与 workflow 相同的 `go vet ./...` 与 `make test` 通过

## 2. 内置词库

- [x] 2.1 获取两个上游词库快照，按分类合并、去空行、按小写去重，生成 `wordlists/` 下 16 个文件；验证：每个文件非空、无空行、无重复，色情词库 587 词，全量并集约 51,103 词
- [x] 2.2 在 `options.go` 中定义 16 个 `go:embed` 词库常量；验证：编译通过且单元测试断言 16 个常量均非空
- [x] 2.3 编写 `docs/wordlists.md` 记录分类、上游来源与词数；验证：README 中的词库文档链接指向该文件且存在

## 3. 文本归一化

- [x] 3.1 移植 `internal/jianfan` 的繁简映射；验证：`T2S`/`S2T` 表驱动单元测试通过
- [x] 3.2 实现 `internal/normalize`（配置、单字符规则、索引映射）；验证：单测覆盖大小写、全角、零宽、数字、重复、繁简、英文变体、同形字
- [x] 3.3 在根包实现 `NormalizerConfig`、`DefaultNormalizer`、`StrictNormalizer`、`NormalizeWord`、`NormalizeTextWithMap`；验证：默认与严格配置的行为测试通过
- [x] 3.4 实现 `wrapper.go` 归一化包装与区间映射；验证：`FindOne`/`FindAll` 返回原文片段，`Replace`/`Remove` 仅作用于命中区间
- [x] 3.5 编写 `docs/normalization.md`；验证：文档中的示例代码可编译运行

## 4. 存储与加载

- [x] 4.1 定义 `store.Store` 并集接口（加载、增删、批量、来源、导出、统计、合并、生命周期）；验证：编译通过且 `MemoryModel` 实现接口的断言成立
- [x] 4.2 移植 `fetch` 包；验证：基于 `httptest` 的 GET、POST 与非 200 场景测试通过
- [x] 4.3 实现 `store/memory.go` 基础部分（`sync.RWMutex`、词计数、通道、可注入词归一化）；验证：添加/删除/读取单元测试通过
- [x] 4.4 实现五种加载方式（Embed/Path/Reader/Callback/HTTP）；验证：每种方式均有单测，文件不存在或回调出错时返回错误且不 panic
- [x] 4.5 实现批量增删、替换、清空、合并、导出与统计；验证：对应单元测试通过，统计为返回副本
- [x] 4.6 实现待处理计数与 `WaitForSync`；验证：写入后调用 `WaitForSync` 再查询立即可见，测试不依赖 `time.Sleep`
- [x] 4.7 实现幂等 `Close`/`Shutdown` 并保证关闭后操作返回错误；验证：生命周期单测通过，`runtime.NumGoroutine` 在关闭后无残留增长
- [x] 4.8 编写 `docs/word-loading.md`、`docs/word-management.md`、`docs/lifecycle.md`；验证：文档中的命令与示例按原样可运行

## 5. 过滤内核

- [x] 5.1 定义 `filter.Filter`、`RangedFilter`、`MatchResult`；验证：接口实现断言编译通过
- [x] 5.2 移植 DFA（去除内部转小写，保留 `FindAllRanges`）并实现可停止的监听；验证：检测、替换、删除与区间单测通过
- [x] 5.3 移植 AC（原子切换 + COW + 窗口合并）并实现 `Close`；验证：检测、替换、删除与区间单测通过
- [x] 5.4 验证 DFA/AC 结果等价与并发安全；验证：等价表驱动测试与 `go test -race ./filter/...` 通过
- [x] 5.5 添加基准测试；验证：`make bench` 可运行并产出两种算法数据
- [x] 5.6 编写 `docs/algorithm-guide.md`、`docs/ac.md`、`docs/dfa.md`；验证：文档中给出的基准命令可执行

## 6. Manager 与公共 API

- [x] 6.1 实现 `NewFilter`：组装 Store、内核 Filter、归一化包装与监听；验证：DFA 与 AC 两种初始化测试通过
- [x] 6.2 在 `options.go` 提供 `FilterAc` 与 `FilterAC` 同值别名、`StoreOption`/`FilterOption` 与可选归一化参数；验证：两别名比较相等，原有双参数调用编译通过
- [x] 6.3 实现 Manager 跨组件方法（`FindAllWithSource`、`FindAllCountWithSource`、`ExportToFile`、`MergeFromManager`、`RefreshFromPath`、`WaitForSync`、`Close`、`Shutdown`）；验证：各有单测覆盖
- [x] 6.4 实现 `tool.go` 邮箱/URL/数字/微信号检测与脱敏；验证：工具函数表驱动测试通过
- [x] 6.5 编写 `docs/api-reference.md`、`docs/word-source-tracking.md`、`docs/tools.md`；验证：README 导航链接均存在

## 7. 文档与示例

- [x] 7.1 合并两方 README 生成根 `README.md`，包含“主要上游工程”章节与贡献说明；验证：章节覆盖简介、特性、安装、快速开始、API、词库、文档、示例、性能、声明
- [x] 7.2 合并其余文档（快速开始、最佳实践、FAQ、unicode、零宽字符、更多案例、基准）；验证：README 中所有 `docs/` 链接均指向存在的文件
- [x] 7.3 合并 `examples/` 示例并统一导入路径；验证：`go build ./examples/...` 通过，至少 `examples/basic` 可运行并输出预期结果
- [x] 7.4 编写 `CHANGELOG.md`，记录整合自两个上游的能力与修复项；验证：版本条目与 README 特性一致

## 8. 集成验证

- [x] 8.1 运行 `make lint`；验证：无错误输出
- [x] 8.2 运行 `make test` 与 `go test -race ./...`；验证：全部通过
- [x] 8.3 运行 `make bench`；验证：DFA 与 AC 基准均产出结果
- [x] 8.4 验证零第三方依赖与包可见性；验证：`go list -deps ./...` 无外部模块，外部模块导入 `internal/normalize` 编译失败
- [x] 8.5 按 README 快速开始示例新建临时程序并运行；验证：DFA 与 AC 两种示例均输出预期检测结果

## Workflow follow-up

- 在评审要求满足后归档该变更。
- 验证归档后的主 spec 与变更目录状态。
