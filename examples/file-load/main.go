package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	sensitive "github.com/cysnb/go-sensitive-word"
)

func main() {
	dir, err := os.MkdirTemp("", "gsw")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(dir) }()

	path := filepath.Join(dir, "dict.txt")
	if err := os.WriteFile(path, []byte("文件敏感词\n"), 0o600); err != nil {
		log.Fatal(err)
	}

	f, err := sensitive.NewFilter(
		sensitive.StoreOption{Type: sensitive.StoreMemory},
		sensitive.FilterOption{Type: sensitive.FilterDfa},
	)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	if err := f.LoadDictPath(path); err != nil {
		log.Fatal(err)
	}
	if err := f.WaitForSync(context.Background()); err != nil {
		log.Fatal(err)
	}
	fmt.Println("命中:", f.IsSensitive("包含文件敏感词的文本"))
}
