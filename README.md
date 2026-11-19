# go-sensitive-word

敏感词（敏感词/违禁词/违法词/脏词）检测工具，基于 DFA / AC 算法实现的高性能 Go 敏感词过滤框架。

> 本项目整合并维护自以下两个**主要上游工程**，在两者最新版本的基础上取长补短：
>
> - [zmexing/go-sensitive-word](https://github.com/zmexing/go-sensitive-word)
> - [LuYongwang/go-sensitive-word](https://github.com/LuYongwang/go-sensitive-word)
>
> 上游贡献说明见 [主要上游工程](#主要上游工程)。

## ✨ 特性

- 🚀 **双算法**：支持 DFA 与 AC 自动机，AC 采用原子切换 + Copy-on-Write + 窗口合并
- 🔄 **动态维护**：运行时增删、批量替换、清空、合并，无需重启
- 🛡️ **智能归一化**：默认忽略大小写与全角；严格模式支持零宽、繁简、同形字、数字与英文变体
- 📦 **多种加载**：内置词库、文件、HTTP、Reader、回调（数据库/Redis 等）
- 🔍 **来源追踪**：记录词条所属词库，并在匹配结果中返回来源
- 🛠️ **字符串工具**：邮箱、URL、数字、微信号检测与脱敏
- 🔒 **并发安全**：全链路并发安全，支持高并发查询
- ♻️ **资源管理**：幂等关闭，避免 goroutine 与 ticker 泄漏
- 📚 **零第三方依赖**：仅使用 Go 标准库

## 安装

```bash
go get github.com/cysnb/go-sensitive-word
```

要求 Go 1.21+。

## 快速开始

### DFA

```go
package main

import (
    "context"
    "fmt"
    "log"

    sensitive "github.com/cysnb/go-sensitive-word"
)

func main() {
    f, err := sensitive.NewFilter(
        sensitive.StoreOption{Type: sensitive.StoreMemory},
        sensitive.FilterOption{Type: sensitive.FilterDfa},
    )
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()

    // 加载内置词库
    if err := f.LoadDictEmbed(
        sensitive.DictPolitical,
        sensitive.DictViolence,
        sensitive.DictPornography,
    ); err != nil {
        log.Fatal(err)
    }

    // 动态添加自定义词
    if err := f.Store.AddWord("秦始皇", "李世民"); err != nil {
        log.Fatal(err)
    }

    // 等待词库变更应用
    if err := f.WaitForSync(context.Background()); err != nil {
        log.Fatal(err)
    }

    text := "秦始皇和李世民是忘年交"

    fmt.Println(f.IsSensitive(text))     // true
    fmt.Println(f.FindOne(text))         // 秦始皇
    fmt.Println(f.FindAll(text))         // [秦始皇 李世民]
    fmt.Println(f.FindAllCount(text))    // map[秦始皇:1 李世民:1]
    fmt.Println(f.Replace(text, '*'))    // ***和***是忘年交
    fmt.Println(f.Remove(text))          // 和是忘年交
}
```

### AC

```go
f, err := sensitive.NewFilter(
    sensitive.StoreOption{Type: sensitive.StoreMemory},
    sensitive.FilterOption{Type: sensitive.FilterAc}, // FilterAC 等价
)
```

## 核心功能

### 文本检测

| 方法 | 说明 |
|------|------|
| `IsSensitive()` | 判断文本中是否存在敏感词 |
| `FindOne()` | 查找第一个敏感词 |
| `FindAll()` | 查找所有敏感词（去重） |
| `FindAllCount()` | 查找所有敏感词及出现次数 |
| `Replace()` | 替换所有敏感词 |
| `Remove()` | 删除所有敏感词 |

### 词库管理

通过 `f.Store` 调用：

- `AddWord()` / `AddWords()` / `DelWord()` / `DelWords()`
- `ReplaceWords()` / `Clear()` / `Merge()`
- `GetStats()` / `ExportToString()` / `ExportToWriter()`
- `WaitForSync()` 等待变更应用

### 词库加载

- `LoadDictEmbed()`：内置词库
- `LoadDictPath()`：本地文件
- `LoadDictHttp()`：远程 HTTP
- `LoadDict()`：任意 `io.Reader`
- `LoadDictCallback()`：数据库、Redis、配置中心等
- `LoadDictEmbedWithSource()`：带来源加载

### 归一化

默认启用忽略大小写与全角转半角；严格模式额外支持：

- 连续重复字符压缩
- 数字写法归一
- 繁简归一
- 英文变体归一
- 零宽字符剔除
- 同形字映射

```go
f, _ := sensitive.NewFilter(storeOpt, dfaOpt, sensitive.WithStrictNormalizer())
```

详见 [归一化文档](docs/normalization.md)。

### 来源追踪

```go
f.Store.AddWordsWithSource([]string{"违禁词A"}, "custom")
f.WaitForSync(ctx)

fmt.Println(f.GetWordSources("违禁词A")) // [custom]
for _, r := range f.FindAllWithSource("包含违禁词A") {
    fmt.Println(r.Word, r.Source)
}
```

详见 [来源追踪文档](docs/word-source-tracking.md)。

### 字符串工具

`HasEmail` / `MaskEmail`、`HasURL` / `MaskURL`、`HasDigit` / `MaskDigit`、`HasWechatID` / `MaskWechatID`。

### 生命周期

```go
defer f.Close()

ctx, cancel := context.WithTimeout(context.Background(), time.Second)
defer cancel()
_ = f.Shutdown(ctx)
```

## 内置词库

内置 16 个分类词库，全量去重并集约 **51,103** 词。完整列表与词数见 [词库文档](docs/wordlists.md)。

常用常量：

```go
sensitive.DictPolitical
sensitive.DictViolence
sensitive.DictPornography
sensitive.DictCorruption
sensitive.DictReactionary
sensitive.DictGFWAdditional
sensitive.DictTemporaryTencent
sensitive.DictIllegalURL
```

## 主要上游工程

本项目为两个上游工程的整合版，README 结构综合自两方 README。

### zmexing/go-sensitive-word

贡献：

- 16 个分类词库（COVID-19、GFW 补充、其他、反动、广告、政治、暴恐、民生、涉枪涉爆、网易前端、色情类型、色情、补充、贪腐、零时-Tencent、非法网址）
- 公开 `filter` / `store` 包与 `Store`/`Filter` 嵌入风格 API
- HTTP 远程词库加载与 `fetch` 包
- `FilterAc` 命名与十六个 `DictXxx` 常量

### LuYongwang/go-sensitive-word

贡献：

- 并发安全的 AC 自动机（原子切换 + Copy-on-Write + 窗口合并）
- DFA `FindAllRanges` 区间接口
- 文本归一化与繁简转换
- 词库来源追踪
- 批量维护、统计、导出、合并、回调加载与生命周期管理
- 完善的示例与文档体系

## 文档导航

- [快速开始](docs/getting-started.md)
- [API 参考](docs/api-reference.md)
- [归一化](docs/normalization.md)
- [算法选择](docs/algorithm-guide.md)
- [DFA](docs/dfa.md)
- [AC 自动机](docs/ac.md)
- [词库管理](docs/word-management.md)
- [词库加载](docs/word-loading.md)
- [内置词库](docs/wordlists.md)
- [来源追踪](docs/word-source-tracking.md)
- [字符串工具](docs/tools.md)
- [生命周期](docs/lifecycle.md)
- [最佳实践](docs/best-practices.md)
- [性能基准](docs/benchmark.md)
- [Unicode 相似字符](docs/unicode.md)
- [零宽字符](docs/zero-width.md)
- [更多案例](docs/more-case.md)
- [常见问题](docs/faq.md)

## 示例代码

| 示例 | 说明 |
|------|------|
| [examples/basic](examples/basic/main.go) | 基础功能 |
| [examples/ac](examples/ac/main.go) | AC 算法 |
| [examples/dynamic](examples/dynamic/main.go) | 动态维护 |
| [examples/callback](examples/callback/main.go) | 回调加载 |
| [examples/file-load](examples/file-load/main.go) | 文件加载 |
| [examples/http](examples/http/main.go) | HTTP 加载 |
| [examples/normalize](examples/normalize/main.go) | 归一化 |
| [examples/tools](examples/tools/main.go) | 字符串工具 |
| [examples/lifecycle](examples/lifecycle/main.go) | 生命周期 |
| [examples/multi-instance](examples/multi-instance/main.go) | 多实例 |
| [examples/word-source](examples/word-source/main.go) | 来源追踪 |
| [examples/comprehensive](examples/comprehensive/main.go) | 综合演示 |

运行：

```bash
go run ./examples/basic
```

## 性能

环境：Intel Core i9-9980HK，词库 1000 词。

| 操作 | DFA | AC |
|------|-----|-----|
| IsSensitive | 558 ns/op | 187 ns/op |
| FindAll | 1510 ns/op | 839 ns/op |
| IsSensitive（并发） | 213 ns/op | 48 ns/op |

```bash
make bench
```

详见 [性能基准](docs/benchmark.md)。

## 参考资料

- 基于 Java DFA 实现的敏感词过滤：https://github.com/houbb/sensitive-word
- 敏感词库：https://github.com/konsheng/Sensitive-lexicon
- unicode 字词的神奇组合：https://www.zhihu.com/question/30873035
- unicode 违规技巧：https://zhuanlan.zhihu.com/p/545309061
- unicode 视觉欺骗：https://zhuanlan.zhihu.com/p/611904676
- unicode 字符列表：https://symbl.cc/en/unicode-table

## 声明

本项目包含敏感词库，其设计目的是帮助开发者在互联网产品中过滤和屏蔽不适当或敏感的内容，营造健康、安全的网络环境。

请注意：

1. **项目目的**：为开发者提供内容过滤工具，不用于传播或扩散敏感信息。
2. **使用限制**：内置词库仅供技术研究与内容过滤功能使用。
3. **责任声明**：使用本项目产生的后果由使用者自行承担。
4. **使用规范**：请遵守相关法律法规，禁止用于违法或不正当用途。

通过下载和使用本项目，即表示您同意并接受上述声明。
