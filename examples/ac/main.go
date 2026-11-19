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
		sensitive.FilterOption{Type: sensitive.FilterAc},
	)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	if err := f.LoadDictEmbed(sensitive.DictViolence, sensitive.DictPornography); err != nil {
		log.Fatal(err)
	}
	if err := f.Store.AddWord("毒品"); err != nil {
		log.Fatal(err)
	}
	if err := f.WaitForSync(context.Background()); err != nil {
		log.Fatal(err)
	}

	text := "他参与毒品交易"
	fmt.Println("AC FindAll:", f.FindAll(text))
	fmt.Println("AC Replace:", f.Replace(text, '*'))
}
