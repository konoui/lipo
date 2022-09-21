package lipo

import (
	"debug/macho"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

func (l *Lipo) Arches() error {
	if len(l.in) == 0 {
		return errors.New("no inputs")
	}
	if len(l.in) > 1 {
		return errors.New("only one input file allowed")
	}

	abs, err := filepath.Abs(l.in[0])
	if err != nil {
		return err
	}

	fat, err := macho.OpenFat(abs)
	if err != nil {
		if err != macho.ErrNotFat {
			return err
		}

		// if not fat file, assume single macho file
		f, err := macho.Open(abs)
		if err != nil {
			return err
		}
		defer f.Close()

		fmt.Println(cpu(f.Cpu.String()))
		return nil
	}
	defer fat.Close()

	cpus := []string{}
	for _, hdr := range fat.Arches {
		cpus = append(cpus, cpu(hdr.Cpu.String()))
	}

	fmt.Println(strings.Join(cpus, " "))
	return nil
}
