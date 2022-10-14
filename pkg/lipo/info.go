package lipo

import (
	"debug/macho"
	"errors"
	"fmt"
	"strings"
)

func (l *Lipo) Info() ([]string, error) {
	if len(l.in) == 0 {
		return nil, errors.New("no input files specified")
	}

	fat := make([]string, 0, len(l.in))
	thin := make([]string, 0, len(l.in))
	for _, in := range l.in {
		arches, err := archs(in)
		if err != nil {
			return nil, err
		}
		v := strings.Join(arches, " ")
		fatFmt := "Architectures in the fat file: %s are: %s"
		if len(arches) > 1 {
			fat = append(fat, fmt.Sprintf(fatFmt, in, v))
			continue
		}

		f, err := macho.Open(in)
		if err == nil {
			f.Close()
			thin = append(thin, fmt.Sprintf("Non-fat file: %s is architecture: %s", in, v))
			continue
		}

		fat = append(thin, fmt.Sprintf(fatFmt, in, v))
	}

	return append(fat, thin...), nil
}
