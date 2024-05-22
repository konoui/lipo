package lipo

import (
	"fmt"
	"io"
	"strings"
)

func (l *Lipo) Info(stdout, stderr io.Writer) {
	if len(l.in) == 0 {
		fmt.Fprintln(stderr, "fatal error: "+errNoInput.Error())
		return
	}

	fat := make([]string, 0, len(l.in))
	thin := make([]string, 0, len(l.in))
	for _, bin := range l.in {
		v, typ, err := info(bin)
		if err != nil {
			fmt.Fprintln(stderr, err.Error())
			return
		}
		if typ == inspectFat {
			fat = append(fat, v)
		} else {
			thin = append(thin, v)
		}
	}

	out := strings.Join(append(fat, thin...), "\n")
	fmt.Fprintln(stdout, out)
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
