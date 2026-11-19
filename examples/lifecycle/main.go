package main

import (
	"context"
	"fmt"
	"log"
	"time"

	sensitive "github.com/cysnb/go-sensitive-word"
)

func main() {
	f, err := sensitive.NewFilter(
		sensitive.StoreOption{Type: sensitive.StoreMemory},
		sensitive.FilterOption{Type: sensitive.FilterAc},
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := f.LoadDictEmbed(sensitive.DictViolence); err != nil {
		log.Fatal(err)
	}
	if err := f.WaitForSync(context.Background()); err != nil {
		log.Fatal(err)
	}
	fmt.Println("命中:", f.IsSensitive("毒品"))

	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("已关闭")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := f.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}
