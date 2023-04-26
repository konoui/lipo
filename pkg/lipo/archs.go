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

	all := fat.AllArches()
	cpus := make([]string, len(all))
	for i, hdr := range all {
		cpus[i] = lmacho.ToCpuString(hdr.Cpu, hdr.SubCpu)
	}
	return cpus, nil
}
