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
		sensitive.WithStrictNormalizer(),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	if err := f.Store.AddWord("fuck"); err != nil {
		log.Fatal(err)
	}
	if err := f.WaitForSync(context.Background()); err != nil {
		log.Fatal(err)
	}

	fmt.Println("全角:", f.FindOne("ｆｕｃｋ"))
	fmt.Println("零宽:", f.FindOne("f\u200Buck"))
	fmt.Println("替换:", f.Replace("ＦＵＣＫ!", '*'))
	fmt.Println("删除:", f.Remove("f\u200Buck!"))
}
