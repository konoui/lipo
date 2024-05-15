package lipo

import (
	"fmt"
	"strings"
)

func (l *Lipo) Info() ([]string, error) {
	if len(l.in) == 0 {
		return nil, errNoInput
	}

	fat := make([]string, 0, len(l.in))
	thin := make([]string, 0, len(l.in))
	for _, bin := range l.in {
		v, typ, err := info(bin)
		if err != nil {
			return nil, err
		}
		if typ == inspectFat {
			fat = append(fat, v)
		} else {
			thin = append(thin, v)
		}
	}

	return append(fat, thin...), nil
}

func info(bin string) (string, inspectType, error) {
	arches, typ, err := archs(bin)
	if err != nil {
		return "", typ, err
	}

	v := strings.Join(arches, " ")
	switch typ {
	case inspectThin:
		fallthrough
	case inspectArchive:
		return fmt.Sprintf("Non-fat file: %s is architecture: %s", bin, v), typ, nil
	case inspectFat:
		return fmt.Sprintf("Architectures in the fat file: %s are: %s", bin, v), typ, nil
	default:
		return "", inspectUnknown, fmt.Errorf("unexpected type: %d", typ)
	}
}
