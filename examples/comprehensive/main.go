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
		sensitive.WithStrictNormalizer(),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	if err := f.LoadDictEmbedWithSource(sensitive.DictViolence, "violence"); err != nil {
		log.Fatal(err)
	}
	if err := f.Store.AddWord("秦始皇"); err != nil {
		log.Fatal(err)
	}
	if err := f.WaitForSync(context.Background()); err != nil {
		log.Fatal(err)
	}

	text := "秦始皇与毒品"
	fmt.Println("IsSensitive:", f.IsSensitive(text))
	fmt.Println("FindAllWithSource:", f.FindAllWithSource(text))
	fmt.Println("Replace:", f.Replace(text, '*'))
	fmt.Println("Remove:", f.Remove(text))
	fmt.Println("Stats:", f.GetStats().TotalWords)
	fmt.Println("Email:", sensitive.MaskEmail("user@example.com"))
}
