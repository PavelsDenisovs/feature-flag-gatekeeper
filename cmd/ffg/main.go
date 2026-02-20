package main

import (
	"os"
	"github.com/PavelsDenisovs/feature-flag-gatekeeper/cmd/ffg/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}