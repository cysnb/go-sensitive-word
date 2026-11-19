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

	if err := f.Store.AddWordsWithSource([]string{"违禁词A"}, "custom"); err != nil {
		log.Fatal(err)
	}
	if err := f.Store.AddWordsWithSource([]string{"违禁词A"}, "political"); err != nil {
		log.Fatal(err)
	}
	if err := f.LoadDictEmbedWithSource(sensitive.DictViolence, "violence"); err != nil {
		log.Fatal(err)
	}
	if err := f.WaitForSync(context.Background()); err != nil {
		log.Fatal(err)
	}

	fmt.Println("来源:", f.GetWordSources("违禁词A"))
	for _, r := range f.FindAllWithSource("违禁词A") {
		fmt.Println(r.Word, r.Source)
	}
	counts := f.FindAllCountWithSource("违禁词A违禁词A")
	fmt.Println("次数:", counts["违禁词A"].Count)
}
