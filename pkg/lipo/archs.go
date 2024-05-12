package lipo

import (
	"fmt"
)

func (l *Lipo) Archs() ([]string, error) {
	if err := validateOneInput(l.in); err != nil {
		return nil, err
	}

	bin := l.in[0]
	return archs(bin)
}

func archs(bin string) ([]string, error) {
	obj, typ, err := inspect(bin)
	if err != nil {
		return nil, err
	}

	switch typ {
	case inspectThin:
		fallthrough
	case inspectArchive:
		return []string{obj.CPUString()}, nil
	case inspectFat:
		fat, err := OpenFatFile(bin)
		if err != nil {
			return nil, fmt.Errorf("internal error: %w", err)
		}
		defer fat.Close()

		cpus := make([]string, len(fat.Arches))
		for i := range cpus {
			cpus[i] = fat.Arches[i].CPUString()
		}
		return cpus, nil
	default:
		return nil, fmt.Errorf("unexpected type: %d", typ)
	}
}
