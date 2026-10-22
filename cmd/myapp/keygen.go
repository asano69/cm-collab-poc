package main

import (
	"flag"
	"fmt"

	"github.com/asano69/myapp/internal/relayauth"
)

// keygen creates the relay's key pair and prints the line that the main
// server needs in its authorized_keys file.
func keygen(args []string) error {
	fs := flag.NewFlagSet("keygen", flag.ExitOnError)
	out := fs.String("o", "relay-private.pem", "private key file to create")
	fs.Parse(args)

	key, err := relayauth.GenerateKey(*out)
	if err != nil {
		return err
	}
	fmt.Printf("wrote %s\n\nAdd this line to authorized_keys on the main server:\n\n%s\n",
		*out, relayauth.AuthorizedKey(key))
	return nil
}
