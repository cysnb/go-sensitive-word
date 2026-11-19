# Spec Delta

## Purpose

定义模块身份、依赖边界、包可见性、上游声明与交付物组织，确保该库可独立发布、可追溯来源且长期可维护。

## ADDED Requirements

### Requirement: 模块身份
系统 SHALL 使用模块路径 `github.com/cysnb/go-sensitive-word`，根包名 MUST 为 `go_sensitive_word`，最低 Go 版本 MUST 为 1.21。

#### Scenario: 导入根包
- **WHEN** 外部工程导入 `github.com/cysnb/go-sensitive-word`
- **THEN** 可使用 `NewFilter`、选项类型与词库常量

#### Scenario: Go 版本声明
- **WHEN** 检查 `go.mod`
- **THEN** 其中声明的 Go 版本不低于 1.21

### Requirement: 零第三方依赖
系统的 `go.mod` MUST 不包含任何第三方 `require`，构建与测试仅依赖 Go 标准库。

#### Scenario: 依赖检查
- **WHEN** 执行依赖整理后检查 `go.mod`
- **THEN** 不存在除本模块外的任何依赖项

### Requirement: 包可见性
系统 MUST 公开 `filter`、`store`、`fetch` 包，并 MUST 将 `normalize`、`jianfan` 置于 `internal` 下，外部模块不得直接导入。

#### Scenario: 使用公开包
- **WHEN** 外部工程导入 `filter` 或 `store`
- **THEN** 编译成功并可引用其中的导出类型

#### Scenario: 导入内部包
- **WHEN** 外部工程尝试导入 `internal/normalize` 或 `internal/jianfan`
- **THEN** 编译失败

### Requirement: 上游声明
README MUST 明确标注 `zmexing/go-sensitive-word` 与 `LuYongwang/go-sensitive-word` 为本项目的主要上游工程，并分别说明其贡献。

#### Scenario: 上游章节
- **WHEN** 阅读 README
- **THEN** 可找到专门章节，包含两个上游仓库地址及其主要贡献列表

#### Scenario: README 结构
- **WHEN** 阅读 README
- **THEN** 至少覆盖项目简介、特性、安装、快速开始、API、内置词库、文档导航、示例、性能与声明

### Requirement: 文档与示例
系统 SHALL 提供 `docs/` 文档与 `examples/` 可运行示例，示例 MUST 使用本模块的导入路径并可编译。

#### Scenario: 示例编译
- **WHEN** 执行构建全部包的命令
- **THEN** `examples/` 下所有示例编译通过

#### Scenario: 文档导航
- **WHEN** 阅读 README 的文档导航
- **THEN** 所有列出的 `docs/` 文件均存在

### Requirement: 构建与质量工具
系统 SHALL 提供 Makefile 的构建、测试、静态检查与基准测试目标，并 SHALL 提供持续集成配置。

#### Scenario: 运行测试目标
- **WHEN** 执行 Makefile 的测试目标
- **THEN** 全部单元测试执行并通过

#### Scenario: 静态检查目标
- **WHEN** 执行 Makefile 的静态检查目标
- **THEN** 代码通过已配置的检查规则
