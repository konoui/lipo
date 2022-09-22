package lipo

import (
	"debug/macho"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

func (l *Lipo) Archs() error {
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

		fmt.Fprintln(l.stdout, CpuString(f.Cpu, f.SubCpu))
		return nil
	}
	defer fat.Close()

	cpus := make([]string, 0, len(fat.Arches))
	for _, hdr := range fat.Arches {
		cpus = append(cpus, CpuString(hdr.Cpu, hdr.SubCpu))
	}

	fmt.Fprintln(l.stdout, strings.Join(cpus, " "))
	return nil
}
