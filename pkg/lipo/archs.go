package lipo

import (
	"debug/macho"
	"errors"

	"github.com/konoui/lipo/pkg/lipo/mcpu"
)

func (l *Lipo) Archs() ([]string, error) {
	if len(l.in) != 1 {
		return nil, errors.New("only one input file allowed")
	}

	bin := l.in[0]
	fat, err := macho.OpenFat(bin)
	if err != nil {
		if err != macho.ErrNotFat {
			return nil, err
		}

		// if not fat file, assume single macho file
		f, err := macho.Open(bin)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		return []string{mcpu.ToString(f.Cpu, f.SubCpu)}, nil
	}
	defer fat.Close()

	cpus := make([]string, 0, len(fat.Arches))
	for _, hdr := range fat.Arches {
		cpus = append(cpus, mcpu.ToString(hdr.Cpu, hdr.SubCpu))
	}
	return cpus, nil
}
