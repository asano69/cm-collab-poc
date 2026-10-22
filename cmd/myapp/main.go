package main

import (
	"log"
	"net/http"

	"github.com/asano69/myapp/internal/collab"
	"github.com/asano69/myapp/internal/static"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("/connection/collab", collab.NewHub())
	mux.Handle("/", http.FileServerFS(static.FS))

	addr := "127.0.0.1:3002"
	log.Printf("listening on http://%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
