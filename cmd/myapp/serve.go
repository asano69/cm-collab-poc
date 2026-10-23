package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"github.com/asano69/cmgo/filesync"
	"github.com/asano69/cmgo/persistence"
	"github.com/asano69/myapp/internal/authority"
	"github.com/asano69/myapp/internal/linehub"
	"github.com/asano69/myapp/internal/metasync"
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
	authorityPath := fs.String("authority", "Authority.pem",
		"private key that signs the lines (created if missing)")
	knownHosts := fs.String("known-hosts", "known_hosts",
		"file that the public key of this server is written to, for relays")
	fs.Parse(args)

	auth, err := authority.Load(*authorityPath)
	if err != nil {
		return err
	}
	if err := auth.WriteKnownHosts(*knownHosts, *addr); err != nil {
		return err
	}
	log.Printf("authority public key: %s (saved to %s)", auth.PublicKey(), *knownHosts)

	hub := linehub.New()
	hub.SetSigner(auth)
	// The document is stored as the file dataPath in the working directory.
	store, err := persistence.NewFilePersistence(".")
	if err != nil {
		return err
	}
	if err := filesync.Start(context.Background(), hub, store, dataPath); err != nil {
		return err
	}
	// After filesync.Start: the saved line IDs are matched against the loaded text.
	if err := metasync.Start(hub, metaPath); err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.Handle("/connection/collab", hub)
	mux.Handle(relayauth.Path, relayauth.Require(*keys, hub))
	mux.Handle(authority.KeyPath, authority.KeyHandler(auth.PublicKey()))
	mux.Handle("/", http.FileServerFS(static.FS))

	log.Printf("listening on http://%s (relay keys: %s)", *addr, *keys)
	return http.ListenAndServe(*addr, mux)
}
