# 快速开始

## 安装

```bash
go get github.com/cysnb/go-sensitive-word
```

## 创建过滤器

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

    if err := f.LoadDictEmbed(sensitive.DictPolitical, sensitive.DictViolence); err != nil {
        log.Fatal(err)
    }
    if err := f.Store.AddWord("秦始皇"); err != nil {
        log.Fatal(err)
    }
    if err := f.WaitForSync(context.Background()); err != nil {
        log.Fatal(err)
    }

    text := "秦始皇统一六国"
    fmt.Println(f.IsSensitive(text)) // true
    fmt.Println(f.FindAll(text))     // [秦始皇]
    fmt.Println(f.Replace(text, '*')) // ***统一六国
}
```

## 切换到 AC

```go
f, _ := sensitive.NewFilter(
    sensitive.StoreOption{Type: sensitive.StoreMemory},
    sensitive.FilterOption{Type: sensitive.FilterAc},
)
```

## 下一步

- [词库加载](word-loading.md)
- [词库管理](word-management.md)
- [归一化](normalization.md)
- [来源追踪](word-source-tracking.md)
