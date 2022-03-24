package main

import (
	"fmt"
	"net/http"

	"github.com/Code-Hex/dd"
	"github.com/Code-Hex/dd/p"
	"github.com/alecthomas/chroma/styles"
)

func main() {
	srv := &http.Server{
		Addr:    ":8080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
	}
	fmt.Println("--- monokai")
	p.New(
		p.WithDumpOptions(dd.WithExportedOnly()),
	).P(srv)

	fmt.Println("--- doom-one")
	p.New(
		p.WithStyle(styles.DoomOne),
		p.WithDumpOptions(dd.WithExportedOnly()),
	).P(srv)
}
