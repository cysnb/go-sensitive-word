# 更多案例

## 嵌套与重叠词

词库同时包含 `毒品` 与 `毒品销售` 时，`FindAll` 返回去重后的命中词，`Replace`/`Remove` 按所有命中区间处理。

```go
f.Store.AddWord("毒品", "毒品销售")
f.WaitForSync(ctx)
fmt.Println(f.FindAll("毒品销售")) // [毒品 毒品销售]
```

## 同一词多次出现

```go
fmt.Println(f.FindAllCount("毒品毒品毒品")) // map[毒品:3]
```

## 中英混合

```go
fmt.Println(f.Replace("fuCK 毒品", '*')) // **** **
```

## 动态白名单

```go
f.Store.DelWord("允许的词")
f.WaitForSync(ctx)
```

## 多实例隔离

```go
a, _ := sensitive.NewFilter(storeOpt, dfaOpt)
b, _ := sensitive.NewFilter(storeOpt, acOpt)
```

两个实例的词库与算法互不影响。

## 从数据库加载

```go
f.LoadDictCallback(func() ([]string, error) {
    return queryWordsFromDB()
}, "database")
```
