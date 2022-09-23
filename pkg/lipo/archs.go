package lipo

import (
	"debug/macho"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

func (l *Lipo) Archs() error {
	arches, err := l.archs()
	if err != nil {
		return err
	}

	fmt.Fprintln(l.stdout, strings.Join(arches, " "))
	return nil
}

func (l *Lipo) archs() ([]string, error) {
	if len(l.in) == 0 {
		return nil, errors.New("no inputs")
	}
	if len(l.in) > 1 {
		return nil, errors.New("only one input file allowed")
	}

	abs, err := filepath.Abs(l.in[0])
	if err != nil {
		return nil, err
	}

	fat, err := macho.OpenFat(abs)
	if err != nil {
		if err != macho.ErrNotFat {
			return nil, err
		}

		// if not fat file, assume single macho file
		f, err := macho.Open(abs)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		return []string{CpuString(f.Cpu, f.SubCpu)}, nil
	}
	defer fat.Close()

	cpus := make([]string, 0, len(fat.Arches))
	for _, hdr := range fat.Arches {
		cpus = append(cpus, CpuString(hdr.Cpu, hdr.SubCpu))
	}
	return cpus, nil
}
