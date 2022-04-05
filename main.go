package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/konoui/lipo/pkg/lipo"
)

var osExit = os.Exit

func fatal(msg string) {
	fmt.Printf("Error %s\n", msg)
	osExit(1)
}

func main() {
	var out string
	var create bool
	fs := flag.NewFlagSet("lipo", flag.ContinueOnError)
	fs.StringVar(&out, "output", "", "output file")
	fs.BoolVar(&create, "create", true, "create")
	if err := fs.Parse(os.Args[1:]); err != nil {
		fatal(err.Error())
	}

	if out == "" {
		fatal("-output flag is required")
	}
	if !create {
		fatal("-create flag is required")
	}

	in := fs.Args()
	l := lipo.New(out, in...)
	if err := l.Create(); err != nil {
		fatal(err.Error())
	}
}
