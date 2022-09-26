package main

import (
	"fmt"
	"os"
	"strings"

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
	var remove, extract, verifyArch string
	create := false
	archs := false

	argIn := make([]string, 4)

	flaggy.SetName("lipo")
	flaggy.SetDescription("create an universal binary for macOS")
	flaggy.String(&out, "output", "output", "output file")
	flaggy.Bool(&create, "create", "create", "create flag")
	flaggy.String(&remove, "remove", "remove", "remove <arch>")
	flaggy.String(&extract, "extract", "extract", "extract <arch>")
	flaggy.String(&verifyArch, "verify_arch", "verify_arch", "extract <arch>")
	flaggy.Bool(&archs, "archs", "archs", "show arch")

	for idx := range argIn {
		required := true
		if idx > 0 {
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

	if remove != "" {
		if len(in) == 0 {
			fatal("no inputs files")
		}
		if out == "" {
			fatal("-output flag is required")
		}
		l := lipo.New(lipo.WithOutput(out), lipo.WithInputs(in...))
		if err := l.Remove(remove); err != nil {
			fatal(err.Error())
		}
		return
	}

	if extract != "" {
		if len(in) == 0 {
			fatal("no inputs files")
		}
		if out == "" {
			fatal("-output flag is required")
		}
		l := lipo.New(lipo.WithOutput(out), lipo.WithInputs(in...))
		if err := l.Extract(extract); err != nil {
			fatal(err.Error())
		}
		return
	}

	if create {
		if len(in) == 0 {
			fatal("no inputs files")
		}
		if out == "" {
			fatal("-output flag is required")
		}
		l := lipo.New(lipo.WithOutput(out), lipo.WithInputs(in...))
		if err := l.Create(); err != nil {
			fatal(err.Error())
		}
		return
	}

	if archs {
		l := lipo.New(lipo.WithInputs(in...))
		arches, err := l.Archs()
		if err != nil {
			fatal(err.Error())
		}
		fmt.Fprintln(os.Stdout, strings.Join(arches, " "))
		return
	}

	if verifyArch != "" {
		l := lipo.New(lipo.WithInputs(in...))
		ok, err := l.VerifyArch(verifyArch)
		if err != nil {
			fatal(err.Error())
		}
		if !ok {
			os.Exit(1)
		}
		return
	}

}
