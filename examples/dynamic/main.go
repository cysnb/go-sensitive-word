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

	ctx := context.Background()
	if err := f.Store.AddWords([]string{"旧词", "保留词"}); err != nil {
		log.Fatal(err)
	}
	if err := f.Store.ReplaceWords([]string{"旧词"}, []string{"新词"}); err != nil {
		log.Fatal(err)
	}
	if err := f.WaitForSync(ctx); err != nil {
		log.Fatal(err)
	}
	fmt.Println("新词:", f.IsSensitive("新词"))
	fmt.Println("旧词:", f.IsSensitive("旧词"))

	if err := f.Store.DelWord("保留词"); err != nil {
		log.Fatal(err)
	}
	if err := f.WaitForSync(ctx); err != nil {
		log.Fatal(err)
	}
	stats := f.GetStats()
	fmt.Println("TotalWords:", stats.TotalWords)
	fmt.Println("UpdateCount:", stats.UpdateCount)
}
