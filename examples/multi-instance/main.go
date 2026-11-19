package main

import (
	"context"
	"fmt"
	"log"

	sensitive "github.com/cysnb/go-sensitive-word"
)

func main() {
	a, err := sensitive.NewFilter(
		sensitive.StoreOption{Type: sensitive.StoreMemory},
		sensitive.FilterOption{Type: sensitive.FilterDfa},
	)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = a.Close() }()

	b, err := sensitive.NewFilter(
		sensitive.StoreOption{Type: sensitive.StoreMemory},
		sensitive.FilterOption{Type: sensitive.FilterAc},
	)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = b.Close() }()

	if err := a.Store.AddWord("实例A的词"); err != nil {
		log.Fatal(err)
	}
	if err := b.Store.AddWord("实例B的词"); err != nil {
		log.Fatal(err)
	}
	if err := a.WaitForSync(context.Background()); err != nil {
		log.Fatal(err)
	}
	if err := b.WaitForSync(context.Background()); err != nil {
		log.Fatal(err)
	}

	fmt.Println("A 命中 A 词:", a.IsSensitive("实例A的词"))
	fmt.Println("A 命中 B 词:", a.IsSensitive("实例B的词"))
	fmt.Println("B 命中 B 词:", b.IsSensitive("实例B的词"))
}
