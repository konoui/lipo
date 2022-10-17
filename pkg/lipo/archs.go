package lipo

import (
	"debug/macho"
	"fmt"

	"github.com/konoui/lipo/pkg/lipo/mcpu"
)

func (l *Lipo) Archs() ([]string, error) {
	if err := l.validateOneInput(); err != nil {
		return nil, err
	}

	bin := l.in[0]
	return archs(bin)
}

func archs(bin string) ([]string, error) {
	fat, err := OpenFat(bin)
	if err != nil {
		if err != macho.ErrNotFat {
			return nil, err
		}

		// if not fat file, assume single macho file
		f, err := macho.Open(bin)
		if err != nil {
			return nil, fmt.Errorf("not fat/thin file: %w", err)
		}
		defer f.Close()

		return []string{mcpu.ToString(f.Cpu, f.SubCpu)}, nil
	}
	defer fat.Close()

	cpus := make([]string, 0, len(fat.Arches)+len(fat.HiddenArches))
	for _, hdr := range append(fat.Arches, fat.HiddenArches...) {
		cpus = append(cpus, mcpu.ToString(hdr.Cpu, hdr.SubCpu))
	}
	return cpus, nil
}
