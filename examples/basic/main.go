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

	if err := f.LoadDictEmbed(sensitive.DictPolitical, sensitive.DictViolence); err != nil {
		log.Fatal(err)
	}
	if err := f.Store.AddWord("秦始皇", "李世民"); err != nil {
		log.Fatal(err)
	}
	if err := f.WaitForSync(context.Background()); err != nil {
		log.Fatal(err)
	}

	text := "秦始皇和李世民是忘年交"
	fmt.Println("IsSensitive:", f.IsSensitive(text))
	fmt.Println("FindOne:", f.FindOne(text))
	fmt.Println("FindAll:", f.FindAll(text))
	fmt.Println("FindAllCount:", f.FindAllCount(text))
	fmt.Println("Replace:", f.Replace(text, '*'))
	fmt.Println("Remove:", f.Remove(text))
}
