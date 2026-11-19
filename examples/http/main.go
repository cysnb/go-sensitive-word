package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"

	sensitive "github.com/cysnb/go-sensitive-word"
)

func main() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("远程敏感词\n"))
	}))
	defer srv.Close()

	f, err := sensitive.NewFilter(
		sensitive.StoreOption{Type: sensitive.StoreMemory},
		sensitive.FilterOption{Type: sensitive.FilterDfa},
	)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	if err := f.LoadDictHttp(srv.URL); err != nil {
		log.Fatal(err)
	}
	if err := f.WaitForSync(context.Background()); err != nil {
		log.Fatal(err)
	}
	fmt.Println("命中:", f.IsSensitive("包含远程敏感词"))
	fmt.Println("来源:", f.GetStats().Source)
}
