# 词库加载

词库支持五种加载方式，全部通过 `filter.Store` 调用。

## 内置词库

```go
err := f.LoadDictEmbed(
    sensitive.DictPolitical,
    sensitive.DictViolence,
    sensitive.DictPornography,
)
```

## 本地文件

```go
err := f.LoadDictPath("dict/a.txt", "dict/b.txt")
```

文件不存在或无权限时返回错误，不会 panic。

## 任意 Reader

```go
f, _ := os.Open("dict.txt")
defer f.Close()
err := f.LoadDict(f)
```

## HTTP 远程加载

```go
err := f.LoadDictHttp(
    "https://example.com/dict.txt",
    "https://example.com/dict2.txt",
)
```

非 200 响应返回包含状态码的错误。

## 回调加载

适合从数据库、Redis、配置中心读取：

```go
err := f.LoadDictCallback(func() ([]string, error) {
    return loadFromDB()
}, "database")
```

来源标识会记录到统计信息中：

```go
fmt.Println(f.GetStats().Source) // [callback://database]
```

## 带来源加载

```go
err := f.LoadDictEmbedWithSource(sensitive.DictViolence, "violence")
```

之后可用 `GetWordSources` 查询某个词的来源。
