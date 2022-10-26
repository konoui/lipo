package lipo

import (
	"debug/macho"
	"errors"
	"fmt"

	"github.com/konoui/lipo/pkg/lipo/lmacho"
)

func (l *Lipo) Archs() ([]string, error) {
	if err := validateOneInput(l.in); err != nil {
		return nil, err
	}

	bin := l.in[0]
	return archs(bin)
}

func archs(bin string) ([]string, error) {
	fat, err := lmacho.OpenFat(bin)
	if err != nil {
		if !errors.Is(err, macho.ErrNotFat) {
			return nil, err
		}

		// if not fat file, assume single macho file
		f, err := macho.Open(bin)
		if err != nil {
			return nil, fmt.Errorf("not fat/thin file: %w", err)
		}
		defer f.Close()

		return []string{lmacho.ToCpuString(f.Cpu, f.SubCpu)}, nil
	}

	cpus := make([]string, 0, len(fat.Arches))
	for _, hdr := range fat.AllArches() {
		cpus = append(cpus, lmacho.ToCpuString(hdr.Cpu, hdr.SubCpu))
	}
	return cpus, nil
}
