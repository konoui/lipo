package main

import (
	"fmt"
	"os"

	"github.com/integrii/flaggy"
	"github.com/konoui/lipo/pkg/lipo"
)

var osExit = os.Exit

func fatal(msg string) {
	fmt.Printf("Error %s\n", msg)
	osExit(1)
}

func main() {
	var out string
	create := true
	in := make([]string, 2)

	flaggy.SetName("lipo")
	flaggy.SetDescription("create an universal binary for macOS")
	flaggy.String(&out, "output", "output", "output file")
	flaggy.Bool(&create, "create", "create", "create flag")

	for idx := range in {
		required := true
		if idx > 1 {
			required = false
		}
		flaggy.AddPositionalValue(&in[idx], "input", idx+1, required, "input file")
	}

	flaggy.Parse()
	if out == "" {
		fatal("-output flag is required")
	}
	if !create {
		fatal("-create flag is required")
	}

	l := lipo.New(out, in...)
	if err := l.Create(); err != nil {
		fatal(err.Error())
	}
}
