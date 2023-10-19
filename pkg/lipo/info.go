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
	ret, err := inspectFile(bin)
	if err != nil {
		return "", false, err
	}

	v := strings.Join(ret.arches, " ")
	if ret.fileType == typeFat {
		return fmt.Sprintf("Architectures in the fat file: %s are: %s", bin, v), true, nil
	}

	return fmt.Sprintf("Non-fat file: %s is architecture: %s", bin, v), false, nil
}
