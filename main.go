package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/ClarkGuan/transfer/client"
	"github.com/ClarkGuan/transfer/server"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "`client` or `server` must be added in arguments")
		return
	}

	subCommand := os.Args[1]
	switch strings.ToLower(subCommand) {
	case "client":
		if err := client.Main(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	case "server":
		if err := server.Main(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown argument: `%s`", subCommand)
	}
}
