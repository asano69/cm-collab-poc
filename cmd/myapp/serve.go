package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/asano69/myapp/internal/collab"
	"github.com/asano69/myapp/internal/filesync"
	"github.com/asano69/myapp/internal/relayauth"
	"github.com/asano69/myapp/internal/static"
)

// dataPath is the file that persists the shared document.
const dataPath = "data.txt"

// metaPath is the file that persists the line IDs of the document.
const metaPath = "meta.json"

// serve runs the main server. Browsers connect to it directly, relays connect
// through the authenticated relay endpoint.
func serve(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	addr := fs.String("addr", "127.0.0.1:3002", "listen address")
	keys := fs.String("authorized-keys", "authorized_keys",
		"file listing the relay public keys that may connect")
	fs.Parse(args)

	hub := collab.NewHub()
	if err := filesync.Start(hub, dataPath); err != nil {
		return err
	}
	// After Start: the saved line IDs are matched against the loaded text.
	if err := filesync.StartMeta(hub, metaPath); err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.Handle("/connection/collab", hub)
	mux.Handle(relayauth.Path, relayauth.Require(*keys, hub))
	mux.Handle("/", http.FileServerFS(static.FS))

	log.Printf("listening on http://%s (relay keys: %s)", *addr, *keys)
	return http.ListenAndServe(*addr, mux)
}
