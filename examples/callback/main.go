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
	defer func() { _ = f.Close() }()

	err = f.LoadDictCallback(func() ([]string, error) {
		return []string{"来自数据库的词"}, nil
	}, "database")
	if err != nil {
		log.Fatal(err)
	}
	if err := f.WaitForSync(context.Background()); err != nil {
		log.Fatal(err)
	}

	fmt.Println("命中:", f.IsSensitive("这是来自数据库的词"))
	fmt.Println("来源:", f.GetStats().Source)
}
