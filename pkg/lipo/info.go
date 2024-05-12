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
		v, isFat, err := info(bin)
		if err != nil {
			return nil, err
		}
		if isFat {
			fat = append(fat, v)
		} else {
			thin = append(thin, v)
		}
	}

	return append(fat, thin...), nil
}

func info(bin string) (string, bool, error) {
	fatFmt := "Architectures in the fat file: %s are: %s"

	arches, err := archs(bin)
	if err != nil {
		return "", false, err
	}

	v := strings.Join(arches, " ")

	_, typ, err := inspect(bin)
	if err != nil {
		return "", false, err
	}

	switch typ {
	case inspectThin:
		fallthrough
	case inspectArchive:
		return fmt.Sprintf("Non-fat file: %s is architecture: %s", bin, v), false, nil
	case inspectFat:
		return fmt.Sprintf(fatFmt, bin, v), true, nil
	default:
		return "", false, fmt.Errorf("unexpected type: %d", typ)
	}
}
