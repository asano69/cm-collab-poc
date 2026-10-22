package main

import (
	"log"
	"net/http"

	"github.com/asano69/myapp/internal/collab"
	"github.com/asano69/myapp/internal/filesync"
	"github.com/asano69/myapp/internal/static"
)

// dataPath is the file that persists the shared document.
const dataPath = "data.txt"

func main() {
	hub := collab.NewHub()
	if err := filesync.Start(hub, dataPath); err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/connection/collab", hub)
	mux.Handle("/", http.FileServerFS(static.FS))

	addr := "127.0.0.1:3002"
	log.Printf("listening on http://%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
