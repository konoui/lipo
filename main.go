package main

import (
	"os"

	"github.com/konoui/lipo/cmd"
)

func main() {
	os.Exit(cmd.Execute(os.Stdout, os.Args[1:]))
}
