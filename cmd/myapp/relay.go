package main

import (
	"errors"
	"flag"
	"log"
	"net/http"

	"github.com/asano69/myapp/internal/relay"
	"github.com/asano69/myapp/internal/relayauth"
	"github.com/asano69/myapp/internal/static"
)

// relayCmd runs a relay: it serves the frontend and forwards every editor
// connection to the main server, signed with the relay's private key.
func relayCmd(args []string) error {
	fs := flag.NewFlagSet("relay", flag.ExitOnError)
	upstream := fs.String("upstream", "", "main server URL, e.g. wss://host:3002 (required)")
	addr := fs.String("addr", "127.0.0.1:3004", "listen address")
	keyPath := fs.String("key", "relay-private.pem", "private key file")
	fs.Parse(args)

	if *upstream == "" {
		return errors.New("--upstream is required")
	}
	key, err := relayauth.LoadPrivateKey(*keyPath)
	if err != nil {
		return err
	}
	proxy, err := relay.New(*upstream, key)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.Handle("/connection/collab", proxy)
	mux.Handle("/", http.FileServerFS(static.FS))

	log.Printf("relay public key: %s", relayauth.AuthorizedKey(key))
	log.Printf("relaying http://%s to %s", *addr, *upstream)
	return http.ListenAndServe(*addr, mux)
}
