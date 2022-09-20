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
	argIn := make([]string, 4)

	flaggy.SetName("lipo")
	flaggy.SetDescription("create an universal binary for macOS")
	flaggy.String(&out, "output", "output", "output file")
	flaggy.Bool(&create, "create", "create", "create flag")

	for idx := range argIn {
		required := true
		if idx > 1 {
			required = false
		}
		flaggy.AddPositionalValue(&argIn[idx], "input", idx+1, required, "input file")
	}

	flaggy.Parse()

	// validate
	in := make([]string, 0, len(argIn))
	for idx := range argIn {
		if argIn[idx] == "" {
			continue
		}
		in = append(in, argIn[idx])
	}

	if out == "" {
		fatal("-output flag is required")
	}
	if !create {
		fatal("-create flag is required")
	}

	l := lipo.New(lipo.WithOutput(out), lipo.WithInputs(in...))
	if err := l.Create(); err != nil {
		fatal(err.Error())
	}
}
