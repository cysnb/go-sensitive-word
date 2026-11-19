# 生命周期

过滤器会启动监听 goroutine；AC 算法还会启动窗口合并 ticker。使用完毕后应关闭以释放资源。

## Close

```go
f, _ := sensitive.NewFilter(
    sensitive.StoreOption{Type: sensitive.StoreMemory},
    sensitive.FilterOption{Type: sensitive.FilterAc},
)
defer f.Close()
```

`Close` 幂等，重复调用安全。关闭后词库操作返回 `store.ErrStoreClosed`。

## Shutdown

需要超时控制时使用：

```go
ctx, cancel := context.WithTimeout(context.Background(), time.Second)
defer cancel()
if err := f.Shutdown(ctx); err != nil {
    log.Println(err)
}
```

## 关闭顺序

`Manager.Close` 先停止过滤器的监听与后台 goroutine，再关闭词库。直接使用 `store.MemoryModel` 时：

```go
m := store.NewMemoryModel()
defer m.Close()
```

关闭后不应再调用 `WaitForSync`，此时会返回 `ErrStoreClosed`。
