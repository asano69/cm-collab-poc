package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	// Without a subcommand the server starts, as it did before.
	cmd, args := "serve", os.Args[1:]
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		cmd, args = args[0], args[1:]
	}

	var err error
	switch cmd {
	case "serve":
		err = serve(args)
	case "relay":
		err = relayCmd(args)
	case "keygen":
		err = keygen(args)
	default:
		err = fmt.Errorf("unknown command %q (want serve, relay or keygen)", cmd)
	}
	if err != nil {
		log.Fatal(err)
	}
}
