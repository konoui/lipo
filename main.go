package main

import (
	"os"

	"github.com/konoui/lipo/cmd"
)

func main() {
	os.Exit(cmd.Execute(os.Stdout, os.Stderr, os.Args[1:]))
}
