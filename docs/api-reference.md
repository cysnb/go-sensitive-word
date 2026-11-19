# API 参考

## 初始化

```go
func NewFilter(storeOption StoreOption, filterOption FilterOption, opts ...Option) (*Manager, error)
```

| 选项 | 值 |
|------|----|
| `StoreOption{Type}` | `StoreMemory` |
| `FilterOption{Type}` | `FilterDfa`、`FilterAc`（`FilterAC` 等价） |
| 可选参数 | `WithNormalizer(cfg)`、`WithStrictNormalizer()` |

## 文本检测

通过 `Manager`（嵌入 `filter.Filter`）调用：

| 方法 | 说明 |
|------|------|
| `IsSensitive(text) bool` | 是否存在敏感词 |
| `FindOne(text) string` | 第一个敏感词 |
| `FindAll(text) []string` | 所有敏感词（去重） |
| `FindAllCount(text) map[string]int` | 所有敏感词及次数 |
| `Replace(text, repl) string` | 替换敏感词 |
| `Remove(text) string` | 删除敏感词 |

## 词库操作

通过 `Manager.Store` 调用：

| 方法 | 说明 |
|------|------|
| `LoadDictEmbed(...string)` | 加载内置词库 |
| `LoadDictPath(...string)` | 加载文件 |
| `LoadDictHttp(...string)` | 加载远程词库 |
| `LoadDict(io.Reader)` | 从 Reader 加载 |
| `LoadDictCallback(loader, source)` | 回调加载 |
| `LoadDictEmbedWithSource(content, source)` | 加载内置词库并记录来源 |
| `AddWord(...string)` / `DelWord(...string)` | 动态增删 |
| `AddWords([]string)` / `DelWords([]string)` | 批量增删 |
| `ReplaceWords(old, new []string)` | 批量替换 |
| `AddWordsWithSource(words, source)` | 带来源添加 |
| `GetWordSources(word) []string` | 查询词来源 |
| `GetAllWordSources() map[string][]string` | 全部来源 |
| `ReadString() []string` / `ReadChan() <-chan string` | 读取词库 |
| `ExportToString() (string, error)` / `ExportToWriter(io.Writer)` | 导出 |
| `GetStats() Stats` | 统计信息 |
| `Clear()` / `Merge(other)` | 清空 / 合并 |
| `WaitForSync(ctx) error` | 等待变更应用 |

## Manager 扩展

| 方法 | 说明 |
|------|------|
| `FindAllWithSource(text) []filter.MatchResult` | 命中词及来源 |
| `FindAllCountWithSource(text) map[string]filter.MatchResult` | 次数与来源 |
| `ExportToFile(path) error` | 导出到文件 |
| `MergeFromManager(other)` | 合并另一个 Manager |
| `RefreshFromPath(path, replace)` | 从文件刷新 |
| `Close() error` / `Shutdown(ctx) error` | 生命周期 |

## 类型

```go
type Stats struct {
    TotalWords  int
    LastUpdate  time.Time
    UpdateCount int
    Source      []string
}

type MatchResult struct {
    Word   string
    Source []string
    Count  int
}

type NormalizerConfig struct {
    IgnoreCase         bool
    ToHalfWidth        bool
    IgnoreRepeat       bool
    IgnoreDigitType    bool
    IgnoreSimpTrad     bool
    IgnoreEnglishStyle bool
    RemoveZeroWidth    bool
    HomoglyphMap       map[rune]rune
}
```
