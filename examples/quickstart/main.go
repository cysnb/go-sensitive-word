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

	if err := f.LoadDictEmbed(sensitive.DictPolitical); err != nil {
		log.Fatal(err)
	}
	if err := f.Store.AddWord("秦始皇"); err != nil {
		log.Fatal(err)
	}
	if err := f.WaitForSync(context.Background()); err != nil {
		log.Fatal(err)
	}
	fmt.Println(f.IsSensitive("秦始皇"))
	fmt.Println(f.FindAll("秦始皇"))
}
