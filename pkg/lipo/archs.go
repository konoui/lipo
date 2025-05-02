package lipo

import (
	"fmt"
)

func (l *Lipo) Archs() ([]string, error) {
	if err := validateOneInput(l.in); err != nil {
		return nil, err
	}

	bin := l.in[0]
	cpus, _, err := archs(bin)
	return cpus, err
}

func archs(bin string) ([]string, inspectType, error) {
	typ, err := inspect(bin)
	if err != nil {
		return nil, inspectUnknown, err
	}

	switch typ {
	case inspectThin:
		objs, err := OpenArches([]*ArchInput{{Bin: bin}})
		if err != nil {
			return nil, typ, err
		}
		defer close(objs...)
		return []string{objs[0].CPUString()}, inspectThin, nil
	case inspectArchive:
		archive, err := OpenArchive(bin)
		if err != nil {
			return nil, typ, err
		}
		defer archive.Close()
		return []string{archive.Arches[0].CPUString()}, typ, nil
	case inspectFat:
		fat, err := OpenFatFile(bin)
		if err != nil {
			return nil, typ, fmt.Errorf("internal error: %w", err)
		}
		defer fat.Close()

		cpus := make([]string, len(fat.Arches))
		for i := range cpus {
			cpus[i] = fat.Arches[i].CPUString()
		}
		return cpus, typ, nil
	default:
		return nil, inspectUnknown, fmt.Errorf("unexpected type: %d", typ)
	}
}
